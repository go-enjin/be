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

	"github.com/go-enjin/be/drivers/fs/local"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/theme"
)

type ThemeLocalSupport interface {
	// LocalTheme constructs a local theme.Theme instance and adds it to the
	// enjin during the build phase
	LocalTheme(path string) MakeFeature

	// LocalThemes constructs all local theme.Theme instances found and adds
	// them to the enjin during the build phase
	LocalThemes(path string) MakeFeature
}

func (f *CFeature) loadLocalTheme(path string) (t *theme.Theme, err error) {
	if !bePath.IsDir(path) {
		err = fmt.Errorf("directory not found: %v", path)
		return
	}
	var themeFs, staticFs fs.FileSystem
	if themeFs, err = local.New(f.Tag().String(), path); err != nil {
		err = fmt.Errorf("error mounting local filesystem: %v - %v", path, err)
		return
	}
	if staticFs, err = local.New(f.Tag().String(), path+"/static"); err == nil {
		fs.RegisterFileSystem("/", staticFs)
	} else {
		staticFs = nil
	}

	if t, err = theme.New(f.Tag().String(), path, themeFs, staticFs); err != nil {
		err = fmt.Errorf("error loading theme: %v - %v", path, err)
		return
	}

	return
}

func (f *CFeature) LocalTheme(path string) MakeFeature {
	var err error
	var t *theme.Theme

	if t, err = f.loadLocalTheme(path); err != nil {
		log.FatalDF(1, "%v", err)
	}

	log.DebugF("loaded local theme: %v", t.Name)

	f.AddTheme(t)
	return f
}

func (f *CFeature) LocalThemes(path string) MakeFeature {
	var err error
	var paths []string
	if paths, err = bePath.ListDirs(path); err != nil {
		log.FatalF("error listing path: %v", err)
		return nil
	}
	for _, p := range paths {
		if t, e := f.loadLocalTheme(p); e != nil {
			log.FatalDF(1, "%s", err)
		} else {
			f.themes[t.Name] = t
			log.DebugF("loaded local theme: %v", t.Name)
		}
	}
	return f
}