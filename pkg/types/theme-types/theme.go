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

package types

import (
	"html/template"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/fs"
)

type Theme interface {
	FS() fs.FileSystem
	GetParentTheme() (parent Theme)
	GetBlockThemeNames() (names []string)
	NewHtmlTemplate(name string) (tmpl *template.Template, err error)
	NewHtmlTemplateWithContext(name string, ctx context.Context) (tmpl *template.Template, err error)
	NewFuncMapWithContext(ctx context.Context) (fm template.FuncMap)
}