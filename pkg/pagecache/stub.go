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

package pagecache

import (
	"fmt"
	"strings"

	"github.com/go-enjin/golang-org-x-text/language"

	beFs "github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page"
	types "github.com/go-enjin/be/pkg/types/theme-types"
)

type Stub struct {
	Bfs      beFs.FileSystem
	Point    string
	Shasum   string
	Source   string
	Language language.Tag
	Fallback language.Tag
}

func NewStub(bfs beFs.FileSystem, point, source, shasum string, fallback language.Tag) (s *Stub, p *page.Page, err error) {
	s = &Stub{
		Bfs:      bfs,
		Point:    point,
		Shasum:   shasum,
		Source:   source,
		Language: fallback,
		Fallback: fallback,
	}
	return
}

func (s *Stub) Make(formats types.FormatProvider) (p *page.Page, err error) {
	var data []byte
	if data, err = s.Bfs.ReadFile(s.Source); err != nil {
		err = fmt.Errorf("error reading %v mount file: %v - %v", s.Bfs.Name(), s.Source, err)
		return
	}

	path := trimPrefixes(s.Source, s.Fallback.String())
	var epoch, created, updated int64

	if epoch, err = s.Bfs.FileCreated(s.Source); err == nil {
		created = epoch
	} else {
		log.ErrorF("error getting page created epoch: %v", err)
	}

	if epoch, err = s.Bfs.LastModified(s.Source); err == nil {
		updated = epoch
	} else {
		log.ErrorF("error getting page last modified epoch: %v", err)
	}

	if p, err = page.New(path, string(data), s.Shasum, created, updated, formats); err == nil {
		if language.Compare(p.LanguageTag, language.Und) {
			p.SetLanguage(s.Fallback)
		}
		p.SetSlugUrl(strings.ReplaceAll(s.Point+p.Url, "//", "/"))
		// log.DebugF("made page from %v stub: [%v] %v (%v)", s.Bfs.Name(), p.Language, s.Source, p.Url)
	} else {
		err = fmt.Errorf("error: new %v mount page %v - %v", s.Bfs.Name(), path, err)
	}
	return
}