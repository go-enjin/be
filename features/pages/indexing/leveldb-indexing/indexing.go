//go:build leveldb_indexing || all

// Copyright (c) 2022  The Go-Enjin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package indexing

import (
	"fmt"
	"sync"

	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/maruel/natural"
	"github.com/syndtr/goleveldb/leveldb/opt"

	// "github.com/robaho/leveldb"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps/kvm"
	"github.com/go-enjin/be/pkg/pagecache"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "LeveldbIndexing"

type Feature interface {
	feature.Feature
	pagecache.QueryEnjinFeature
	pagecache.SearchEnjinFeature
	pagecache.PageContextProvider
}

type CFeature struct {
	feature.CFeature

	cli   *cli.Context
	enjin feature.Internals

	docMaps map[language.Tag]map[string]*mapping.DocumentMapping

	stubs   map[string]*pagecache.Stub
	dbpath  string
	leveldb *leveldb.DB

	// useMemCache bool
	// memCache    map[string]map[string]int
	// nextPqlIdx  map[string]int

	cliStartup bool
	cliBatch   int

	sync.RWMutex
}

type MakeFeature interface {
	SetPath(dbpath string) MakeFeature

	Make() Feature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	return f
}

func (f *CFeature) SetPath(dbpath string) MakeFeature {
	f.dbpath = dbpath
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.docMaps = make(map[language.Tag]map[string]*mapping.DocumentMapping)
	f.stubs = make(map[string]*pagecache.Stub)
	// f.memCache = make(map[string]map[string]int)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	b.AddFlags(&cli.PathFlag{
		Name:    "leveldb-path",
		Usage:   fmt.Sprintf("specify the path to use for the %v database", Tag),
		EnvVars: b.MakeEnvKeys("LEVELDB_PATH"),
	})
	b.AddCommands(&cli.Command{
		Name:   "leveldb-precache",
		Usage:  "precache content indexing data (requires --leveldb-path)",
		Action: f.leveldbKwsCommandAction,
	})
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.enjin = enjin
	locales := f.enjin.SiteLocales()
	for _, feat := range f.enjin.Features() {
		if v, ok := feat.Self().(pagecache.SearchDocumentMapperFeature); ok {
			for _, tag := range locales {
				if _, exists := f.docMaps[tag]; !exists {
					f.docMaps[tag] = make(map[string]*mapping.DocumentMapping)
				}
				doctype, dm := v.SearchDocumentMapping(tag)
				f.docMaps[tag][doctype] = dm
			}
		}
	}
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	f.cli = ctx
	if dbpath := ctx.Path("leveldb-path"); dbpath != "" {
		f.dbpath = dbpath
	}
	var db *leveldb.DB
	options := &opt.Options{
		// ReadOnly:       true,
		ErrorIfMissing: true,
	}
	if f.cliStartup {
		options.ReadOnly = false
		options.ErrorIfMissing = false
	}
	if db, err = leveldb.OpenFile(f.dbpath, options); err != nil {
		err = fmt.Errorf("error opening leveldb: %v - %v", f.dbpath, err)
	} else {
		f.leveldb = db
		log.DebugF("using leveldb: %v", f.dbpath)
	}
	return
}

func (f *CFeature) Shutdown() {
	if f.leveldb != nil {
		var stats leveldb.DBStats
		if err := f.leveldb.Stats(&stats); err != nil {
			log.Error(err)
		} else {
			log.DebugF("leveldb stats:\n%#v", stats)
		}
		if err := f.leveldb.Close(); err != nil {
			log.ErrorF("error shutting down indexing leveldb: %v", err)
		} else {
			log.DebugF("leveldb closed")
		}
	}
}

func (f *CFeature) leveldbKeyCompare(a, b []byte) (r int) {
	sa, sb := string(a), string(b)
	switch {
	case sa == sb:
		r = 0
	case natural.Less(sa, sb):
		r = -1
	default:
		r = 1
	}
	return
}

func (f *CFeature) GetPageContextValueStubs(key string) (valueStubs map[interface{}]pagecache.Stubs, err error) {
	valueStubs = make(map[interface{}]pagecache.Stubs)

	lower := []byte(fmt.Sprintf(gPqlContextKeyValueFormat, key, 0))

	iter := f.leveldb.NewIterator(nil, nil)
	for ok := iter.Seek(lower); ok; ok = iter.Next() {

		dataKey := iter.Key()
		dataValue := iter.Value()

		var e error
		var idx int
		var contextKey, vs string
		if contextKey, idx, vs, e = f.parsePqlValueStubsKey(string(dataKey)); e != nil {
			// log.ErrorF("end of pql iteration: %v", e)
			break
		}
		if contextKey != key || vs != "value" {
			continue // safety check?
		}

		var vi kvm.Value
		if err = (&vi).UnmarshalBinary(dataValue); err != nil {
			log.ErrorF("unmarshal error: \"%v\" - %v", string(dataValue), err)
			break
		}

		var values []interface{}
		switch tv := vi.Get().(type) {
		case []string:
			for _, vtv := range tv {
				values = append(values, vtv)
			}
		case []interface{}:
			for _, vtv := range tv {
				values = append(values, vtv)
			}
		default:
			values = append(values, tv)
		}

		var csvData []byte
		var csvStubs string
		if csvData, err = f.leveldb.Get([]byte(fmt.Sprintf(gPqlContextKeyStubsFormat, key, idx)), nil); err != nil {
			log.Error(err)
			break
		}
		csvStubs = string(csvData)

		var shasums []string
		if shasums, err = f.getShasumsFromIndexes(f.parseCsvInts(csvStubs)); err != nil {
			log.Error(err)
			break
		}

		for _, val := range values {
			for _, shasum := range shasums {
				if stub, ok := f.stubs[shasum]; ok {
					valueStubs[val] = append(valueStubs[val], stub)
				}
			}
		}

	}
	iter.Release()
	if err != nil {
		return
	}
	if err = iter.Error(); err != nil {
		log.Error(err)
		return
	}

	return
}