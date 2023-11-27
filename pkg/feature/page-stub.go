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
	"strings"

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
	Origin   string
	FS       fs.FileSystem
	Point    string
	Shasum   string
	Source   string
	Language language.Tag
	Fallback language.Tag
	EnjinCtx beContext.Context
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

func (ps *PageStub) MarshalBinary() (data []byte, err error) {
	return ps.GobEncode()
}

func (ps *PageStub) UnmarshalBinary(data []byte) (err error) {
	return ps.GobDecode(data)
}

func (ps *PageStub) GobEncode() (data []byte, err error) {
	var ctx []byte
	if ctx, err = ps.EnjinCtx.AsJSON(); err != nil {
		return
	}
	parts := []string{
		ps.Origin,
		ps.FS.ID(),
		ps.Point,
		ps.Shasum,
		ps.Source,
		ps.Language.String(),
		ps.Fallback.String(),
		string(ctx),
	}
	text := strings.Join(parts, "\n")
	data = []byte(text)
	return
}

func (ps *PageStub) GobDecode(data []byte) (err error) {
	text := string(data)
	parts := strings.Split(text, "\n")
	if len(parts) != 8 {
		err = fmt.Errorf("invalid number of data parts: %#+v", parts)
		return
	}
	ps.Origin = parts[0]
	id := parts[1]
	if v, ok := fs.GetFileSystem(id); ok {
		ps.FS = v
	} else {
		err = fmt.Errorf("filesystem not found: %v", id)
		return
	}
	ps.Point = parts[2]
	ps.Shasum = parts[3]
	ps.Source = parts[4]
	var lt language.Tag
	if lt, err = language.Parse(parts[5]); err != nil {
		err = fmt.Errorf("error parsing filesystem language tag: %v - %v", parts[5], err)
		return
	}
	ps.Language = lt
	var ft language.Tag
	if ft, err = language.Parse(parts[6]); err != nil {
		err = fmt.Errorf("error parsing filesystem language tag: %v - %v", parts[6], err)
		return
	}
	ps.Fallback = ft
	ps.EnjinCtx = beContext.Context{}
	if err = json.Unmarshal([]byte(parts[7]), &ps.EnjinCtx); err != nil {
		err = fmt.Errorf("error parsing filesystem context: %v", err)
		return
	}
	return
}
