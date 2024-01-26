// Copyright (c) 2023  The Go-Enjin Authors
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

package locales

import (
	"context"
	"net/http"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request/argv"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "srv-locales"

type Feature interface {
	feature.Feature
	feature.LocaleHandler
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
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
	return
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

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) LocaleHandler(next http.Handler) (this http.Handler) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := forms.CleanRequestPath(r.URL.Path)
		langMode := f.Enjin.SiteLanguageMode()
		defaultTag := f.Enjin.SiteDefaultLanguage()

		var reqPath string
		var requested language.Tag

		var reqOk bool
		if requested, reqPath, reqOk = langMode.FromRequest(defaultTag, r); !reqOk {
			log.WarnRF(r, "language mode rejecting request: %#v", r)
			f.Enjin.Serve404(w, r) // specifically not ServeNotFound()
			return
		} else if !f.Enjin.SiteSupportsLanguage(requested) {
			log.DebugRF(r, "%v language not supported, using default: %v", requested, defaultTag)
			requested = defaultTag
			reqPath = urlPath
		}

		// TODO: determine what to do with Accept-Language request headers
		// if acceptLanguage := r.Header.Get("Accept-Language"); acceptLanguage != "" {
		// 	if parsed, err := language.Parse(acceptLanguage); err == nil {
		// 		if e.SiteSupportsLanguage(parsed) && !language.Compare(requested, parsed) {
		// 			// requested = parsed
		// 			// e.ServeRedirect(langMode.ToUrl(requested, parsed, reqPath), w, r)
		// 			// return
		// 		}
		// 	}
		// }

		if v, ok := r.Context().Value(lang.LanguageTag).(language.Tag); ok {
			// a request modifier feature is specifying the user's language
			requested = v
		}

		tag, printer := f.Enjin.MakeLanguagePrinter(requested.String())
		ctx := context.WithValue(r.Context(), lang.LanguageTag, tag)
		ctx = context.WithValue(ctx, lang.LanguagePrinter, printer)
		ctx = context.WithValue(ctx, lang.LanguageDefault, f.Enjin.SiteDefaultLanguage())

		if reqPath == "" {
			reqPath = "/"
		} else if reqPath[0] != '/' {
			reqPath = "/" + reqPath
		}

		if reqArgv, ok := ctx.Value(argv.RequestKey).(*argv.Argv); ok {
			reqArgv.Path = reqPath
			reqArgv.Language = tag
			ctx = context.WithValue(ctx, argv.RequestKey, reqArgv)
		}

		r.URL.Path = reqPath
		r.RequestURI = strings.Replace(r.RequestURI, urlPath, reqPath, 1)
		if r.RequestURI == "" {
			r.RequestURI = "/"
		}

		next.ServeHTTP(w, r.Clone(ctx))
	})
}
