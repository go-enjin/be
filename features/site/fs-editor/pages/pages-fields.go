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

package pages

import (
	"net/http"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/path"
)

func (f *CFeature) MakePageArchetypeContextFields(r *http.Request, name string) (fields beContext.Fields) {

	tc := f.Enjin.MustGetTheme().GetConfig()
	fields = beContext.Fields{}
	basename := path.Base(name)

	if found, ok := tc.Supports.Archetypes[basename]; ok {
		// general fields for any format of archetype
		for k, v := range found {
			fields[k] = v
		}
	}

	if basename != name {
		if found, ok := tc.Supports.Archetypes[name]; ok {
			// fields for a specific archetype, clobbering generals
			for k, v := range found {
				fields[k] = v
			}
		}
	}

	printer := lang.GetPrinterFromRequest(r)
	parsers := f.Enjin.PageContextParsers()
	fields.Init(printer, parsers)
	return
}

func (f *CFeature) MakePageContextFields(r *http.Request, archetype string) (fields beContext.Fields) {
	fields = f.Enjin.MakePageContextFields(r)
	for k, v := range f.MakePageArchetypeContextFields(r, archetype) {
		if _, present := fields[k]; !present {
			// no clobbering allowed here
			fields[k] = v
		}
	}
	return
}