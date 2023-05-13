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

package matter

import (
	"encoding/gob"

	"github.com/go-enjin/golang-org-x-text/language"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/fs"
)

func init() {
	gob.Register(&PageStub{})
}

type ValueStubPair struct {
	Value interface{}
	Stub  *PageStub
}

type PageStub struct {
	Bfs      fs.FileSystem
	Point    string
	Shasum   string
	Source   string
	Language language.Tag
	Fallback language.Tag
	EnjinCtx beContext.Context
}

func NewPageStub(enjin beContext.Context, bfs fs.FileSystem, point, source, shasum string, fallback language.Tag) (s *PageStub, err error) {
	s = &PageStub{
		Bfs:      bfs,
		Point:    point,
		Shasum:   shasum,
		Source:   source,
		Language: fallback,
		Fallback: fallback,
		EnjinCtx: enjin,
	}
	return
}