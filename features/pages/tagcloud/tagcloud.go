// Copyright (c) 2024  The Go-Enjin Authors
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

/*
Package tagcloud provides tag cloud enjin features

From Wikipedia: (see: https://en.wikipedia.org/wiki/Tag_cloud)

	A tag cloud is a visual representation of text data which is often used
	to depict keyword metadata on websites, or to visualize free form text.
	Tags are usually single words, and the importance of each tag is shown
	with font size or color. When used as website navigation aids, the terms
	are hyperlinked to items associated with the tag.

- indexing support to cache and maintain a mapping of tag weights
- njn block to render a variable-sized widget
- shortcode to wrap the njn block
*/
package tagcloud

import (
	"net/http"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/maruel/natural"
	"github.com/urfave/cli/v2"

	clPath "github.com/go-corelibs/path"
	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request/argv"
	"github.com/go-enjin/be/pkg/signals"
	"github.com/go-enjin/be/types/page"
)

var (
	DefaultPageUrl  = "/tags"
	DefaultIndexKey = "tags"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "tagcloud"

type Feature interface {
	feature.Feature
}

type MakeFeature interface {
	SetPageUrl(path string) MakeFeature
	SetIndexKey(name string) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	pageUrl  string
	indexKey string

	ignoreTypes []string
	archetypes  []string

	cache *tagCache
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.CFeature.Construct(f)
	f.pageUrl = DefaultPageUrl
	f.indexKey = DefaultIndexKey
	f.cache = newTagCache()
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	return
}

func (f *CFeature) SetPageUrl(path string) MakeFeature {
	f.pageUrl = clPath.CleanWithSlash(path)
	return f
}

func (f *CFeature) SetIndexKey(name string) MakeFeature {
	f.indexKey = strcase.ToKebab(name)
	return f
}

func (f *CFeature) IgnorePageTypes(types ...string) MakeFeature {
	f.ignoreTypes = types
	return f
}

func (f *CFeature) OnlyArchetypes(types ...string) MakeFeature {
	f.archetypes = types
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)
	enjin.Connect(signals.ContentAddIndexing, f.Tag().String(), f.addPageIndexingFn)
	enjin.Connect(signals.ContentRemoveIndexing, f.Tag().String(), f.removePageIndexingFn)
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) GetTagCloud() (tags feature.TagCloud) {
	tags = f.cache.TagCloud()
	tags.Update(f.pageUrl)
	return
}

func (f *CFeature) GetPageTagCloud(shasum string) (tags feature.TagCloud) {
	tags = f.cache.Find(shasum)
	return
}

func (f *CFeature) MakePageContextFields(r *http.Request) (fields context.Fields) {
	printer := message.GetPrinter(r)
	fields = context.Fields{
		"no-tag-indexing": {
			Key:      "no-tag-indexing",
			Tab:      "page",
			Category: "tag-cloud",
			Label:    printer.Sprintf("Omit this page from tag indexing"),
			Input:    "checkbox",
			Format:   "bool",
		},
		f.indexKey: {
			Key:      f.indexKey,
			Tab:      "page",
			Category: "tag-cloud",
			Label:    printer.Sprintf("Space separated list of tag words"),
			Format:   "string",
		},
	}
	return
}

func (f *CFeature) PageTypeNames() (names []string) {
	names = []string{"tagcloud"}
	return
}

func (f *CFeature) ProcessRequestPageType(r *http.Request, p feature.Page) (pg feature.Page, redirect string, processed bool, err error) {
	if p.Type() != "tagcloud" {
		return
	}

	// prepare arguments
	reqArgv := argv.Get(r)
	reqArgv.NumPerPage = -1
	reqArgv.PageNumber = -1
	p.Context().SetSpecific(argv.RequestConsumedKey, true)

	var word string
	if len(reqArgv.Argv) > 0 {
		if len(reqArgv.Argv[0]) > 0 {
			word = reqArgv.Argv[0][0]
		}
	}

	t := f.Enjin.MustGetTheme()
	if word = forms.StrictSanitize(word); word != "" {
		var ok bool
		var tw *tagWord
		if tw, ok = f.cache.Get(word); !ok {
			redirect = reqArgv.Path
			log.WarnF("unknown tag requested, redirecting to: %v", redirect)
			return
		}

		var pages []*feature.CloudTagPage
		for _, shasum := range tw.Shasums() {
			if stub := f.Enjin.FindPageStub(shasum); stub != nil {
				if found, err := page.NewPageFromStub(stub, t); err == nil {
					pages = append(pages, &feature.CloudTagPage{
						Title:   found.Title(),
						Url:     found.Url(),
						Tags:    f.GetPageTagCloud(shasum),
						Created: found.CreatedAt(),
						Updated: found.UpdatedAt(),
					})
				}
			}
		}
		sortDirV := p.Context().String("SortDir", "dsc")
		sortMode := p.Context().String("SortTags", "created")
		switch v := strings.ToLower(sortMode); v {
		case "updated", "natural":
			sortMode = v
		default:
			sortMode = "created"
		}
		sortAsc := strings.ToLower(sortDirV) == "asc"
		sort.Slice(pages, func(i, j int) (less bool) {
			a, b := pages[i], pages[j]
			switch sortMode {
			case "natural":
				if sortAsc {
					less = natural.Less(a.Title, b.Title)
				} else {
					less = natural.Less(b.Title, a.Title)
				}
			case "created":
				if sortAsc {
					less = a.Created.Unix() < b.Created.Unix()
				} else {
					less = a.Created.Unix() > b.Created.Unix()
				}
			case "updated":
				if sortAsc {
					less = a.Updated.Unix() < b.Updated.Unix()
				} else {
					less = a.Updated.Unix() > b.Updated.Unix()
				}
			}
			return
		})
		p.Context().SetSpecific("TagCloudWord", word)
		p.Context().SetSpecific("TagCloudPages", pages)
		processed = true
	}

	if !processed {
		// full-page tag cloud
		tc := f.cache.TagCloud()
		tc.Update(p.Url())
		tc.Sort()
		p.Context().SetSpecific("TagCloud", tc)
	}

	pg = p
	processed = true
	return
}
