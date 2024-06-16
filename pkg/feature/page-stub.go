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

	"github.com/go-corelibs/x-text/language"
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
	Origin   string        `json:"origin"`
	FS       fs.FileSystem `json:"fs"`
	Point    string        `json:"point"`
	Shasum   string        `json:"shasum"`
	Source   string        `json:"source"`
	Language language.Tag  `json:"language"`
	Fallback language.Tag  `json:"fallback"`
}

func NewPageStub(origin string, bfs fs.FileSystem, point, source, shasum string, fallback language.Tag) (s *PageStub) {
	s = &PageStub{
		Origin:   origin,
		FS:       bfs,
		Point:    point,
		Shasum:   shasum,
		Source:   source,
		Language: fallback,
		Fallback: fallback,
	}
	return
}

func MakePageStub(data []byte) (s *PageStub, err error) {
	s = &PageStub{}
	err = s.Unmarshal(data)
	return
}

type encodedPageStub struct {
	Origin   string `json:"origin"`
	FS       string `json:"fs"`
	Point    string `json:"point"`
	Shasum   string `json:"shasum"`
	Source   string `json:"source"`
	Language string `json:"language"`
	Fallback string `json:"fallback"`
}

func (ps *PageStub) Marshal() (data []byte, err error) {
	data, err = json.Marshal(encodedPageStub{
		Origin:   ps.Origin,
		FS:       ps.FS.ID(),
		Point:    ps.Point,
		Shasum:   ps.Shasum,
		Source:   ps.Source,
		Language: ps.Language.String(),
		Fallback: ps.Fallback.String(),
	})
	return
}

func (ps *PageStub) Unmarshal(data []byte) (err error) {
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
	if f, ok := fs.GetFileSystem(es.FS); ok {
		ps.FS = f
	} else {
		err = fmt.Errorf("filesystem not found: %v", es.FS)
	}
	return
}

func (ps *PageStub) MarshalJSON() (data []byte, err error) {
	return ps.Marshal()
}

func (ps *PageStub) UnmarshalJSON(data []byte) (err error) {
	return ps.Unmarshal(data)
}

func (ps *PageStub) MarshalBinary() (data []byte, err error) {
	return ps.Marshal()
}

func (ps *PageStub) UnmarshalBinary(data []byte) (err error) {
	return ps.Unmarshal(data)
}

func (ps *PageStub) Copy() *PageStub {
	return &PageStub{
		Origin:   ps.Origin,
		FS:       ps.FS,
		Point:    ps.Point,
		Shasum:   ps.Shasum,
		Source:   ps.Source,
		Language: ps.Language,
		Fallback: ps.Fallback,
	}
}

func (ps *PageStub) DeepCopy() interface{} {
	return ps.Copy()
}
