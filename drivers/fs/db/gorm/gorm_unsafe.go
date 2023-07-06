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
	"strings"

	"gorm.io/gorm"

	bePath "github.com/go-enjin/be/pkg/path"
)

func (f *DBFileSystem) realpath(path string) (out string) {
	out = bePath.SafeConcatRelPath(f.path, path)
	out = strings.TrimPrefix(out, "./")
	out = strings.TrimPrefix(out, "/")
	return
}

func (f *DBFileSystem) getEntryUnsafe(path string) (entry *File, err error) {
	realpath := f.realpath(path)
	entry = &File{}
	//sq := f.tableScopedOrTx().Begin()
	//defer func() {
	//	if sqErr := recover(); sqErr != nil {
	//		sq.Rollback()
	//	} else {
	//		sq.Commit()
	//	}
	//}()
	if err = f.tableScopedOrTx().Where(`path = ?`, realpath).First(entry).Error; err != nil {
		entry = nil
	}
	return
}

func (f *DBFileSystem) getStubUnsafe(path string) (stub *entryStub, err error) {
	realpath := f.realpath(path)
	stub = &entryStub{}
	//sq := f.tableScopedOrTx().Begin()
	//defer func() {
	//	if sqErr := recover(); sqErr != nil {
	//		sq.Rollback()
	//	} else {
	//		sq.Commit()
	//	}
	//}()
	if err = f.tableScopedOrTx().Where(`path = ?`, realpath).First(stub).Error; err != nil {
		stub = nil
	}
	return
}

func (f *DBFileSystem) getStampUnsafe(path string) (stamp *entryStamp, err error) {
	realpath := f.realpath(path)
	stamp = &entryStamp{}
	//sq := f.tableScopedOrTx().Begin()
	//defer func() {
	//	if sqErr := recover(); sqErr != nil {
	//		sq.Rollback()
	//	} else {
	//		sq.Commit()
	//	}
	//}()
	if err = f.tableScopedOrTx().Where(`path = ?`, realpath).First(stamp).Error; err != nil {
		stamp = nil
	}
	return
}

func (f *DBFileSystem) tableScopedOrTx() (tx *gorm.DB) {
	f.RLock()
	defer f.RUnlock()
	if f.tx != nil {
		return f.tx
	}
	return f.tableScoped()
	//var db *gorm.DB
	//if f.tx != nil {
	//	//return f.tx
	//	db = f.tx
	//} else {
	//	db = f.db
	//}
	//return db.Scopes(func(tx *gorm.DB) *gorm.DB {
	//	if f.table != "" {
	//		return tx.Table(f.table)
	//	}
	//	return tx
	//})
}

func (f *DBFileSystem) tableScoped() (tx *gorm.DB) {
	return f.db.Scopes(func(tx *gorm.DB) *gorm.DB {
		if f.table != "" {
			return tx.Table(f.table)
		}
		return tx
	})
}