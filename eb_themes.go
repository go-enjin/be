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
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/theme"
	types "github.com/go-enjin/be/pkg/types/theme-types"
)

func (eb *EnjinBuilder) SetTheme(name string) feature.Builder {
	if _, ok := eb.theming[name]; ok {
		eb.theme = name
	} else {
		log.FatalDF(1, `theme not found: "%v"`, name)
	}
	return eb
}

func (eb *EnjinBuilder) AddTheme(t *theme.Theme) feature.Builder {
	eb.theming[t.Name] = t
	if lfs, ok := t.Locales(); ok {
		eb.localeFiles = append(eb.localeFiles, lfs)
		log.DebugF("including %v theme locales", t.Name)
	}
	for _, f := range eb.features {
		if fp, ok := f.This().(types.FormatProvider); ok {
			eb.theming[t.Name].FormatProviders = append(eb.theming[t.Name].FormatProviders, fp)
		}
	}
	return eb
}