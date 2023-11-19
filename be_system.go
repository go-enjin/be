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
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/maruel/natural"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/userbase"
	"github.com/go-enjin/be/types/page"
)

func (e *Enjin) Router() (router *chi.Mux) {
	router = e.router
	return
}

func (e *Enjin) ServerName() (name string) {
	name = globals.BinName
	if e.debug {
		if globals.Version != "" && !e.production {
			name += " " + globals.Version
		}
	}
	return
}

func (e *Enjin) GetThemeName() (name string) {
	name = e.eb.context.String("Theme", e.eb.theme)
	return
}

func (e *Enjin) GetTheme() (t feature.Theme, err error) {
	var name string
	if name = e.GetThemeName(); name == "" {
		err = fmt.Errorf(`theme not found: "%v" %v`, name, e.ThemeNames())
		return
	}
	t, err = e.GetThemeNamed(name)
	return
}

func (e *Enjin) MustGetTheme() (t feature.Theme) {
	var err error
	if t, err = e.GetTheme(); err != nil {
		log.FatalDF(1, "error getting enjin theme: %v", err)
	}
	return
}

func (e *Enjin) ThemeNames() (names []string) {
	for name := range e.eb.theming {
		names = append(names, name)
	}
	sort.Sort(natural.StringSlice(names))
	return
}

func (e *Enjin) GetThemeNamed(name string) (t feature.Theme, err error) {
	var ok bool
	if t, ok = e.eb.theming[name]; !ok {
		err = fmt.Errorf(`theme not found: "%v" %v`, name, e.ThemeNames())
		return
	}
	return
}

func (e *Enjin) MustGetThemeNamed(name string) (t feature.Theme) {
	var err error
	if t, err = e.GetThemeNamed(name); err != nil {
		log.FatalDF(1, "error getting enjin theme: %v", err)
	}
	return
}

func (e *Enjin) MakeFuncMap(ctx context.Context) (fm feature.FuncMap) {
	fm = feature.FuncMap{}
	for _, fmp := range e.eb.fFuncMapProviders {
		made := fmp.MakeFuncMap(ctx)
		for name, fn := range made {
			fm[name] = fn
		}
	}
	return
}

func (e *Enjin) Prefix() (prefix string) {
	return e.prefix
}

func (e *Enjin) Context() (ctx context.Context) {
	ctx = e.eb.context.Copy()
	for _, ecp := range e.eb.fEnjinContextProvider {
		ctx.ApplySpecific(ecp.EnjinContext())
	}
	if e.debug {
		ctx.SetSpecific("Debug", true)
	}
	ctx.SetSpecific("Server", e.ServerName())
	ctx.SetSpecific("Prefix", e.prefix)
	if e.production {
		ctx.SetSpecific("PrefixLabel", "")
	} else {
		ctx.SetSpecific("PrefixLabel", "["+strings.ToUpper(e.prefix)+"] ")
	}
	tName := ctx.String("Theme", e.eb.theme)
	if _, ok := e.eb.theming[tName]; ok {
		ctx.SetSpecific("Theme", tName)
	} else {
		if tNames := e.ThemeNames(); len(tNames) > 0 {
			ctx.SetSpecific("Theme", tNames[0])
		} else {
			ctx.SetSpecific("Theme", "")
		}
	}
	now := time.Now()
	ctx.SetSpecific("Year", now.Year())
	ctx.SetSpecific("CurrentYear", now.Year())
	ctx.SetSpecific("Release", globals.BinHash)
	ctx.SetSpecific("Version", globals.Version)
	ctx.SetSpecific("EnjinInfo", feature.MakeEnjinInfo(e))
	ctx.SetSpecific("EnjinBase", feature.EnjinBase(e))
	return
}

func (e *Enjin) FindRedirection(url string) (p feature.Page) {
	for _, provider := range e.eb.fPageProviders {
		if p = provider.FindRedirection(url); p != nil {
			return
		}
	}
	for _, pg := range e.eb.pages {
		if pg.IsRedirection(url) {
			p = pg
			return
		}
	}
	return
}

func (e *Enjin) FindTranslations(url string) (pages feature.Pages) {
	for _, provider := range e.eb.fPageProviders {
		if found := provider.FindTranslations(url); len(found) > 0 {
			pages = append(pages, found...)
		}
	}
	for _, pg := range e.eb.pages {
		if _, ok := pg.Match(url); ok {
			pages = append(pages, pg)
		}
	}
	return
}

func (e *Enjin) FindTranslationUrls(url string) (pages map[language.Tag]string) {
	pages = make(map[language.Tag]string)
	for _, provider := range e.eb.fPageProviders {
		for tag, path := range provider.FindTranslationUrls(url) {
			pages[tag] = path
		}
	}
	for _, pg := range e.eb.pages {
		if _, ok := pg.Match(url); ok {
			pages[pg.LanguageTag()] = pg.Url()
		}
	}
	return
}

func (e *Enjin) FindFile(path string) (data []byte, mime string, err error) {
	for _, provider := range e.eb.fFileProviders {
		if d, m, ee := provider.FindFile(path); ee == nil {
			data = d
			mime = m
			return
		}
	}
	err = os.ErrNotExist
	return
}

