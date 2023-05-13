//go:build driver_fs_db_gorm || drivers_fs_db || drivers_fs || drivers || all

// Copyright (c) 2022  The Go-Enjin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this File except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package embed

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/fvbommel/sortorder"
	times "github.com/go-enjin/github-com-djherbis-times"
	"gorm.io/gorm"

	"github.com/go-enjin/be/pkg/globals"
	bePath "github.com/go-enjin/be/pkg/path"
)

type CFileSystem struct {
	path string
	wrap string

	table string
	db    *gorm.DB
}

func New(path, table string, db *gorm.DB) (f CFileSystem, err error) {
	if db == nil {
		err = fmt.Errorf("db arugment can not be nil")
		return
	}
	switch table {
	case "", "-", "nil":
		table = "be_filesystem"
	}
	f = CFileSystem{
		path:  path,
		wrap:  "",
		table: table,
		db:    db,
	}
	err = f.tx().AutoMigrate(&File{})
	return
}

func (f CFileSystem) tx() (tx *gorm.DB) {
	tx = f.db.Scopes(func(tx *gorm.DB) *gorm.DB {
		if f.table != "" {
			return tx.Table(f.table)
		}
		return tx
	})
	return
}

func (f CFileSystem) Name() (name string) {
	name = f.path
	return
}

func (f CFileSystem) realpath(path string) (out string) {
	out = bePath.SafeConcatRelPath(f.path, path)
	return
}

func (f CFileSystem) pruneEntries(paths []string) (pruned []string) {
	rp := f.path
	for _, entry := range paths {
		if strings.HasPrefix(entry, rp) {
			if entry = entry[len(rp):]; entry != "" && entry[0] == '/' {
				entry = entry[1:]
			}
		}
		pruned = append(pruned, entry)
	}
	return
}

func (f CFileSystem) getEntry(path string) (entry *File, err error) {
	realpath := f.realpath(path)
	entry = &File{}
	if err = f.tx().Where(`path = ?`, realpath).First(entry).Error; err != nil {
		entry = nil
	}
	return
}

func (f CFileSystem) getStub(path string) (stub *entryStub, err error) {
	realpath := f.realpath(path)
	stub = &entryStub{}
	if err = f.tx().Where(`path = ?`, realpath).First(stub).Error; err != nil {
		stub = nil
	}
	return
}

func (f CFileSystem) getStamp(path string) (stamp *entryStamp, err error) {
	realpath := f.realpath(path)
	stamp = &entryStamp{}
	if err = f.tx().Where(`path = ?`, realpath).First(stamp).Error; err != nil {
		stamp = nil
	}
	return
}

func (f CFileSystem) Exists(path string) (exists bool) {
	if entry, err := f.getEntry(path); err == nil {
		exists = entry != nil && entry.ID > 0
	}
	return
}

func (f CFileSystem) Open(path string) (file fs.File, err error) {
	var entry *File
	if entry, err = f.getEntry(path); err != nil {
		return
	} else if entry == nil {
		err = fs.ErrNotExist
		return
	}
	if entry.IsDir() {
		err = fmt.Errorf("not a file")
		return
	}
	file = entry
	return
}

func (f CFileSystem) ListDirs(path string) (paths []string, err error) {
	realpath := f.realpath(path)
	var stubs []entryStub
	if err = f.tx().Where(`path LIKE ?`, sqlEscapeLIKE(realpath)+"/%").Find(stubs).Error; err != nil {
		return
	}
	for _, stub := range stubs {
		if stub.Mime == InodeDirectoryMimeType {
			// not sub-directories
			if isDirectChild(realpath, stub.Path) {
				paths = append(paths, stub.Path)
			}
		}
	}
	sort.Sort(sortorder.Natural(paths))
	return
}

func (f CFileSystem) ListFiles(path string) (paths []string, err error) {
	realpath := f.realpath(path)
	var stubs []entryStub
	if err = f.tx().Where(`path LIKE ?`, sqlEscapeLIKE(realpath)+"/%").Find(stubs).Error; err != nil {
		return
	}
	for _, stub := range stubs {
		if stub.Mime != InodeDirectoryMimeType {
			if isDirectChild(realpath, stub.Path) {
				paths = append(paths, stub.Path)
			}
		}
	}
	sort.Sort(sortorder.Natural(paths))
	return
}

func (f CFileSystem) ListAllDirs(path string) (paths []string, err error) {
	realpath := f.realpath(path)
	var stubs []entryStub
	if err = f.tx().Where(`path LIKE ?`, sqlEscapeLIKE(realpath)+"/%").Find(stubs).Error; err != nil {
		return
	}
	for _, stub := range stubs {
		if stub.Mime == InodeDirectoryMimeType {
			paths = append(paths, stub.Path)
		}
	}
	sort.Sort(sortorder.Natural(paths))
	return
}

