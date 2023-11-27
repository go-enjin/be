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

package pages

import (
	"path/filepath"
	"strings"

	bePath "github.com/go-enjin/be/pkg/path"
)

func GetUrlPathSectionSlug(url string) (fullpath, path, section, slug string) {
	var notPath bool
	if notPath = strings.HasPrefix(url, "!"); notPath {
		url = url[1:]
	}
	fullpath = bePath.TrimSlashes(url)
	fullpath = strings.ToLower(fullpath)
	if path = filepath.Dir(fullpath); path == "." {
		path = "/"
	} else {
		path = bePath.CleanWithSlash(path)
	}
	slug = filepath.Base(fullpath)
	if parts := strings.Split(fullpath, "/"); len(parts) > 1 {
		section = parts[0]
	} else {
		section = ""
	}
	if notPath {
		fullpath = "!" + fullpath
	} else {
		fullpath = "/" + fullpath
	}
	fullpath = filepath.Clean(fullpath)
	return
}
