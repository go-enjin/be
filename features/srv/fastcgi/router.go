//go:build srv_fastcgi || srv || all

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
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yookoala/gofast"

	clPath "github.com/go-corelibs/path"
)

// TODO: figure out fix for wp urls including index.php, ie: `/index.php/category/slug...`
//       gofast: unable to process request error access path outside of filesystem docroot: /.../src/github.com/go-enjin/be-wordpress/docroot - /wordpress/2023/07/28/hello-world

var (
	DefaultDirIndex = "index.php"
)

var rxPathInfo = regexp.MustCompile(`^(.+\.php)(/?.+)$`)

type phpFS struct {
	DocRoot    string
	DirIndex   string
	EnvKeys    []string
	Extensions []string
}

func newPhpFS(dirIndex, root string, envKeys []string) gofast.Middleware {
	if dirIndex == "" {
		dirIndex = DefaultDirIndex
	}
	fs := &phpFS{
		DocRoot:    root,
		DirIndex:   dirIndex,
		EnvKeys:    envKeys,
		Extensions: []string{"php"},
	}
	return gofast.Chain(
		gofast.BasicParamsMap,
		gofast.MapHeader,
		fs.Router(),
	)
}

func (fs *phpFS) Router() gofast.Middleware {
	docroot := filepath.Join(fs.DocRoot) // converts to absolute path
	return func(inner gofast.SessionHandler) gofast.SessionHandler {
		return func(client gofast.Client, req *gofast.Request) (*gofast.ResponsePipe, error) {

			// define some required cgi parameters
			// with the given http request
			fastcgiScriptName := req.Raw.URL.Path

			var fastcgiPathInfo string
			if matches := rxPathInfo.FindStringSubmatch(fastcgiScriptName); len(matches) > 0 {
				fastcgiScriptName, fastcgiPathInfo = matches[1], matches[2]
			}

			// If accessing a directory, try accessing document index file

			docRootFileName := filepath.Join(fs.DocRoot, fastcgiScriptName)
			if clPath.IsDir(docRootFileName) {
				fastcgiScriptName = filepath.Join(docRootFileName, fs.DirIndex)
				if !clPath.IsFile(fastcgiScriptName) {
					fastcgiScriptName = filepath.Join(fs.DocRoot, fs.DirIndex)
				}
			} else if clPath.IsFile(docRootFileName) {
				fastcgiScriptName = docRootFileName
			}

			for _, key := range fs.EnvKeys {
				req.Params[key] = os.Getenv(key)
			}

			req.Params["PATH_INFO"] = fastcgiPathInfo
			req.Params["PATH_TRANSLATED"] = req.Raw.URL.Path
			req.Params["SCRIPT_NAME"] = fastcgiScriptName
			req.Params["SCRIPT_FILENAME"] = fastcgiScriptName
			req.Params["DOCUMENT_URI"] = req.Raw.URL.Path
			req.Params["DOCUMENT_ROOT"] = docroot

			// check if the script filename is within docroot.
			// triggers error if not.
			if !strings.HasPrefix(req.Params["SCRIPT_FILENAME"], docroot) {
				err := fmt.Errorf("error access path outside of filesystem docroot: %v - %v", docroot, req.Params["SCRIPT_FILENAME"])
				return nil, err
			}

			// handle directory index

			return inner(client, req)
		}
	}
}
