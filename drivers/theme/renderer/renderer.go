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
	"fmt"
	htmlTemplate "html/template"
	"strings"

	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request/argv"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "driver-theme-renderer"

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

func (f *CFeature) Render(view string, ctx beContext.Context) (data []byte, err error) {

	var tt *htmlTemplate.Template
	if tt, err = f.NewHtmlTemplateFromContext(view, ctx); err == nil {
		log.TraceF("template used: (%v) %v - \"%v\"", view, tt.Name(), ctx.String("Url", "nil"))
		var wr bytes.Buffer
		if err = tt.Execute(&wr, ctx); err != nil {
			return
		}
		data = wr.Bytes()
	}

	return
}

func (f *CFeature) RenderPage(ctx beContext.Context, p feature.Page) (data []byte, redirect string, err error) {

	t := f.Enjin.MustGetTheme()

	ctx.Apply(p.Context().Copy())
	ctx.Set("Theme", t.GetConfig())

	var output string

	if p.Format() == "html.tmpl" {
		if output, err = f.RenderHtmlTemplateContent(ctx, p.Content()); err != nil {
			ctx["Content"] = f.renderErrorPage("Template Render Error", err.Error(), p.String())
		}
	} else if strings.HasSuffix(p.Format(), ".tmpl") {
		// TODO: find a more safe way to pre-render .njn.tmpl files
		if output, err = f.RenderTextTemplateContent(ctx, p.Content()); err != nil {
			ctx["Content"] = f.renderErrorPage("Template Render Error", err.Error(), p.String())
		}
	} else {
		output = p.Content()
	}

	if err == nil {
		if format := t.GetFormat(p.Format()); format != nil {
			if html, redir, ee := format.Process(ctx, output); ee != nil {
				if enjerr, ok := ee.(*errors.EnjinError); ok {
					log.ErrorF("error processing %v page format: %v - %v", p.Format(), enjerr.Title, enjerr.Summary)
					ctx["Content"] = enjerr.Html()
				} else {
					log.ErrorF("error processing %v page format: %v", p.Format(), ee.Error())
					ctx["Content"] = "<p>" + ee.Error() + "</p>"
				}
			} else if redir != "" {
				redirect = redir
				return
			} else {
				ctx["Content"] = html
				log.TraceF("page format success: %v", format.Name())
			}
		} else {
			ctx["Content"] = f.renderErrorPage("Unsupported Page Format", fmt.Sprintf(`Unknown page format specified: "%v"`, p.Format()), p.String())
		}
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

	data, err = f.Render("single", ctx)

	return
}