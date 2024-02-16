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

package editor

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/go-corelibs/x-text/language"

	"github.com/go-corelibs/mime"
	clPath "github.com/go-corelibs/path"
	beContext "github.com/go-enjin/be/pkg/context"
)

type File struct {
	FSBT   string        `json:"fsbt"`
	FSID   string        `json:"fsid"`
	Code   string        `json:"code"`
	Path   string        `json:"path"`
	File   string        `json:"file"`
	Locale *language.Tag `json:"lang"`

	MountPoint interface{} `json:"-"`
	Tilde      string      `json:"-"`

	Name     string `json:"name"`
	Shasum   string `json:"shasum"`
	MimeType string `json:"mimeType"`

	HasDraft bool   `json:"hasDraft"`
	Locked   bool   `json:"locked"`
	LockedBy string `json:"lockedBy"`
	ReadOnly bool   `json:"readOnly"`
	Binary   bool   `json:"binary"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`

	Actions    Actions    `json:"actions"`
	Indicators Indicators `json:"indicators,omitempty"`

	Context beContext.Context `json:"-"`
}

func ParseDirectory(fsid, filePath string) *File {
	topDir := clPath.TopDirectory(filePath)
	dirs := filePath
	if dirs != "" && dirs[0] == '.' {
		dirs = dirs[1:]
	}
	dirs = clPath.TrimSlashes(dirs)

	var locale language.Tag
	if topDir == "" {
		locale = language.Und
	} else if v, eee := language.Parse(topDir); eee != nil {
		locale = language.Und
	} else if locale = v; dirs != locale.String() {
		dirs = strings.TrimPrefix(dirs, locale.String()+"/")
	} else {
		dirs = ""
	}

	return &File{
		FSID:     fsid,
		Code:     topDir,
		Path:     dirs,
		Locale:   &locale,
		MimeType: "inode/directory",
	}
}

func ParseFile(fsid, filePath string) *File {
	topDir := clPath.TopDirectory(filePath)
	file := filepath.Base(filePath)
	dirs := filepath.Dir(filePath)
	if dirs != "" && dirs[0] == '.' {
		dirs = dirs[1:]
	}
	dirs = clPath.TrimSlashes(dirs)

	code := topDir
	var locale language.Tag
	if topDir == "" {
		code = "und"
		locale = language.Und
	} else if v, eee := language.Parse(topDir); eee != nil {
		locale = language.Und
	} else if locale = v; dirs != locale.String() {
		dirs = strings.TrimPrefix(dirs, locale.String()+"/")
	} else {
		dirs = ""
	}

	var name string
	var tilde string
	if file != "" {
		if v, wf, ok := ParseEditorWorkFile(file); ok {
			tilde = wf.String()
			name = filepath.Base(v)
			file = v
		} else {
			file = filepath.Base(file)
			name = file
		}
	} else if dirs != "" {
		name = filepath.Base(dirs)
	}

	return &File{
		FSID:     fsid,
		Code:     code,
		Path:     dirs,
		File:     file,
		Name:     name,
		Tilde:    tilde,
		Locale:   &locale,
		MimeType: mime.FromPathOnly(file),
	}
}

func (f *File) DirectoryPath() (dirPath string) {
	if f.Path != "" && f.Path != "." && f.Path != "/" {
		dirPath = f.Path
	}
	return
}

func (f *File) FileName() (name string) {
	name = clPath.Base(f.File)
	return
}

func (f *File) BaseName() (fileName string) {
	fileName = clPath.Base(f.File)
	return
}

func (f *File) BaseNamePath() (filePath string) {
	var parts []string
	if f.Path != "" && f.Path != "." && f.Path != "/" {
		parts = append(parts, f.Path)
	}
	if f.File != "" {
		parts = append(parts, clPath.Base(f.File))
	}
	filePath = strings.Join(parts, "/")
	return
}

func (f *File) FilePath() (filePath string) {
	var parts []string
	if value := f.Code; value != "" {
		if f.Locale != nil {
			if f.Code != language.Und.String() {
				parts = append(parts, value)
			}
		} else {
			parts = append(parts, value)
		}
	}
	if f.Path != "" && f.Path != "." && f.Path != "/" {
		parts = append(parts, clPath.TrimSlashes(f.Path))
	}
	if f.File != "" {
		parts = append(parts, f.File)
	}
	filePath = strings.Join(parts, "/")
	return
}

func (f *File) Url() (path string) {
	if f.File == "" {
		return
	}
	path = clPath.Dir(f.EditPath())
	if basename := clPath.Base(f.File); basename != "~index" && basename != "" {
		path = clPath.Dir(f.EditPath()) + "/" + basename
	}
	path = clPath.CleanWithSlash(path)
	return
}

