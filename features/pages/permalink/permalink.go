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
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/context"
	"github.com/go-corelibs/enjinql"
	clPath "github.com/go-corelibs/path"
	"github.com/go-corelibs/x-text/language"
	"github.com/go-corelibs/x-text/message"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/pages/page_fields"
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
	feature.QueryIndexSourceFeature
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
}

func (f *CFeature) MakeFuncMap(ctx context.Context) (fm feature.FuncMap) {
	r, _ := ctx.Get("R").(*http.Request)
	fm = feature.FuncMap{
		"_permalink": func(v interface{}) (url string, err error) {
			url, err = f._permalink(r, v)
			return
		},
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
	f.CFeature.Construct(f)
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

func (f *CFeature) ListPageContextFields() (kebabs []string) {
	return []string{"permalink"}
}

func (f *CFeature) MakePageContextFields(r *http.Request) (list page_fields.Fields) {
	printer := message.GetPrinter(r)
	id, _ := uuid.NewV4()
	list = page_fields.Fields{
		"permalink": {
			Key:          "permalink",
			Tab:          "page",
			Label:        printer.Sprintf("Set this page's permalink"),
			Category:     "file",
			Weight:       54,
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

			if permalink, short, ok := f._parsePath(path); ok {
				permalinkPath := clPath.CleanWithSlash(permalink)

				log.DebugF("permalink detected: %v", permalinkPath)

				reqTag := message.GetTag(r)
				defTag := f.Enjin.SiteDefaultLanguage()
				var checkTags []language.Tag
				if reqTag != defTag {
					checkTags = append(checkTags, reqTag)
				} else if defTag != language.Und {
					checkTags = append(checkTags, defTag)
				}
				checkTags = append(checkTags, language.Und)

				for _, checkTag := range checkTags {

					var eql string
					if short {
						eql = `LOOKUP .Url, %[1]s.Short, .Shasum WITHIN (.Language == "%[2]s") AND (%[1]s.Short == %[3]q);`
					} else {
						eql = `LOOKUP .Url, %[1]s.Short, .Shasum WITHIN (.Language == "%[2]s") AND (%[1]s.Long  == %[3]q);`
					}

					if _, results, ee := f.Enjin.PerformLookup(eql, enjinql.PagePermalinkSource, checkTag.String(), permalink); ee == nil {

						if len(results) == 0 {
							log.ErrorRF(r, "query success, no results found: %v", eql)

						} else {

							var destination string
							shasum := results[0].String("Shasum", "")
							pUrl := results[0].String("Url", "")
							if pUrl == "" || pUrl == "." || pUrl == "/" {
								destination = "/"
							} else {
								destination = pUrl + "-"
							}
							destination += results[0].String("Short", "")

							if path != destination {
								http.Redirect(w, r, destination, http.StatusSeeOther)
								return
							} else if pages, eee := f.Enjin.PerformQueryPages(r, `QUERY WITHIN .Shasum == %q;`, shasum); eee != nil {
								log.ErrorRF(r, "error querying page by shasum %q: %q - %v", pUrl, eee)
								return
							} else if len(pages) == 1 {
								if err := f.Enjin.ServePage(pages[0], w, r); err != nil {
									log.ErrorRF(r, "error serving permalink page: [%v] %v - %v", pages[0].Language(), path, err)
								} else {
									return
								}
							}
						}

					} else {
						log.ErrorRF(r, "error performing permalink lookup [%v]: %q - %v", checkTag, permalink, ee)
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
