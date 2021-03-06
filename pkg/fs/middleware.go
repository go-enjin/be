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
	"github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/page"
	bePath "github.com/go-enjin/be/pkg/path"
)

func EnumerateCheckPaths(src string) (checkPaths []string) {
	var extensions []string
	if extn := bePath.Ext(src); extn != "" {
		switch extn {
		case "css":
			extensions = append(extensions, "scss", "sass", "css")
		default:
			extensions = append(extensions, extn)
		}
		trimPath := bePath.TrimExt(src)
		for _, ext := range extensions {
			checkPaths = append(checkPaths, trimPath+"."+ext)
		}
		return
	}
	checkPaths = append(checkPaths, src)
	return
}

func CheckForFileData(fs FileSystem, url, mount string) (data []byte, mime, path string, ok bool) {
	p := net.TrimQueryParams(url)
	p = bePath.TrimPrefix(p, mount)
	p = bePath.TrimSlashes(p)
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

func FindAllFilePages(fs FileSystem, mount, path string) (list []*page.Page, err error) {
	p := net.TrimQueryParams(path)
	p = bePath.TrimPrefix(p, mount)
	p = bePath.TrimSlashes(p)
	var files []string
	if files, err = fs.ListAllFiles(path); err != nil {
		return
	}
	for _, file := range files {
		var data []byte
		if data, err = fs.ReadFile(file); err != nil {
			return
		}
		var p *page.Page
		if p, err = page.NewFromString(file, string(data)); err != nil {
			return
		}
		list = append(list, p)
	}
	return
}