func (f CFileSystem) ListAllFiles(path string) (paths []string, err error) {
	realpath := f.realpath(path)
	var stubs []entryStub
	if err = f.tx().Where(`path LIKE ?`, sqlEscapeLIKE(realpath)+"/%").Find(stubs).Error; err != nil {
		return
	}
	for _, stub := range stubs {
		if stub.Mime != "" && stub.Mime != InodeDirectoryMimeType {
			paths = append(paths, stub.Path)
		}
	}
	sort.Sort(sortorder.Natural(paths))
	return
}

func (f CFileSystem) ReadDir(path string) (paths []fs.DirEntry, err error) {
	realpath := f.realpath(path)
	var entries []*File
	if err = f.tx().Where(`path LIKE ?`, sqlEscapeLIKE(realpath)+"/%").Find(entries).Error; err != nil {
		return
	}
	for _, entry := range entries {
		if entry.Mime != InodeDirectoryMimeType {
			if isDirectChild(realpath, entry.Path) {
				paths = append(paths, entry)
			}
		}
	}
	sort.Slice(entries, func(i, j int) (less bool) {
		a := entries[i]
		b := entries[j]
		less = sortorder.NaturalLess(a.Path, b.Path)
		return
	})
	return
}

func (f CFileSystem) ReadFile(path string) (content []byte, err error) {
	var entry *File
	if entry, err = f.getEntry(path); err != nil {
		return
	} else if entry != nil && entry.ID > 0 {
		content = entry.Content
	} else {
		err = fs.ErrNotExist
	}
	return
}

func (f CFileSystem) MimeType(path string) (mime string, err error) {
	var stub *entryStub
	if stub, err = f.getStub(path); err != nil {
		return
	} else if stub.Mime != "" && stub.Mime != InodeDirectoryMimeType {
		mime = stub.Mime
	} else {
		err = fs.ErrNotExist
	}
	return
}

func (f CFileSystem) Shasum(path string) (shasum string, err error) {
	var stub *entryStub
	if stub, err = f.getStub(path); err != nil {
		return
	} else if stub.Mime != "" && stub.Mime != InodeDirectoryMimeType {
		shasum = stub.Shasum
	} else {
		err = fs.ErrNotExist
	}
	return
}

func (f CFileSystem) FileCreated(path string) (created int64, err error) {
	var stamp *entryStamp
	if stamp, err = f.getStamp(path); err != nil {
		return
	} else if stamp != nil {
		created = stamp.CreatedAt.Unix()
	} else {
		err = fs.ErrNotExist
	}
	return
}

func (f CFileSystem) LastModified(path string) (modTime int64, err error) {
	var stamp *entryStamp
	if stamp, err = f.getStamp(path); err != nil {
		return
	} else if stamp != nil {
		modTime = stamp.UpdatedAt.Unix()
	} else {
		err = fs.ErrNotExist
	}
	return
}

func (f CFileSystem) FileStats(path string) (mime, shasum string, created, updated time.Time, err error) {
	if mime, err = f.MimeType(path); err != nil {
		return
	}
	if shasum, err = f.Shasum(path); err != nil {
		return
	}
	var ts times.Timespec
	if ts, err = globals.BuildFileInfo(); err != nil {
		return
	}
	updated = ts.ModTime()
	if ts.HasBirthTime() {
		created = ts.BirthTime()
	}
	return
}

// func (f CFileSystem) Mkdir(path string) (err error) {
// 	var entry *File
// 	if entry, err = f.getEntry(path); err == nil {
// 		if entry.Mime == InodeDirectoryMimeType {
// 			// err = fmt.Errorf("directory exists already")
// 			return
// 		}
// 		err = fmt.Errorf("path is a File")
// 	} else {
// 		entry = &File{
// 			Path:    path,
// 			Mime:    InodeDirectoryMimeType,
// 			Shasum:  "",
// 			Content: []byte{},
// 			Context: []byte{},
// 		}
// 		err = f.tx().Save(entry).Error
// 	}
// 	return
// }
//
// func (f CFileSystem) Rmdir(path string) (err error) {
// 	// must return an error if there are any children
// 	return
// }
//
// func (f CFileSystem) RmdirAll(path string) (err error) {
// 	// delete all things with the prefix of `#{path}/%`
// 	return
// }
//
// func (f CFileSystem) WriteFile(path, mime string, content []byte, ctx []byte) (err error) {
// 	var shasum string
// 	if shasum, err = sha.DataHash64(content); err != nil {
// 		return
// 	}
// 	var entry *File
// 	if entry, err = f.getEntry(path); err != nil {
// 		err = nil
// 		entry = &File{
// 			Path:    path,
// 			Mime:    mime,
// 			Shasum:  shasum,
// 			Content: content,
// 			Context: ctx,
// 		}
// 	} else {
// 		entry.Path = path
// 		entry.Mime = mime
// 		entry.Shasum = shasum
// 		entry.Content = content
// 		entry.Context = ctx
// 	}
// 	err = f.tx().Save(entry).Error
// 	return
// }
//
// func (f CFileSystem) DeleteFile(path string) (err error) {
// 	realpath := f.realpath(path)
// 	err = f.tx().Where(`path = ?`, realpath).Delete(&File{}).Error
// 	return
// }