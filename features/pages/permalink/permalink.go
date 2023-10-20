//go:build page_permalink || pages || all

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

package permalink

import (
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/pages"
	bePath "github.com/go-enjin/be/pkg/path"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-permalink"

type Feature interface {
	feature.Feature
	feature.UseMiddleware
	feature.PageContextModifier
	feature.FuncMapProvider
	feature.PageContextFieldsProvider
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
}

func (f *CFeature) MakeFuncMap(ctx context.Context) (fm feature.FuncMap) {
	fm = feature.FuncMap{
		"_permalink": f._permalink,
		"newPermalink": func() (permalink string) {
			if id, err := uuid.NewV4(); err == nil {
				permalink = id.String()
			}
			return
		},
	}
	return
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	feature.RegisterPageMatcherFuncs(f._permalinkMatcher)
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) MakePageContextFields(r *http.Request) (fields context.Fields) {
	printer := lang.GetPrinterFromRequest(r)
	id, _ := uuid.NewV4()
	fields = context.Fields{
		"permalink": {
			Key:          "permalink",
			Tab:          "page",
			Label:        printer.Sprintf("Set this page's permalink"),
			Category:     "file",
			Weight:       10,
			Input:        "text",
			Format:       "uuid",
			DefaultValue: id.String(),
			LockNonEmpty: true,
		},
	}
	return
}

func (f *CFeature) FilterPageContext(themeCtx, pageCtx context.Context, r *http.Request) (out context.Context) {
	out = themeCtx
	out.SetSpecific("SitePermalinkable", true)
	return
}

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := forms.TrimQueryParams(r.URL.Path)
			if _, p, ok := lang.ParseLangPath(path); ok {
				path = p
			}

			if permalink, ok := f._parsePath(path); ok {
				permalinkPath := bePath.CleanWithSlash(permalink)

				log.DebugF("permalink detected: %v", permalinkPath)

				reqTag := lang.GetTag(r)
				defTag := f.Enjin.SiteDefaultLanguage()
				var checkTags []language.Tag
				if reqTag != defTag {
					checkTags = append(checkTags, reqTag)
				} else if defTag != language.Und {
					checkTags = append(checkTags, defTag)
				}
				checkTags = append(checkTags, language.Und)

				for _, checkTag := range checkTags {
					if p := f.Enjin.FindPage(checkTag, permalinkPath); p != nil {

						var destination string
						if p.Url() == "" || p.Url() == "." || p.Url() == "/" {
							destination = "/"
						} else {
							destination = p.Url() + "-"
						}
						destination += p.PermalinkSha()

						if path != destination {
							http.Redirect(w, r, destination, http.StatusSeeOther)
							return
						} else if err := f.Enjin.ServePage(p, w, r); err == nil {
							return
						} else {
							log.ErrorRF(r, "error serving permalink page: [%v] %v - %v", p.Language(), path, err)
						}

					} else {
						log.ErrorF("permalinked page not found [%v]", checkTag)
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (f *CFeature) _parsePath(path string) (permalink string, ok bool) {
	if permalink, ok = pages.ParsePermalink(path); ok {
		if size := len(permalink); size == 48 {
			log.TraceDF(1, "found permalink root: %v - %v", path, permalink)
		} else if size == 10 {
			log.TraceDF(1, "found permalink slug: %v - %v", path, permalink)
		}
	}
	return
}

func (f *CFeature) _permalink(input interface{}) (url string, err error) {
	var permalink uuid.UUID
	if vs, ok := input.(string); ok {
		if permalink, err = uuid.FromString(vs); err != nil {
			return
		}
	} else if vu, ok := input.(uuid.UUID); ok {
		permalink = vu
	} else {
		err = fmt.Errorf("expected uuid.UUID or string; received %T", input)
		return
	}
	if permalink != uuid.Nil {
		url = "/" + permalink.String()
		for _, tag := range f.Enjin.SiteLocales() {
			if f.Enjin.SiteSupportsLanguage(tag) {
				if p := f.Enjin.FindPage(tag, url); p != nil {
					url = p.Url() + "-" + p.PermalinkSha()
					return
				}
			}
		}
	}
	return
}

func (f *CFeature) _permalinkMatcher(path string, p feature.Page) (found string, ok bool) {
	found = path
	if p.Permalink() != uuid.Nil {
		if parsed, valid := f._parsePath(path); valid {
			switch len(parsed) {
			case 10:
				ok = parsed == p.PermalinkSha()
			case 36:
				// e0f7ae8b-85e0-4c3f-b6c7-4c84b59bd3e7
				if parsedUuid := uuid.FromStringOrNil(parsed); parsedUuid != uuid.Nil {
					ok = parsedUuid.String() == parsedUuid.String()
				}
			}
		}
	}
	return
}