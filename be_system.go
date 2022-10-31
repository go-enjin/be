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
	"sort"
	"strings"
	"time"

	"github.com/fvbommel/sortorder"
	"github.com/go-chi/chi/v5"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/theme"
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

func (e *Enjin) GetTheme() (t *theme.Theme, err error) {
	tName := e.eb.context.String("Theme", e.eb.theme)
	var ok bool
	if t, ok = e.eb.theming[tName]; !ok {
		err = fmt.Errorf(`theme not found: "%v" %v`, tName, e.ThemeNames())
		return
	}
	return
}

func (e *Enjin) ThemeNames() (names []string) {
	for name := range e.eb.theming {
		names = append(names, name)
	}
	sort.Sort(sortorder.Natural(names))
	return
}

func (e *Enjin) Prefix() (prefix string) {
	return e.prefix
}

func (e *Enjin) Context() (ctx context.Context) {
	ctx = e.eb.context.Copy()
	if e.debug {
		ctx.Set("Debug", true)
	}
	ctx.Set("Server", e.ServerName())
	ctx.Set("Prefix", e.prefix)
	if e.production {
		ctx.Set("PrefixLabel", "")
	} else {
		ctx.Set("PrefixLabel", "["+strings.ToUpper(e.prefix)+"] ")
	}
	tName := ctx.String("Theme", e.eb.theme)
	if _, ok := e.eb.theming[tName]; ok {
		ctx.Set("Theme", tName)
	} else {
		if tNames := e.ThemeNames(); len(tNames) > 0 {
			ctx.Set("Theme", tNames[0])
		} else {
			ctx.Set("Theme", "")
		}
	}
	now := time.Now()
	ctx.Set("Year", now.Year())
	ctx.Set("Release", globals.BinHash)
	ctx.Set("Version", globals.Version)
	return
}

func (e *Enjin) FindTranslations(url string) (pages []*page.Page) {
	for _, f := range e.Features() {
		if provider, ok := f.(feature.PageProvider); ok {
			if found := provider.FindTranslations(url); len(found) > 0 {
				pages = append(pages, found...)
			}
		}
	}
	for _, pg := range e.eb.pages {
		if _, ok := pg.Match(url); ok {
			pages = append(pages, pg.Copy())
		}
	}
	return
}

func (e *Enjin) FindPage(tag language.Tag, url string) (p *page.Page) {
	for _, f := range e.Features() {
		if provider, ok := f.(feature.PageProvider); ok {
			if p = provider.FindPage(tag, url); p != nil {
				return
			}
		}
	}
	for _, pg := range e.eb.pages {
		if pg.Url == url && language.Compare(pg.LanguageTag, tag) {
			p = pg
			break
		}
	}
	return
}