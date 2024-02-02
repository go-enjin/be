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

package fs

import (
	clPath "github.com/go-corelibs/path"
	"github.com/go-enjin/be/pkg/forms"
)

func EnumerateCheckPaths(src string) (checkPaths []string) {
	var extensions []string
	if extn := clPath.Ext(src); extn != "" {
		switch extn {
		case "css":
			extensions = append(extensions, "scss", "sass", "css")
		default:
			extensions = append(extensions, extn)
		}
		trimPath := clPath.TrimExt(src)
		for _, ext := range extensions {
			checkPaths = append(checkPaths, trimPath+"."+ext)
		}
		return
	}
	checkPaths = append(checkPaths, src)
	return
}

func CheckForFileData(fs FileSystem, url, mount string) (data []byte, mime, path string, ok bool) {
	p := forms.TrimQueryParams(url)
	p = clPath.TrimPrefix(p, mount)
	p = clPath.TrimSlashes(p)
	checkPaths := EnumerateCheckPaths(p)
	var err error
	for _, checkPath := range checkPaths {
		// log.DebugF("checking for %v file data in %v for %v (%v)", fs.Name(), mount, checkPath, p)
		if data, err = fs.ReadFile(checkPath); err == nil {
			mime, _ = fs.MimeType(checkPath)
			path = checkPath
			ok = true
			return
		}
	}
	return
}