func (f *File) EditPath() (filePath string) {
	var parts []string
	if f.Path != "" && f.Path != "/" {
		parts = append(parts, f.Path)
	}
	if f.File != "" {
		parts = append(parts, f.File)
	}
	filePath = strings.Join(parts, "/")
	return
}

func (f *File) EditFilePath() (filePath string) {
	var parts []string
	if f.FSID != "" {
		parts = append(parts, f.FSID)
	}

	if f.Locale != nil && f.Locale.String() == f.Code {
		parts = append(parts, f.Locale.String())
	} else if f.Code != "" {
		parts = append(parts, f.Code)
	}

	if f.Path != "" && f.Path != "/" {
		parts = append(parts, f.Path)
	}
	if f.File != "" {
		parts = append(parts, f.File)
	}
	filePath = strings.Join(parts, "/")
	return
}

func (f *File) EditDirectoryPath() (directory string) {
	var parts []string
	if f.FSID != "" {
		parts = append(parts, f.FSID)
	}
	if value := f.Code; value != "" {
		parts = append(parts, value)
	} else if f.Path != "" {
		parts = append(parts, "und")
	}
	if f.Path != "" && f.Path != "/" {
		parts = append(parts, f.Path)
	}
	directory = strings.Join(parts, "/")
	return
}

func (f *File) EditParentDirectoryPath() (directory string) {
	var parts []string
	if f.FSID != "" {
		parts = append(parts, f.FSID)
	}
	if value := f.Code; value != "" {
		parts = append(parts, value)
	} else if f.Path != "" {
		parts = append(parts, "und")
	}
	if f.Path != "" && f.Path != "/" {
		parts = append(parts, filepath.Dir(f.Path))
	}
	directory = strings.Join(parts, "/")
	return
}

func (f *File) localeOrEmpty() (value string) {
	if f.Locale != nil {
		value = f.Locale.String()
	}
	return
}

func (f *File) CodeFilePath() (filePath string) {
	var parts []string
	if f.Code != "" {
		parts = append(parts, f.Code)
	}
	if f.Path != "" && f.Path != "." && f.Path != "/" {
		parts = append(parts, f.Path)
	}
	if f.File != "" {
		parts = append(parts, f.File)
	}
	filePath = strings.Join(parts, "/")
	return
}

func (f *File) EditCodeFilePath() (filePath string) {
	var parts []string
	if f.FSID != "" {
		parts = append(parts, f.FSID)
	}
	if f.Code != "" {
		parts = append(parts, f.Code)
	}
	if f.Path != "" && f.Path != "/" {
		parts = append(parts, f.Path)
	}
	if f.File != "" {
		parts = append(parts, f.File)
	}
	filePath = strings.Join(parts, "/")
	return
}

func (f *File) EditCodeDirectoryPath() (directory string) {
	var parts []string
	if f.FSID != "" {
		parts = append(parts, f.FSID)
	}
	if f.Code != "" {
		parts = append(parts, f.Code)
	}
	if f.Path != "" && f.Path != "/" {
		parts = append(parts, f.Path)
	}
	directory = strings.Join(parts, "/")
	return
}

func (f *File) EditCodeParentDirectoryPath() (directory string) {
	var parts []string
	if f.FSID != "" {
		parts = append(parts, f.FSID)
	}
	if f.Code != "" {
		parts = append(parts, f.Code)
	}
	if f.Path != "" && f.Path != "/" {
		parts = append(parts, filepath.Dir(f.Path))
	}
	directory = strings.Join(parts, "/")
	return
}

func (f *File) Clone() (file *File) {
	locale := *f.Locale
	file = &File{
		FSBT:       f.FSBT,
		FSID:       f.FSID,
		Code:       f.Code,
		Path:       f.Path,
		File:       f.File,
		Locale:     &locale,
		MountPoint: f.MountPoint,
		Tilde:      f.Tilde,
		Name:       f.Name,
		Shasum:     f.Shasum,
		MimeType:   f.MimeType,
		HasDraft:   f.HasDraft,
		Locked:     f.Locked,
		LockedBy:   f.LockedBy,
		ReadOnly:   f.ReadOnly,
		Binary:     f.Binary,
		Created:    time.UnixMicro(f.Created.UnixMicro()),
		Updated:    time.UnixMicro(f.Updated.UnixMicro()),
		Actions:    append(Actions{}, f.Actions...),
		Indicators: append(Indicators{}, f.Indicators...),
	}
	return
}
