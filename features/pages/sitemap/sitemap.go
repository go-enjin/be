//go:build page_sitemap || pages || all

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

package sitemap

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/net/html"

	"github.com/go-corelibs/x-text/language"
	"github.com/go-corelibs/x-text/message"

	"github.com/go-corelibs/slices"
	"github.com/go-corelibs/values"
	"github.com/go-enjin/be/pkg/context"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

var (
	DefaultChangeFreq = "never"
	ChangeFreqOptions = []string{"always", "hourly", "daily", "weekly", "monthly", "yearly", "never"}
)

var (
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-sitemap"

var (
	DefaultSiteScheme = "https"
)

type Feature interface {
	feature.Feature
	feature.ApplyMiddleware
	feature.PageContextFieldsProvider
	feature.PageContextParsersProvider
}

type MakeFeature interface {
	SetDomain(domain string) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	domain string
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
}

func (f *CFeature) SetDomain(domain string) MakeFeature {
	f.domain = domain
	return f
}

func (f *CFeature) Make() Feature {
	if f.domain != "" && !strings.HasPrefix(f.domain, "http://") && !strings.HasPrefix(f.domain, "https://") {
		log.FatalDF(1, "http:// or https:// required for sitemap domain setting")
	}
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

func (f *CFeature) PageContextParsers() (parsers context.Parsers) {
	parsers = context.Parsers{
		"sitemap-change-freq": f.ChangeFreqParser,
	}
	return
}

func (f *CFeature) ChangeFreqParser(spec *context.Field, input interface{}) (parsed interface{}, err error) {
	switch t := input.(type) {
	case string:
		t = strings.ToLower(t)
		if slices.Within(t, ChangeFreqOptions) {
			parsed = t
		} else {
			err = fmt.Errorf("not a change frequency")
		}
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}

func (f *CFeature) MakePageContextFields(r *http.Request) (fields context.Fields) {
	printer := message.GetPrinter(r)
	fields = context.Fields{
		"sitemap-ignored": {
			Key:          "sitemap-ignored",
			Tab:          "page",
			Label:        printer.Sprintf(`Enable to have this page omitted from the sitemap`),
			Category:     "sitemap",
			Input:        "checkbox",
			Format:       "bool",
			DefaultValue: "",
		},
		"sitemap-priority": {
			Key:          "sitemap-priority",
			Tab:          "page",
			Label:        printer.Sprintf(`Specify the priority for this page in the sitemap`),
			Category:     "sitemap",
			Input:        "range",
			Format:       "decimal-percent",
			DefaultValue: 0.5,
			Minimum:      0.0,
			Maximum:      1.0,
		},
		"sitemap-change-freq": {
			Key:          "sitemap-change-freq",
			Tab:          "page",
			Label:        printer.Sprintf("Specify the change frequency for this page in the sitemap"),
			Category:     "sitemap",
			Input:        "select",
			Format:       "sitemap-change-freq",
			DefaultValue: DefaultChangeFreq,
			ValueOptions: ChangeFreqOptions,
		},
	}
	return
}

func (f *CFeature) Apply(s feature.System) (err error) {
	s.Router().Get("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		langMode := f.Enjin.SiteLanguageMode()
		defaultTag := f.Enjin.SiteDefaultLanguage()

		var domain string
		if domain = f.domain; domain == "" {
			domain = DefaultSiteScheme + "://" + r.Host
		}

		spec, _ := f.Enjin.MakePageContextField("sitemap-change-freq", r)

		pages := make(map[string]feature.Page)
		for _, found := range f.Enjin.FindPages("/") {
			if ignored, ok := found.Context().Boolean("SitemapIgnored"); !ok || (ok && !ignored) {
				priority := found.Context().Float64("SitemapPriority", 0.5)
				found.Context().SetSpecific("SitemapPriority", priority)

				if changeFreq := found.Context().String("SitemapChangeFreq", ""); changeFreq != "" {
					if safe, ee := f.ChangeFreqParser(spec, changeFreq); ee == nil {
						found.Context().SetSpecific("SitemapChangeFreq", safe)
					} else {
						log.ErrorRF(r, "error: page has invalid sitemap-change-freq: %v", changeFreq)
						found.Context().Delete("SitemapChangeFreq")
					}
				}

				tag := found.LanguageTag()
				if language.Compare(tag, language.Und) {
					tag = defaultTag
				}

				fullUrl := langMode.ToUrl(defaultTag, tag, found.Url())
				if !strings.HasPrefix(fullUrl, "http") && domain != "" {
					fullUrl = domain + fullUrl
				}

				pages[fullUrl] = found
			}
		}

		var contents string
		contents += `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
		contents += `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n"

		for _, fullUrl := range maps.SortedKeys(pages) {
			pg := pages[fullUrl]
			contents += "\t<url>\n"
			contents += "\t\t<loc>" + html.EscapeString(fullUrl) + "</loc>\n"
			contents += "\t\t<lastmod>" + pg.UpdatedAt().Format("2006-01-02") + "</lastmod>\n"
			if priority := pg.Context().Float64("SitemapPriority", -1.0); priority >= 0.0 {
				contents += fmt.Sprintf("\t\t<priority>%0.1f</priority>\n", priority)
			}
			if changeFreq := pg.Context().String("SitemapChangeFreq", ""); changeFreq != "" {
				contents += fmt.Sprintf("\t\t<changefreq>%s</changefreq>\n", changeFreq)
			}
			contents += "\t</url>\n"
		}

		contents += `</urlset>`

		f.Enjin.ServeData([]byte(contents), "application/xml", w, r)
	})
	return
}
