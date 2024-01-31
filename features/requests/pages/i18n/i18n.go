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

package i18n

import (
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/x-text/language"
	"github.com/go-corelibs/x-text/message"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request/argv"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "requests-pages-i18n"

type Feature interface {
	feature.Feature
	feature.PrepareServePagesFeature
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

func (f *CFeature) PrepareServePage(ctx context.Context, t feature.Theme, p feature.Page, w http.ResponseWriter, r *http.Request) (out context.Context, modified *http.Request, stop bool) {
	reqLangTag := message.GetTag(r)
	ctx.SetSpecific("ReqLangTag", reqLangTag)

	ctx.SetSpecific(lang.PrinterKey, message.GetPrinter(r))
	ctx.SetSpecific(string(argv.RequestKey), argv.Get(r))

	parsedTag := f.Enjin.SiteDefaultLanguage()
	if v := ctx.Get("Language"); v != nil {
		if pageLang, ok := v.(string); ok {
			if pageLang != "" {
				if tag, ee := language.Parse(pageLang); ee == nil {
					parsedTag = tag
				} else {
					log.ErrorRF(r, "invalid language tag: %v - %v", pageLang, ee)
				}
			}
		} else {
			log.ErrorRF(r, "page language tag not a string: %T %+v", v, v)
		}
	}
	ctx.SetSpecific("Language", parsedTag.String())
	ctx.SetSpecific("LanguageTag", parsedTag)

	out = ctx
	return
}
