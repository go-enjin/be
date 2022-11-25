//go:build page_search || pages || all

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

package pql

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/pagecache"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var (
	_ Feature                     = (*CFeature)(nil)
	_ MakeFeature                 = (*CFeature)(nil)
	_ pagecache.QueryEnjinFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "PagesQueryPQL"

type Feature interface {
	feature.Feature
}

type CFeature struct {
	feature.CFeature

	cli   *cli.Context
	enjin feature.Internals

	index map[string]map[interface{}][]*pagecache.Stub

	sync.RWMutex
}

type MakeFeature interface {
	Make() Feature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.index = make(map[string]map[interface{}][]*pagecache.Stub)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.enjin = enjin
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	f.cli = ctx
	return
}

var RxMatchQLOrderBy = regexp.MustCompile(`ORDER BY \.(.+?)\s*(ASC|DSC|DESC)?\s*$`)

func (f *CFeature) PerformQuery(input string) (stubs []*pagecache.Stub, err error) {
	// orderBy := "Title"
	// sortDir := "ASC"
	// if RxMatchQLOrderBy.MatchString(input) {
	// 	m := RxMatchQLOrderBy.FindAllStringSubmatch(input, 1)
	// 	orderBy = strcase.ToCamel(m[0][1])
	// 	switch strings.ToUpper(m[0][2]) {
	// 	case "ASC", "":
	// 	case "DSC", "DESC":
	// 		sortDir = "DSC"
	// 	default:
	// 		log.ErrorF("invalid sort direction specified in PageQL statement: %v", m[0][2])
	// 	}
	// 	input = RxMatchQLOrderBy.ReplaceAllString(input, "")
	// }
	input = RxMatchQLOrderBy.ReplaceAllString(input, "")

	stubs, err = f.processQuery(input)
	return
}

func (f *CFeature) AddToQueryIndex(stub *pagecache.Stub, p *page.Page) (err error) {
	for k, v := range p.Context {
		if beStrings.StringInStrings(strings.ToLower(k), "content", "frontmatter") {
			continue
		}
		if _, check := f.index[k]; !check {
			f.index[k] = make(map[interface{}][]*pagecache.Stub)
		}
		switch t := v.(type) {
		case string,
			float32, float64,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			time.Time, time.Duration:
			f.index[k][t] = append(f.index[k][t], stub)
		}
	}
	return
}

func (f *CFeature) RemoveFromQueryIndex(tag language.Tag, file, shasum string) {
	// panic("implement me")
	return
}