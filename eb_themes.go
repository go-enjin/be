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
)

func (eb *EnjinBuilder) SetTheme(name string) feature.Builder {
	if _, ok := eb.theming[name]; ok {
		eb.theme = name
	} else {
		log.FatalDF(1, `theme not found: "%v"`, name)
	}
	return eb
}

func (eb *EnjinBuilder) AddTheme(t feature.Theme) feature.Builder {
	name := t.Name()
	eb.theming[name] = t
	eb.themeOrder = append(eb.themeOrder, name)
	log.DebugF("adding %v theme", name)
	return eb
}