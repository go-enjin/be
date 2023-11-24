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

package be

import (
	"net/http"
	"time"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/go-enjin/golang-org-x-text/message"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/headers"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
	"github.com/go-enjin/be/pkg/net/headers/policy/permissions"
)

func (e *Enjin) Features() (cache *feature.FeaturesCache) {
	cache = e.eb.features
	return
}

func (e *Enjin) Pages() (pages map[string]feature.Page) {
	pages = e.eb.pages
	return
}

func (e *Enjin) Theme() (theme string) {
	theme = e.eb.theme
	return
}

func (e *Enjin) Theming() (theming map[string]feature.Theme) {
	theming = e.eb.theming
	return
}

func (e *Enjin) Headers() (headers []headers.ModifyHeadersFn) {
	headers = e.eb.headers
	return
}

func (e *Enjin) Domains() (domains []string) {
	domains = e.eb.domains
	return
}

func (e *Enjin) Consoles() (consoles map[feature.Tag]feature.Console) {
	consoles = e.eb.consoles
	return
}

func (e *Enjin) Processors() (processors map[string]feature.ReqProcessFn) {
	processors = e.eb.processors
	return
}

func (e *Enjin) Translators() (translators map[string]feature.TranslateOutputFn) {
	translators = e.eb.translators
	return
}

func (e *Enjin) Transformers() (transformers map[string]feature.TransformOutputFn) {
	transformers = e.eb.transformers
	return
}

func (e *Enjin) Slugsums() (enabled bool) {
	enabled = e.eb.slugsums
	return
}

func (e *Enjin) PermissionsPolicy() (handler *permissions.PolicyHandler) {
	handler = e.permissionsPolicy
	return
}

func (e *Enjin) ContentSecurityPolicy() (handler *csp.PolicyHandler) {
	handler = e.contentSecurityPolicy
	return
}

func (e *Enjin) PublicFileSystems() (registry fs.Registry) {
	registry = e.eb.publicFileSystems
	return
}

func (e *Enjin) ListTemplatePartials(block, position string) (names []string) {
	found := make(map[string]struct{})
	for _, tpp := range e.eb.fTemplatePartialsProvider {
		for _, name := range tpp.ListTemplatePartials(block, position) {
			if _, present := found[name]; present {
				continue
			}
			names = append(names, name)
			found[name] = struct{}{}
		}
	}
	return
}

func (e *Enjin) GetTemplatePartial(block, position, name string) (tmpl string, ok bool) {
	for _, tpp := range e.eb.fTemplatePartialsProvider {
		if tmpl, ok = tpp.GetTemplatePartial(block, position, name); ok {
			return
		}
	}
	return
}

func (e *Enjin) GetThemeRenderer(ctx context.Context) (renderer feature.ThemeRenderer) {

	if namedRenderer := ctx.String("ThemeRenderer", ""); namedRenderer != "" {
		for _, tr := range e.eb.fThemeRenderers {
			if tr.Tag().Equal(feature.Tag(namedRenderer)) {
				renderer = tr
				break
			}
		}
	}
	if renderer == nil {
		renderer = e.eb.fThemeRenderers[0]
	}

	return
}

func (e *Enjin) MakeLanguagePrinter(requested string) (tag language.Tag, printer *message.Printer) {
	tag, printer = lang.NewCatalogPrinter(requested, e.SiteLanguageCatalog())
	return
}

func (e *Enjin) PublicUserActions() (actions feature.Actions) {
	actions = e.eb.publicUser
	return
}

func (e *Enjin) MakePageContextField(key string, r *http.Request) (field *context.Field, ok bool) {
	fields := context.Fields{}

	for _, fp := range e.GetPageContextFieldsProviders() {
		for k, v := range fp.MakePageContextFields(r) {
			if _, present := fields[k]; !present {
				// no clobbering!
				fields[k] = v
			}
		}
	}

	if field, ok = fields[key]; ok {
		printer := lang.GetPrinterFromRequest(r)
		parsers := e.PageContextParsers()
		var fn context.Parser
		if fn, ok = parsers[field.Format]; ok && fn != nil {
			field.Parse = fn
			field.Printer = printer
			if field.Tab == "" {
				field.Tab = "page"
			}
			if field.Category == "" {
				field.Category = "general"
			}
		} else {
			log.ErrorRF(r, "%q field format not found: %q", key, field.Format)
			field = nil
			ok = false
		}
	}

	return
}

