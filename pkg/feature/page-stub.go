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
	"encoding/gob"
	"encoding/json"
	"fmt"

	"github.com/go-enjin/golang-org-x-text/language"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/fs"
)

func init() {
	gob.Register(PageStub{})
}

type ValueStubPair struct {
	Value interface{}
	Stub  *PageStub
}

type PageStub struct {
	Origin   string            `json:"origin"`
	FS       fs.FileSystem     `json:"fs"`
	Point    string            `json:"point"`
	Shasum   string            `json:"shasum"`
	Source   string            `json:"source"`
	Language language.Tag      `json:"language"`
	Fallback language.Tag      `json:"fallback"`
	EnjinCtx beContext.Context `json:"enjin-ctx"`
}

func NewPageStub(origin string, enjin beContext.Context, bfs fs.FileSystem, point, source, shasum string, fallback language.Tag) (s *PageStub) {
	s = &PageStub{
		Origin:   origin,
		FS:       bfs,
		Point:    point,
		Shasum:   shasum,
		Source:   source,
		Language: fallback,
		Fallback: fallback,
		EnjinCtx: enjin,
	}
	return
}

type encodedPageStub struct {
	Origin   string            `json:"origin"`
	FS       string            `json:"fs"`
	Point    string            `json:"point"`
	Shasum   string            `json:"shasum"`
	Source   string            `json:"source"`
	Language string            `json:"language"`
	Fallback string            `json:"fallback"`
	EnjinCtx beContext.Context `json:"enjin-ctx"`
}

func (ps *PageStub) MarshalBinary() (data []byte, err error) {
	data, err = json.Marshal(encodedPageStub{
		Origin:   ps.Origin,
		FS:       ps.FS.ID(),
		Point:    ps.Point,
		Shasum:   ps.Shasum,
		Source:   ps.Source,
		Language: ps.Language.String(),
		Fallback: ps.Fallback.String(),
		EnjinCtx: ps.EnjinCtx,
	})
	return
}

func (ps *PageStub) UnmarshalBinary(data []byte) (err error) {
	var es encodedPageStub
	if err = json.Unmarshal(data, &es); err != nil {
		return
	}
	ps.Origin = es.Origin
	ps.Point = es.Point
	ps.Shasum = es.Shasum
	ps.Source = es.Source
	ps.Language, _ = language.Parse(es.Language)
	ps.Fallback, _ = language.Parse(es.Fallback)
	ps.EnjinCtx = es.EnjinCtx
	if f, ok := fs.GetFileSystem(es.FS); ok {
		ps.FS = f
	} else {
		err = fmt.Errorf("filesystem not found: %v", es.FS)
	}
	return
}
