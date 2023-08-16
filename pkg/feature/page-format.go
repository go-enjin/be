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
	"html/template"

	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
)

type PageFormat interface {
	This() interface{}
	Name() (name string)
	Label() (label string)
	Extensions() (extensions []string)
	Prepare(ctx context.Context, content string) (out context.Context, err error)
	Process(ctx context.Context, content string) (html template.HTML, redirect string, err *EnjinError)
	IndexDocument(pg interface{}) (doc interface{}, err error)
	NewDocumentMapping(tag language.Tag) (doctype, analyzer string, dm *mapping.DocumentMapping)
}

type PageFormatProvider interface {
	ListFormats() (names []string)
	GetFormat(name string) (format PageFormat)
	MatchFormat(filename string) (format PageFormat, match string)
}