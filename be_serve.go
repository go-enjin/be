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

package be

import (
	"net/http"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/page"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/theme"
)

func (e *Enjin) Serve403(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte("403 - Forbidden"))
}

func (e *Enjin) Serve404(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("404 - Not Found"))
}

func (e *Enjin) ServePage(p *page.Page, w http.ResponseWriter, r *http.Request) (err error) {
	var data []byte
	var t *theme.Theme
	if t, err = e.GetTheme(); err != nil {
		return
	}
	ctx := e.Context()
	ctx.Set("BaseUrl", net.BaseURL(r.URL))
	ctx.Apply(p.Context.Copy())
	for _, f := range e.be.features {
		if s, ok := f.(feature.PageContextModifier); ok {
			log.DebugF("filtering page context with: %v", f.Tag())
			ctx = s.FilterPageContext(ctx, p.Context, r)
		}
	}
	if data, err = t.RenderPage(ctx, p); err != nil {
		return
	}
	e.ServeData(data, "text/html; charset=utf-8", w, r)
	return
}

func (e *Enjin) ServeData(data []byte, mime string, w http.ResponseWriter, _ *http.Request) {
	// only one translation allowed, non-feature translators take precedence
	basicMime := beStrings.GetBasicMime(mime)
	if fn, ok := e.be.translators[basicMime]; ok {
		data, mime = fn(data)
		basicMime = beStrings.GetBasicMime(mime)
	} else {
		for _, f := range e.be.features {
			if v, ok := f.(feature.OutputTranslator); ok {
				// log.DebugF("checking output filter: %v", f.Tag())
				if v.CanTranslate(mime) {
					if d, m, err := v.TranslateOutput(e, data, mime); err == nil {
						data, mime = d, m
						basicMime = beStrings.GetBasicMime(mime)
						log.DebugF("filtered output: %v - %v", f.Tag(), mime)
					} else {
						log.DebugF("%v error filtering output: %v", f.Tag(), err)
					}
					break
				}
			}
		}
	}

	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Content-Type", mime)

	for _, f := range e.be.features {
		if tfr, ok := f.(feature.OutputTransformer); ok {
			if tfr.CanTransform(mime) {
				data = tfr.TransformOutput(mime, data)
			}
		}
	}

	// non-feature transformers happen after features
	if fn, ok := e.be.transformers[mime]; ok {
		data = fn(data)
	} else if fn, ok := e.be.transformers[basicMime]; ok {
		data = fn(data)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}