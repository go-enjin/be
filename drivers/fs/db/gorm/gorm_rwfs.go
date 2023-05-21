//go:build driver_fs_db_gorm || drivers_fs_db || drivers_fs || drivers || all

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

package gorm

import (
	"fmt"
	"os"

	"github.com/gabriel-vasile/mimetype"

	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/page/matter"
)

func (f *DBFileSystem) MakeDir(path string, _ os.FileMode) (err error) {
	err = f.MakeDirAll(path, 0)
	return
}

func (f *DBFileSystem) MakeDirAll(path string, _ os.FileMode) (err error) {
	f.Lock()
	defer f.Unlock()
	// TODO: are directory inodes needed?
	var entry *File
	if entry, err = f.getEntryUnsafe(path); err == nil {
		if entry.Mime == InodeDirectoryMimeType {
			// err = fmt.Errorf("directory exists already")
			return
		}
		err = fmt.Errorf("path is a File")
	} else {
		entry = &File{
			Path:    path,
			Mime:    InodeDirectoryMimeType,
			Shasum:  "",
			Content: []byte{},
			Context: []byte{},
		}
		err = f.tx().Save(entry).Error
	}
	return
}

func (f *DBFileSystem) Remove(path string) (err error) {
	f.Lock()
	defer f.Unlock()
	realpath := f.realpath(path)
	err = f.tx().Where(`path = ?`, realpath).Delete(&File{}).Error
	return
}

func (f *DBFileSystem) RemoveAll(path string) (err error) {
	f.Lock()
	defer f.Unlock()
	realpath := f.realpath(path)
	err = f.tx().Where(`path LIKE ?`, realpath+"%").Delete(&File{}).Error
	return
}

func (f *DBFileSystem) WriteFile(path string, data []byte, _ os.FileMode) (err error) {
	f.Lock()
	defer f.Unlock()
	realpath := f.realpath(path)
	var shasum string
	if shasum, err = sha.DataHash64(data); err != nil {
		return
	}
	mime := mimetype.Detect(data).String()
	var entry *File
	if entry, err = f.getEntryUnsafe(path); err != nil {
		err = nil
		entry = &File{
			Path:    realpath,
			Mime:    mime,
			Shasum:  shasum,
			Content: data,
			// Context: ctx,
		}
	} else {
		entry.Path = realpath
		entry.Mime = mime
		entry.Shasum = shasum
		entry.Content = data
		// entry.Context = ctx
	}
	err = f.tx().Save(entry).Error
	return
}

func (f *DBFileSystem) WritePageMatter(pm *matter.PageMatter) (err error) {
	f.Lock()
	defer f.Unlock()
	realpath := f.realpath(pm.Path)

	var data []byte
	if data, err = pm.Bytes(); err != nil {
		err = fmt.Errorf("error getting bytes from page matter: %v", err)
		return
	}

	var shasum string
	if shasum, err = sha.DataHash64(data); err != nil {
		return
	}

	mime := mimetype.Detect(data).String()

	jsonMatter, _ := pm.Matter.AsJSON()

	var entry *File
	if entry, err = f.getEntryUnsafe(pm.Path); err != nil {
		log.DebugDF(2, "creating new entry: %v", realpath)
		err = nil
		entry = &File{
			Path:    realpath,
			Mime:    mime,
			Shasum:  shasum,
			Content: []byte(pm.Body),
			Context: jsonMatter,
		}
	} else {
		log.DebugDF(2, "updating existing entry: %v", realpath)
		entry.Path = realpath
		entry.Mime = mime
		entry.Shasum = shasum
		entry.Content = []byte(pm.Body)
		entry.Context = jsonMatter
	}

	err = f.tx().Save(entry).Error
	return
}

func (f *DBFileSystem) RemovePageMatter(path string) (err error) {
	err = f.Remove(path)
	return
}