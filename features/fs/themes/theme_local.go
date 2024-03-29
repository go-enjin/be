//go:build (fs_theme && (drivers_fs_local || drivers_fs || drivers || locals)) || all

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

package themes

import (
	"fmt"

	clPath "github.com/go-corelibs/path"
	"github.com/go-enjin/be/drivers/fs/local"
	"github.com/go-enjin/be/pkg/log"
)

type ThemeLocalSupport interface {
	// LocalTheme constructs a local theme.Theme instance and adds it to the
	// enjin during the build phase
	LocalTheme(path string) MakeFeature

	// LocalThemes constructs all local theme.Theme instances found and adds
	// them to the enjin during the build phase
	LocalThemes(path string) MakeFeature
}

func (f *CFeature) loadLocalTheme(path string) (err error) {
	if !clPath.IsDir(path) {
		err = fmt.Errorf("directory not found: %v", path)
		return
	}

	stub := &loadTheme{
		support: "local",
		path:    path,
	}

	if stub.rwfs, err = local.New(f.Tag().String(), path); err != nil {
		err = fmt.Errorf("error mounting local filesystem: %v - %v", path, err)
		return
	}
	stub.themeFs = stub.rwfs

	if staticPath := path + "/static"; clPath.IsDir(staticPath) {
		if stub.staticFs, err = local.New(f.Tag().String(), staticPath); err != nil {
			stub.staticFs = nil
			err = nil
		}
	}

	f.loading = append(f.loading, stub)

	return
}

func (f *CFeature) LocalTheme(path string) MakeFeature {
	var err error

	if err = f.loadLocalTheme(path); err != nil {
		log.FatalDF(1, "%v", err)
	}

	return f
}

func (f *CFeature) LocalThemes(path string) MakeFeature {
	var err error
	var paths []string
	if paths, err = clPath.ListDirs(path, true); err != nil {
		log.FatalF("error listing path: %v", err)
		return nil
	}
	for _, p := range paths {
		if e := f.loadLocalTheme(p); e != nil {
			log.FatalDF(1, "%s", err)
		}
	}
	return f
}
