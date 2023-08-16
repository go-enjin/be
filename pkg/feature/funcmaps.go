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

package feature

import (
	htmlTemplate "html/template"
	textTemplate "text/template"

	"github.com/go-enjin/be/pkg/context"
)

type FuncMapProvider interface {
	Feature

	MakeFuncMap(ctx context.Context) (fm FuncMap)
}

type FuncMap textTemplate.FuncMap

func (fm FuncMap) Apply(other FuncMap) {
	for k, v := range other {
		fm[k] = v
	}
}

func (fm FuncMap) AsTEXT() textTemplate.FuncMap {
	return textTemplate.FuncMap(fm)
}

func (fm FuncMap) AsHTML() htmlTemplate.FuncMap {
	return htmlTemplate.FuncMap(fm)
}