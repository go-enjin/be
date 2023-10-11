//go:build !exclude_theme_renderer

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

package renderer

import (
	"bytes"
	"errors"
	"fmt"
	htmlTemplate "html/template"
	"strings"

	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	beErrors "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request/argv"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "srv-theme-renderer"

type Feature interface {
	feature.Feature
	feature.ThemeRenderer
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
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *CFeature) Shutdown() {

}

func (f *CFeature) Render(t feature.Theme, view string, ctx beContext.Context) (data []byte, err error) {

	var tt *htmlTemplate.Template
	if tt, err = f.NewHtmlTemplateFromContext(t, view, ctx); err == nil {
		log.TraceF("template used: (%v) %v - \"%v\"", view, tt.Name(), ctx.String("Url", "nil"))
		var wr bytes.Buffer
		if err = tt.Execute(&wr, ctx); err != nil {
			return
		}
		data = wr.Bytes()
	}

	return
}

func (f *CFeature) PrepareRenderPage(t feature.Theme, ctx beContext.Context, p feature.Page) (data htmlTemplate.HTML, redirect string, err error) {

	ctx.Apply(p.Context().Copy())
	ctx.Set("Theme", t.GetConfig())

	var output string
	pageFormat := p.Format()

	if pageFormat == "html.tmpl" {
		if output, err = f.RenderHtmlTemplateContent(t, ctx, p.Content()); err != nil {
			err = beErrors.ParseTemplateError(err.Error(), p.Content())
			return
		}
	} else if strings.HasSuffix(pageFormat, ".tmpl") {
		if output, err = f.RenderTextTemplateContent(t, ctx, p.Content()); err != nil {
			err = beErrors.ParseTemplateError(err.Error(), p.Content())
			return
		}
	} else {
		output = p.Content()
	}

	if err == nil {
		if format := t.GetFormat(pageFormat); format != nil {
			if data, redirect, err = format.Process(ctx, output); err == nil {
				log.TraceF("page format success: %v", format.Name())
			}
		} else {
			err = fmt.Errorf("unsupported page format")
		}
	}

	return
}

func (f *CFeature) RenderPage(t feature.Theme, ctx beContext.Context, p feature.Page) (data []byte, redirect string, err error) {

	if html, redir, ee := f.PrepareRenderPage(t, ctx, p); ee != nil {
		var enjErr *beErrors.EnjinError
		if errors.As(ee, &enjErr) {
			ctx["Content"] = enjErr.Html()
		} else {
			ctx["Content"] = "<p>" + ee.Error() + "</p>"
		}
		delete(ctx, "Archetype")
		ctx["Layout"] = "full-view"
		ctx["Format"] = "html"
		data, err = f.Render(t, "single", ctx)
		return
	} else if redir != "" {
		redirect = redir
		return
	} else {
		ctx["Content"] = html
	}

	if !p.Context().Bool(argv.RequestArgvIgnoredKey, false) {
		if redirect = ctx.String(argv.RequestRedirectKey, ""); redirect != "" {
			return
		} else if consumed := ctx.Bool(argv.RequestArgvConsumedKey, false); !consumed {
			if reqArgv, ok := ctx.Get(string(argv.RequestArgvKey)).(*argv.RequestArgv); ok && reqArgv != nil && reqArgv.MustConsume() {
				redirect = p.Url()
				return
			}
		}
	}

	data, err = f.Render(t, "single", ctx)

	return
}