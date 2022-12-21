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
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/tidwall/buntdb"

	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
)

type kvs struct {
	parentPath string
	stubPath   string
	kwsPath    string
	pqlPath    map[string]string

	stub *buntdb.DB
	kws  *buntdb.DB
	pql  map[string]*buntdb.DB
	cfg  *buntdb.Config
}

func newKvs(path string, config *buntdb.Config) (s *kvs, err error) {
	s = &kvs{
		parentPath: path,
		pqlPath:    make(map[string]string),
		pql:        make(map[string]*buntdb.DB),
		cfg:        config,
	}
	if !bePath.IsDir(s.parentPath) {
		err = bePath.Mkdir(s.parentPath)
	}

	s.stubPath = fmt.Sprintf("%v/stubs.db", s.parentPath)
	if s.stub, err = buntdb.Open(s.stubPath); err != nil {
		return
	}
	if s.cfg != nil {
		if err = s.stub.SetConfig(*s.cfg); err != nil {
			return
		}
	}

	s.kwsPath = fmt.Sprintf("%v/kws.db", s.parentPath)
	if s.kws, err = buntdb.Open(s.kwsPath); err != nil {
		return
	}
	if s.cfg != nil {
		if err = s.kws.SetConfig(*s.cfg); err != nil {
			return
		}
	}
	return
}

func (s *kvs) DB(key string) (db *buntdb.DB) {
	switch {
	case strings.HasPrefix(key, "contents:"):
		db = s.kws
	case strings.HasPrefix(key, "stub:"):
		db = s.stub
	case strings.HasPrefix(key, "pql:"):
		var ok bool
		contextKey, _, _, _ := parsePqlValueStubsKey(key)
		if db, ok = s.pql[contextKey]; !ok {
			var err error
			s.pqlPath[contextKey] = fmt.Sprintf("%v/pql--%v.db", s.parentPath, contextKey)
			if db, err = buntdb.Open(s.pqlPath[contextKey]); err != nil {
				log.FatalDF(1, "error opening pql context key buntdb: %v - %v", s.pqlPath[contextKey], err)
			} else {
				s.pql[contextKey] = db
				if s.cfg != nil {
					if err = s.pql[contextKey].SetConfig(*s.cfg); err != nil {
						return
					}
				}
			}
		}
	default:
		log.FatalDF(1, "db for key missing: %v", key)
	}
	return
}

func (s *kvs) Shrink() (err error) {
	err = s.stub.Shrink()
	err = s.kws.Shrink()
	for _, db := range s.pql {
		err = db.Shrink()
	}
	return
}

func (s *kvs) shrinkAndLogStats(path string, db *buntdb.DB) {
	var before int64
	if info, err := os.Stat(path); err != nil {
		log.ErrorF("error getting file info for: %v - %v", path, err)
	} else {
		before = info.Size()
	}
	if err := db.Shrink(); err != nil {
		log.ErrorF("error shrinking buntdb: %v", err)
	}
	if info, err := os.Stat(path); err != nil {
		log.ErrorF("error getting file info for: %v - %v", path, err)
	} else {
		current := info.Size()
		switch {
		case before > current:
			log.DebugF("%v shrunk %d bytes (current: %v bytes)", path, before-current, current)
		case before < current:
			log.DebugF("%v grew %d bytes (current: %v bytes)", path, current-before, current)
		default:
			log.DebugF("%v did not change (current: %v bytes)", path, current)
		}
	}
}

func (s *kvs) ShrinkLogStatsAndClose() {

	s.shrinkAndLogStats(s.stubPath, s.stub)
	if err := s.stub.Close(); err != nil {
		log.ErrorF("error closing stub buntdb: %v", err)
	}

	s.shrinkAndLogStats(s.kwsPath, s.kws)
	if err := s.kws.Close(); err != nil {
		log.ErrorF("error closing kws buntdb: %v", err)
	}

	for k, db := range s.pql {
		s.shrinkAndLogStats(s.pqlPath[k], db)
		if err := db.Close(); err != nil {
			log.ErrorF("error closing pql (%v) buntdb: %v", k, err)
		}
	}
	return
}

func (s *kvs) Close() {
	if err := s.stub.Close(); err != nil {
		log.ErrorF("error closing stub buntdb: %v", err)
	}

	if err := s.kws.Close(); err != nil {
		log.ErrorF("error closing kws buntdb: %v", err)
	}

	for k, db := range s.pql {
		if err := db.Close(); err != nil {
			log.ErrorF("error closing pql (%v) buntdb: %v", k, err)
		}
	}
	return
}

var rxPqlNextKey = regexp.MustCompile(`^pql:next:([^:]+?)$`)
var rxPqlDataKey = regexp.MustCompile(`^pql:data:([^:]+?):([^:]+?)$`)
var rxPqlValueStubsKey = regexp.MustCompile(`^pql:([^:]+?):(\*|\d+?):(value|stubs)$`)

func parsePqlValueStubsKey(input string) (contextKey string, idx int, vs string, err error) {
	if rxPqlValueStubsKey.MatchString(input) {
		m := rxPqlValueStubsKey.FindAllStringSubmatch(input, 1)
		parts := m[0][1:]
		if parts[1] == "*" {
			idx = -1
		} else if idx, err = strconv.Atoi(parts[1]); err != nil {
			err = fmt.Errorf("error parsing pql value/stubs key index: %v", parts[2])
			return
		}
		contextKey = parts[0]
		vs = parts[2]
	} else if rxPqlDataKey.MatchString(input) {
		m := rxPqlDataKey.FindAllStringSubmatch(input, 1)
		contextKey = m[0][1]
		idx = -1
		vs = "data"
	} else if rxPqlNextKey.MatchString(input) {
		m := rxPqlNextKey.FindAllStringSubmatch(input, 1)
		contextKey = m[0][1]
		idx = -1
		vs = "next"
	} else {
		err = fmt.Errorf("not pql value/stubs key: %v", input)
	}
	return
}