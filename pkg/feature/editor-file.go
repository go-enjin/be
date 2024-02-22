// Copyright (c) 2024  The Go-Enjin Authors
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

package feature

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/go-corelibs/mime"
	clPath "github.com/go-corelibs/path"
	"github.com/go-corelibs/x-text/language"
	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
)

type EditorFile struct {
	FSBT   string        `json:"fsbt"`
	FSID   string        `json:"fsid"`
	Code   string        `json:"code"`
	Path   string        `json:"path"`
	File   string        `json:"file"`
	Locale *language.Tag `json:"lang"`

	MountPoint interface{} `json:"-"`
	Tilde      string      `json:"-"`

	Base     string `json:"base"`
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

	Actions    editor.Actions    `json:"actions"`
	Indicators editor.Indicators `json:"indicators,omitempty"`

	Context beContext.Context `json:"-"`
}

func ParseDirectory(fsid, filePath string) *EditorFile {
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

	return &EditorFile{
		FSID:     fsid,
		Code:     topDir,
		Path:     dirs,
		Locale:   &locale,
		MimeType: "inode/directory",
	}
}

func ParseFile(fsid, filePath string, t Theme) *EditorFile {
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
		if v, wf, ok := editor.ParseEditorWorkFile(file); ok {
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

	var base string
	if pf, match := t.MatchFormat(name); pf != nil {
		base = strings.TrimSuffix(name, "."+match)
	} else {
		base = name
	}

	return &EditorFile{
		FSID:     fsid,
		Code:     code,
		Path:     dirs,
		File:     file,
		Base:     base,
		Name:     name,
		Tilde:    tilde,
		Locale:   &locale,
		MimeType: mime.FromPathOnly(file),
	}
}

func (f *EditorFile) DirectoryPath() (dirPath string) {
	if f.Path != "" && f.Path != "." && f.Path != "/" {
		dirPath = f.Path
	}
	return
}

func (f *EditorFile) FileName() (name string) {
	name = filepath.Base(f.File)
	return
}

func (f *EditorFile) BaseName() (fileName string) {
	fileName = clPath.Base(f.File)
	return
}

func (f *EditorFile) BaseNamePath() (filePath string) {
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

func (f *EditorFile) FilePath() (filePath string) {
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

func (f *EditorFile) Url() (path string) {
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

func (f *EditorFile) EditPath() (filePath string) {
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

func (f *EditorFile) EditFilePath() (filePath string) {
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

func (f *EditorFile) EditDirectoryPath() (directory string) {
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

func (f *EditorFile) EditParentDirectoryPath() (directory string) {
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

func (f *EditorFile) localeOrEmpty() (value string) {
	if f.Locale != nil {
		value = f.Locale.String()
	}
	return
}

func (f *EditorFile) CodeFilePath() (filePath string) {
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

func (f *EditorFile) EditCodeFilePath() (filePath string) {
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

func (f *EditorFile) EditCodeDirectoryPath() (directory string) {
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

func (f *EditorFile) EditCodeParentDirectoryPath() (directory string) {
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

func (f *EditorFile) Clone() (file *EditorFile) {
	locale := *f.Locale
	file = &EditorFile{
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
		Actions:    append(editor.Actions{}, f.Actions...),
		Indicators: append(editor.Indicators{}, f.Indicators...),
	}
	return
}
