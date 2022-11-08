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

package theme

import (
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/fs/local"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
)

func NewLocal(path string) (t *Theme, err error) {
	if !bePath.IsDir(path) {
		err = bePath.ErrorDirNotFound
		return
	}
	t = new(Theme)
	t.Path = bePath.TrimSlashes(path)
	if t.FileSystem, err = local.New(path); err != nil {
		return
	}
	if staticFs, e := local.New(path + "/static"); e == nil {
		t.StaticFS = staticFs
		fs.RegisterFileSystem("/", staticFs)
		// log.DebugF("registered local static fs: %v/static", path)
	} else {
		t.StaticFS = nil
	}

	t.Name = bePath.Base(path)
	if found := getThemeInstance(t.Name); found != nil {
		t = found
		log.DebugF("found existing theme: %v", t.Name)
		return
	}

	err = t.init()
	return
}