func (e *Enjin) FindPage(tag language.Tag, url string) (p feature.Page) {
	for _, provider := range e.eb.fPageProviders {
		if p = provider.FindPage(tag, url); p != nil {
			return
		}
	}
	for _, pg := range e.eb.pages {
		if _, ok := pg.Match(url); ok {
			if language.Compare(pg.LanguageTag(), tag) || pg.IsTranslation(url) {
				p = pg
				break
			}
		}
	}
	return
}

func (e *Enjin) FindPages(prefix string) (pages []feature.Page) {
	for _, provider := range e.eb.fPageProviders {
		pages = append(pages, provider.LookupPrefixed(prefix)...)
	}
	return
}

func (e *Enjin) ListFormats() (names []string) {
	for _, p := range e.eb.fFormatProviders {
		names = append(names, p.ListFormats()...)
	}
	return
}

func (e *Enjin) GetFormat(name string) (format feature.PageFormat) {
	for _, p := range e.eb.fFormatProviders {
		if format = p.GetFormat(name); format != nil {
			return
		}
	}
	return
}

func (e *Enjin) MatchFormat(filename string) (format feature.PageFormat, match string) {
	for _, p := range e.eb.fFormatProviders {
		if format, match = p.MatchFormat(filename); format != nil {
			return
		}
	}
	return
}

func (e *Enjin) CheckMatchQL(query string) (pages []feature.Page, err error) {
	t, _ := e.GetTheme()
	for _, queryEnjin := range e.eb.fQueryIndexFeatures {
		if matches, ee := queryEnjin.PerformQuery(query); ee != nil {
			err = ee
		} else {
			for _, stub := range matches {
				if p, err := page.NewPageFromStub(stub, t); err != nil {
					log.ErrorF("error making page from cache: %v", err)
				} else {
					pages = append(pages, p)
				}
			}
		}
		// first query index feature wins?
		break
	}
	return
}

func (e *Enjin) MatchQL(query string) (pages []feature.Page) {
	t, _ := e.GetTheme()
	for _, queryEnjin := range e.eb.fQueryIndexFeatures {
		if matches, err := queryEnjin.PerformQuery(query); err != nil {
			log.ErrorF("error performing enjin query: %v", err)
		} else {
			for _, stub := range matches {
				if p, ee := page.NewPageFromStub(stub, t); ee != nil {
					log.ErrorF("error making page from cache: %v", ee)
				} else {
					pages = append(pages, p)
				}
			}
		}
		// first query index feature wins?
		break
	}
	return
}

func (e *Enjin) MatchStubsQL(query string) (stubs []*feature.PageStub) {
	for _, queryEnjin := range e.eb.fQueryIndexFeatures {
		var err error
		if stubs, err = queryEnjin.PerformQuery(query); err != nil {
			log.ErrorF("error performing enjin query: %v", err)
		}
		// first query index feature wins?
		break
	}
	return
}

func (e *Enjin) CheckMatchStubsQL(query string) (stubs []*feature.PageStub, err error) {
	for _, queryEnjin := range e.eb.fQueryIndexFeatures {
		stubs, err = queryEnjin.PerformQuery(query)
		// first query index feature wins?
		break
	}
	return
}

func (e *Enjin) SelectQL(query string) (selected map[string]interface{}) {
	for _, queryEnjin := range e.eb.fQueryIndexFeatures {
		var err error
		if selected, err = queryEnjin.PerformSelect(query); err != nil {
			log.ErrorF("error performing enjin select: %v", err)
		}
		// first query index feature wins?
		break
	}
	return
}

func (e *Enjin) CheckSelectQL(query string) (selected map[string]interface{}, err error) {
	for _, queryEnjin := range e.eb.fQueryIndexFeatures {
		selected, err = queryEnjin.PerformSelect(query)
		// first query index feature wins?
		break
	}
	return
}

func (e *Enjin) FindPageStub(shasum string) (stub *feature.PageStub) {
	for _, pcp := range e.eb.fPageContextProviders {
		if stub = pcp.FindPageStub(shasum); stub != nil {
			return
		}
	}
	return
}

func (e *Enjin) TranslateShortcodes(content string, ctx context.Context) (modified string) {
	if len(e.eb.fPageShortcodeProcessors) > 0 {
		// first shortcode processor wins?
		modified = content
		for _, psp := range e.eb.fPageShortcodeProcessors {
			modified = psp.TranslateShortcodes(modified, ctx)
		}
	} else {
		// passthrough nop
		modified = content
	}
	return
}

func (e *Enjin) ValidateUserRequest(action feature.Action, w http.ResponseWriter, r *http.Request) (valid bool) {
	if action.IsNil() {
		log.ErrorRDF(r, 1, "developer error - action requested is nil")
		return
	}

	eid := userbase.GetCurrentEID(r)
	if valid = userbase.CurrentUserCan(r, action); valid {
		log.DebugRDF(r, 1, "user %q is allowed to: %v", eid, action)
		return
	}

	log.WarnRDF(r, 1, "user %q is not allowed to: %v", eid, action)
	e.ServeNotFound(w, r)
	return
}