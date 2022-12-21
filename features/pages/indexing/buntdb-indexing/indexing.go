//go:build buntdb_indexing || buntdb || all

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
	"github.com/tidwall/buntdb"
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

const Tag feature.Tag = "BuntdbIndexing"

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

	stubs  map[string]*pagecache.Stub
	dbpath string
	kvs    *kvs
	// buntdb *buntdb.DB

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
	f.dbpath = ":memory:"
	f.stubs = make(map[string]*pagecache.Stub)
	// f.memCache = make(map[string]map[string]int)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	b.AddFlags(&cli.PathFlag{
		Name:    "buntdb-path",
		Usage:   fmt.Sprintf("specify the path to use for the %v database", Tag),
		EnvVars: b.MakeEnvKeys("BUNTDB_PATH"),
	})
	b.AddCommands(&cli.Command{
		Name:   "buntdb-precache",
		Usage:  "precache content indexing data (requires --buntdb-path)",
		Action: f.buntdbKwsCommandAction,
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
	if dbpath := ctx.Path("buntdb-path"); dbpath != "" {
		f.dbpath = dbpath
	}
	var cfg *buntdb.Config
	if f.cliStartup {
		cfg = &buntdb.Config{AutoShrinkDisabled: true, SyncPolicy: buntdb.Always}
	}
	if f.kvs, err = newKvs(f.dbpath, cfg); err != nil {
		err = fmt.Errorf("error opening buntdb kvs: %v - %v", f.dbpath, err)
	} else {
		log.DebugF("using buntdb: %v", f.dbpath)
	}
	return
}

func (f *CFeature) Shutdown() {
	if f.cliStartup {
		f.kvs.ShrinkLogStatsAndClose()
	} else {
		f.kvs.Close()
	}
}

func (f *CFeature) GetPageContextValueStubs(key string) (valueStubs map[interface{}]pagecache.Stubs, err error) {
	valueStubs = make(map[interface{}]pagecache.Stubs)
	valueStubsPattern := fmt.Sprintf("pql:%v:*:value", key)
	if err = f.kvs.DB(valueStubsPattern).View(func(tx *buntdb.Tx) (err error) {
		if err = tx.AscendKeys(valueStubsPattern, func(k, v string) (ok bool) {
			var idx int
			var contextKey string
			if contextKey, idx, _, err = parsePqlValueStubsKey(k); err != nil {
				return
			}
			if contextKey != key {
				ok = true
				return
			}

			var vi kvm.Value
			if err = vi.UnmarshalBinary([]byte(v)); err != nil {
				log.Error(err)
				return
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

			var csvStubs string
			if csvStubs, err = tx.Get(fmt.Sprintf("pql:%v:%d:stubs", key, idx)); err != nil {
				log.Error(err)
				return
			}
			f.parseCsvInts(csvStubs)

			var shasums []string
			if shasums, err = f.getShasumsFromIndexes(f.parseCsvInts(csvStubs)); err != nil {
				log.Error(err)
				return
			}

			for _, val := range values {
				for _, shasum := range shasums {
					if stub, ok := f.stubs[shasum]; ok {
						valueStubs[val] = append(valueStubs[val], stub)
					}
				}
			}

			ok = true
			return
		}); err != nil {
			err = fmt.Errorf("error ascending pattern: %v - %v", err)
			return
		}
		return
	}); err != nil {
		err = fmt.Errorf("error getting context key values: %v", err)
		return
	}
	return
}