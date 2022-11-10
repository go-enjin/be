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
	"github.com/go-enjin/be/pkg/search"
)

type Format interface {
	This() interface{}
	Name() (name string)
	Label() (label string)
	Extensions() (extensions []string)
	Process(ctx context.Context, t Theme, content string) (html template.HTML, err *EnjinError)
	IndexDocument(pg interface{}) (doc search.Document, err error)
}

type FormatProvider interface {
	GetFormat(name string) (format Format)
	MatchFormat(filename string) (format Format, match string)
}