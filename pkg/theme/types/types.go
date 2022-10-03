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

type Format interface {
	Name() (name string)
	Label() (label string)
	Process(ctx context.Context, t Theme, content string) (html template.HTML, err error)
}

type Theme interface {
	FS() fs.FileSystem
	NewHtmlTemplate(name string) *template.Template
}