func (e *Enjin) MakePageContextFields(r *http.Request) (fields context.Fields) {
	fields = context.Fields{}

	for _, fp := range e.GetPageContextFieldsProviders() {
		for k, v := range fp.MakePageContextFields(r) {
			if _, present := fields[k]; !present {
				// no clobbering!
				fields[k] = v
			}
		}
	}

	printer := lang.GetPrinterFromRequest(r)
	parsers := e.PageContextParsers()
	for k, v := range fields {
		if fn := parsers[v.Format]; fn != nil {
			v.Parse = fn
			v.Printer = printer
			if v.Tab == "" {
				v.Tab = "page"
			}
			if v.Category == "" {
				v.Category = "general"
			}
		} else {
			log.ErrorRF(r, "%q field format not found: %q", k, v.Format)
			delete(fields, k)
		}
	}

	return
}

func (e *Enjin) PageContextParsers() (parsers context.Parsers) {
	parsers = context.Parsers{} // return a copy, not the source
	for k, fn := range context.DefaultParsers {
		parsers[k] = fn
	}
	for _, fp := range e.GetPageContextParsersProviders() {
		for k, v := range fp.PageContextParsers() {
			if _, present := parsers[k]; !present {
				// no clobbering!
				parsers[k] = v
			}
		}
	}
	return
}

func (e *Enjin) CreateNonce(key string) (value string) {
	if e.eb.fNonceFactory != nil {
		value = e.eb.fNonceFactory.CreateNonce(key)
	} else {
		value = e.nonces.CreateNonce(key)
	}
	return
}

func (e *Enjin) VerifyNonce(key, value string) (valid bool) {
	if e.eb.fNonceFactory != nil {
		valid = e.eb.fNonceFactory.VerifyNonce(key, value)
	} else {
		valid = e.nonces.VerifyNonce(key, value)
	}
	return
}

func (e *Enjin) CreateToken(key string) (value, shasum string) {
	if e.eb.fTokenFactory != nil {
		value, shasum = e.eb.fTokenFactory.CreateToken(key)
	} else {
		value, shasum = e.tokens.CreateToken(key)
	}
	return
}

func (e *Enjin) CreateTokenWith(key string, duration time.Duration) (value, shasum string) {
	if e.eb.fTokenFactory != nil {
		value, shasum = e.eb.fTokenFactory.CreateTokenWith(key, duration)
	} else {
		value, shasum = e.tokens.CreateTokenWith(key, duration)
	}
	return
}

func (e *Enjin) VerifyToken(key, value string) (valid bool) {
	if e.eb.fTokenFactory != nil {
		valid = e.eb.fTokenFactory.VerifyToken(key, value)
	} else {
		valid = e.tokens.VerifyToken(key, value)
	}
	return
}

func (e *Enjin) NewSyncLocker(tag feature.Tag, key string, store feature.KeyValueStore) (l feature.SyncLocker) {
	l = e.eb.fSyncLockerFactory.NewSyncLocker(tag, key, store)
	return
}

func (e *Enjin) NewSyncLockerWith(tag feature.Tag, key string, store feature.KeyValueStore, timeout, interval time.Duration) (l feature.SyncLocker) {
	l = e.eb.fSyncLockerFactory.NewSyncLockerWith(tag, key, store, timeout, interval)
	return
}

func (e *Enjin) NewSyncRWLocker(tag feature.Tag, key string, readStore, writeStore feature.KeyValueStore) (l feature.SyncRWLocker) {
	l = e.eb.fSyncLockerFactory.NewSyncRWLocker(tag, key, readStore, writeStore)
	return
}

func (e *Enjin) NewSyncRWLockerWith(tag feature.Tag, key string, readStore, writeStore feature.KeyValueStore, timeout, interval time.Duration) (l feature.SyncRWLocker) {
	l = e.eb.fSyncLockerFactory.NewSyncRWLockerWith(tag, key, readStore, writeStore, timeout, interval)
	return
}
