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

package page

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/format"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func NewFromPageStub(s *fs.PageStub, formats format.PageFormatProvider) (p *Page, err error) {
	var data []byte
	if data, err = s.FS.ReadFile(s.Source); err != nil {
		err = fmt.Errorf("error reading %v mount file: %v - %v", s.FS.Name(), s.Source, err)
		return
	}

	path := beStrings.TrimPrefixes(s.Source, s.Fallback.String())
	var epoch, created, updated int64

	if epoch, err = s.FS.FileCreated(s.Source); err == nil {
		created = epoch
	} else {
		log.ErrorF("error getting page created epoch: %v", err)
	}

	if epoch, err = s.FS.LastModified(s.Source); err == nil {
		updated = epoch
	} else {
		log.ErrorF("error getting page last modified epoch: %v", err)
	}

	if p, err = New(s.Origin, path, string(data), created, updated, formats, s.EnjinCtx); err == nil {
		if language.Compare(p.LanguageTag, language.Und) {
			p.SetLanguage(s.Fallback)
		}
		if !strings.HasPrefix(p.Url, "!") {
			p.SetSlugUrl(filepath.Clean(s.Point + p.Url))
		}
		// log.DebugF("made page from %v stub: [%v] %v (%v)", s.FS.Name(), p.Language, s.Source, p.Url)
	} else {
		err = fmt.Errorf("error: new %v mount page %v - %v", s.FS.Name(), path, err)
	}
	return
}