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
	"time"

	"github.com/gabriel-vasile/mimetype"

	clPath "github.com/go-corelibs/path"
	sha "github.com/go-corelibs/shasum"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/types/page/matter"
)

func (f *DBFileSystem) BeginTransaction() {
	f.Lock()
	defer f.Unlock()
	if f.tx != nil {
		//log.FatalDF(1, "transactions within transactions not implemented")
	} else {
		f.tx = f.tableScoped().Begin()
		//f.tx = f.db.Begin()
	}
}

func (f *DBFileSystem) RollbackTransaction() {
	f.Lock()
	defer f.Unlock()
	if f.tx != nil {
		f.tx.Rollback()
		f.tx = nil
	}
}

func (f *DBFileSystem) CommitTransaction() {
	f.Lock()
	defer f.Unlock()
	if f.tx != nil {
		f.tx.Commit()
		f.tx = nil
	}
}

func (f *DBFileSystem) EndTransaction() {
	if err := recover(); err != nil {
		f.RollbackTransaction()
	} else {
		f.CommitTransaction()
	}
}

func (f *DBFileSystem) MakeDir(path string, _ os.FileMode) (err error) {
	err = f.MakeDirAll(path, 0)
	return
}

func (f *DBFileSystem) MakeDirAll(path string, _ os.FileMode) (err error) {
	//f.Lock()
	//defer f.Unlock()
	var entry *File
	if entry, err = f.getEntryUnsafe(path); err == nil {
		if entry.Mime == InodeDirectoryMimeType {
			// err = fmt.Errorf("directory exists already")
			return
		}
		err = fmt.Errorf("path is a File")
	} else {
		tx := f.tableScopedOrTx()
		realpath := f.realpath(path)
		parents := clPath.ParseParentPaths(realpath)
		for _, parent := range parents {
			entry = &File{
				Path:    parent,
				Mime:    InodeDirectoryMimeType,
				Shasum:  "",
				Content: []byte{},
				Context: []byte{},
			}
			if _, ee := f.getEntryUnsafe(parent); ee != nil {
				tx.Save(entry)
			}
		}
	}
	return
}

func (f *DBFileSystem) Remove(path string) (err error) {
	//f.Lock()
	//defer f.Unlock()
	realpath := f.realpath(path)
	if tx := f.tableScopedOrTx().Where(`path = ?`, realpath); tx.Error != nil {
		err = tx.Error
		return
	} else if tx = tx.Unscoped().Delete(&File{}); tx.Error != nil {
		err = tx.Error
		return
	}
	return
}

func (f *DBFileSystem) RemoveAll(path string) (err error) {
	//f.Lock()
	//defer f.Unlock()
	realpath := f.realpath(path)
	if err = f.Remove(path); err != nil {
		return
	} else if tx := f.tableScopedOrTx().Where(`path LIKE ?`, realpath+"/%"); tx.Error != nil {
		err = tx.Error
		return
	} else if tx = tx.Unscoped().Delete(&File{}); tx.Error != nil {
		err = tx.Error
		return
	}
	return
}

func (f *DBFileSystem) WriteFile(path string, data []byte, _ os.FileMode) (err error) {
	//f.Lock()
	//defer f.Unlock()
	if dirPath := clPath.Dir(path); dirPath != "" {
		_ = f.MakeDirAll(dirPath, 0)
	}
	realpath := f.realpath(path)
	var shasum string
	if shasum, err = sha.Sum256(data); err != nil {
		return
	}
	mime := mimetype.Detect(data).String()
	var entry *File
	if entry, err = f.getEntryUnsafe(path); err != nil {
		entry = &File{
			Path:    realpath,
			Mime:    mime,
			Shasum:  shasum,
			Content: data,
			// Context: ctx,
		}
		err = f.tableScopedOrTx().Save(entry).Error
	} else {
		entry.Path = realpath
		entry.Mime = mime
		entry.Shasum = shasum
		entry.Content = data
		// entry.Context = ctx
		err = f.tableScopedOrTx().Where(`path = ?`, realpath).Updates(entry).Error
	}

	return
}

func (f *DBFileSystem) ChangeTimes(path string, created, updated time.Time) (err error) {
	//f.Lock()
	//defer f.Unlock()
	realpath := f.realpath(path)
	err = f.tableScopedOrTx().Where(`path = ?`, realpath).Updates(map[string]interface{}{
		"created": created,
		"updated": updated,
	}).Error
	return
}

func (f *DBFileSystem) WritePageMatter(pm *matter.PageMatter) (err error) {
	//f.Lock()
	//defer f.Unlock()
	if dirPath := clPath.Dir(pm.Path); dirPath != "" {
		_ = f.MakeDirAll(dirPath, 0)
	}
	realpath := f.realpath(pm.Path)

	var data []byte
	if data, err = pm.Bytes(); err != nil {
		err = fmt.Errorf("error getting bytes from page matter: %v", err)
		return
	}

	var shasum string
	if shasum, err = sha.Sum256(data); err != nil {
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

	err = f.tableScopedOrTx().Save(entry).Error
	return
}

func (f *DBFileSystem) RemovePageMatter(path string) (err error) {
	err = f.Remove(path)
	return
}
