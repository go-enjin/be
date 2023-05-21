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
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/theme"
)

func (eb *EnjinBuilder) AddPageFromString(path, raw string) feature.Builder {
	if eb.theme == "" {
		log.FatalDF(1, "cannot add pages before theme is set")
	}
	var ok bool
	var t *theme.Theme
	if t, ok = eb.theming[eb.theme]; !ok {
		log.FatalDF(1, "cannot add pages before theme added")
	}
	var created, updated int64
	if info, err := globals.BuildFileInfo(); err == nil {
		if info.HasBirthTime() {
			created = info.BirthTime().Unix()
		}
		updated = info.ModTime().Unix()
	}
	if p, err := page.New("enjin", path, raw, created, updated, t, eb.context); err == nil {
		eb.pages[p.Url] = p
		log.DebugF("adding page from string: %v", p.Url)
	} else {
		log.FatalF("error adding page from string: %v", err)
	}
	return eb
}

func (eb *EnjinBuilder) SetStatusPage(status int, path string) feature.Builder {
	eb.statusPages[status] = path
	return eb
}