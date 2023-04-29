//go:build fastcgi || all

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

package fastcgi

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yookoala/gofast"

	bePath "github.com/go-enjin/be/pkg/path"
)

var rxPathInfo = regexp.MustCompile(`^(.+\.php)(/?.+)$`)

type FileSystemRouter struct {
	DocRoot  string
	Exts     []string
	DirIndex []string
}

func NewPHPFS(root string) gofast.Middleware {
	fs := &FileSystemRouter{
		DocRoot:  root,
		Exts:     []string{"php"},
		DirIndex: []string{"index.php"},
	}
	return gofast.Chain(
		gofast.BasicParamsMap,
		gofast.MapHeader,
		fs.Router(),
	)
}

func (fs *FileSystemRouter) Router() gofast.Middleware {
	docroot := filepath.Join(fs.DocRoot) // converts to absolute path
	return func(inner gofast.SessionHandler) gofast.SessionHandler {
		return func(client gofast.Client, req *gofast.Request) (*gofast.ResponsePipe, error) {

			// define some required cgi parameters
			// with the given http request
			r := req.Raw
			fastcgiScriptName := r.URL.Path

			var fastcgiPathInfo string
			if matches := rxPathInfo.Copy().FindStringSubmatch(fastcgiScriptName); len(matches) > 0 {
				fastcgiScriptName, fastcgiPathInfo = matches[1], matches[2]
			}

			// If accessing a directory, try accessing document index file

			docRootFileName := filepath.Join(fs.DocRoot, fastcgiScriptName)
			if bePath.IsDir(docRootFileName) {
				fastcgiScriptName = filepath.Join(fastcgiScriptName, "index.php")
			}
			// if strings.HasSuffix(fastcgiScriptName, "/") {
			// 	fastcgiScriptName = filepath.Join(fastcgiScriptName, "index.php")
			// }

			req.Params["PATH_INFO"] = fastcgiPathInfo
			req.Params["PATH_TRANSLATED"] = filepath.Join(docroot, fastcgiPathInfo)
			req.Params["SCRIPT_NAME"] = fastcgiScriptName
			req.Params["SCRIPT_FILENAME"] = filepath.Join(docroot, fastcgiScriptName)
			req.Params["DOCUMENT_URI"] = r.URL.Path
			req.Params["DOCUMENT_ROOT"] = docroot

			// check if the script filename is within docroot.
			// triggers error if not.
			if !strings.HasPrefix(req.Params["SCRIPT_FILENAME"], docroot) {
				err := fmt.Errorf("error: access path outside of filesystem docroot")
				return nil, err
			}

			// handle directory index

			return inner(client, req)
		}
	}
}