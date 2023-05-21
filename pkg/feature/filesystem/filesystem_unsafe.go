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

package filesystem

import (
	"os"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/lang"
	bePath "github.com/go-enjin/be/pkg/path"
)

func (f *CFeature[MakeTypedFeature]) findPageMatterPathMount(path string) (realpath string, mountPoint *CMountPoint, locale language.Tag, err error) {

	var ok bool
	var uri, modified string
	theme := f.Enjin.MustGetTheme()
	formats := theme.ListFormats()

	tag := language.Und
	defLang := f.Enjin.SiteDefaultLanguage()

	uri = bePath.CleanWithSlash(path)
	if tag, modified, ok = lang.ParseLangPath(uri); ok {
		uri = modified
	}

	undSrc := bePath.CleanWithSlash(uri)

	switch {
	case tag != language.Und:
		// check for the specific language
		undSrc = "/" + tag.String() + undSrc
		for _, mp := range f.FindPathMountPoint(undSrc) {
			if realpath, err = mp.ROFS.FindFilePath(undSrc, formats...); err == nil {
				mountPoint = mp
				locale = tag
				return
			}
		}
		fallthrough

	default:
		// check for the default language
		defSrc := "/" + defLang.String() + undSrc
		for _, mp := range f.FindPathMountPoint(undSrc) {
			if realpath, err = mp.ROFS.FindFilePath(defSrc, formats...); err == nil {
				mountPoint = mp
				locale = defLang
				return
			}
		}
		// check for the undefined language
		for _, mp := range f.FindPathMountPoint(undSrc) {
			if realpath, err = mp.ROFS.FindFilePath(undSrc, formats...); err == nil {
				mountPoint = mp
				locale = language.Und
				return
			}
		}
	}

	err = os.ErrNotExist
	return
}