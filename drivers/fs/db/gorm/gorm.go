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

package gorm

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/maruel/natural"

	"gorm.io/gorm"

	times "github.com/go-enjin/github-com-djherbis-times"

	clStrings "github.com/go-corelibs/strings"
	beContext "github.com/go-enjin/be/pkg/context"
	beFs "github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/gob"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/types/page/matter"
)

func init() {
	gob.Register(DBFileSystem{})
}

type DBFileSystem struct {
	origin string

	path string
	wrap string

	name  string   // connection name
	table string   // table to use for all operations
	db    *gorm.DB // actual connection
	tx    *gorm.DB // current transaction

	id string

	sync.RWMutex
}

func New(origin, path, table, connection string, db *gorm.DB) (out *DBFileSystem, err error) {
	if db == nil {
		err = fmt.Errorf("db arugment can not be nil")
		return
	}
	switch table {
	case "", "-", "nil":
		table = "be_filesystem"
	}
	out = &DBFileSystem{
		origin: origin,
		path:   path,
		wrap:   "",
		name:   connection,
		table:  table,
		db:     db,
		tx:     nil,
		id:     fmt.Sprintf("%v+%v://%v@%v", origin, connection, table, path),
	}
	err = out.tableScopedOrTx().AutoMigrate(&File{})
	return
}

func (f *DBFileSystem) ID() (id string) {
	return f.id
}

func (f *DBFileSystem) CloneROFS() (cloned beFs.FileSystem) {
	return f.CloneRWFS()
}

func (f *DBFileSystem) CloneRWFS() (cloned beFs.RWFileSystem) {
	cloned = &DBFileSystem{
		origin: f.origin,
		path:   f.path,
		wrap:   f.wrap,
		name:   f.name,
		table:  f.table,
		db:     f.db,
		tx:     nil,
		id:     f.id,
	}
	return
}

func (f *DBFileSystem) EnjinName() (name string) {
	name = f.name
	return
}

func (f *DBFileSystem) Name() (name string) {
	name = f.path
	return
}

func (f *DBFileSystem) Exists(path string) (exists bool) {
	f.RLock()
	defer f.RUnlock()
	if entry, err := f.getEntryUnsafe(path); err == nil {
		exists = entry != nil && entry.ID > 0
	}
	return
}

func (f *DBFileSystem) Open(path string) (file fs.File, err error) {
	f.RLock()
	defer f.RUnlock()
	var entry *File
	if entry, err = f.getEntryUnsafe(path); err != nil {
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

func find[T interface{}](f *DBFileSystem, path string, dst T) (realpath string, err error) {
	realpath = f.realpath(path)
	if tx := f.tableScopedOrTx(); tx == nil {
		err = fmt.Errorf("transaction scope not found")
		return
	} else if stmt := tx.Where(`path LIKE ?`, sqlEscapeLIKE(realpath)+"/%"); stmt.Error != nil {
		err = stmt.Error
		return
	} else if err = stmt.Find(dst).Error; err != nil {
		return
	}
	return
}

func (f *DBFileSystem) ListDirs(path string) (paths []string, err error) {
	f.RLock()
	defer f.RUnlock()
	var stubs []entryStub
	var realpath string
	if realpath, err = find(f, path, &stubs); err != nil {
		return
	}

	for _, stub := range stubs {
		if stub.Mime == InodeDirectoryMimeType {
			// not sub-directories
			if isDirectChild(realpath, stub.Path) {
				paths = append(paths, beFs.PruneRootFrom(f.path, stub.Path))
			}
		}
	}
	sort.Sort(natural.StringSlice(paths))
	return
}

func (f *DBFileSystem) ListFiles(path string) (paths []string, err error) {
	f.RLock()
	defer f.RUnlock()
	var stubs []entryStub
	var realpath string
	if realpath, err = find(f, path, &stubs); err != nil {
		return
	}

	for _, stub := range stubs {
		if stub.Mime != InodeDirectoryMimeType {
			if isDirectChild(realpath, stub.Path) {
				paths = append(paths, beFs.PruneRootFrom(f.path, stub.Path))
			}
		}
	}
	sort.Sort(natural.StringSlice(paths))
	return
}

func (f *DBFileSystem) ListAllDirs(path string) (paths []string, err error) {
	f.RLock()
	defer f.RUnlock()
	var stubs []entryStub
	if _, err = find(f, path, &stubs); err != nil {
		return
	}

	for _, stub := range stubs {
		if stub.Mime == InodeDirectoryMimeType {
			paths = append(paths, beFs.PruneRootFrom(f.path, stub.Path))
		}
	}
	sort.Sort(natural.StringSlice(paths))
	return
}

func (f *DBFileSystem) ListAllFiles(path string) (paths []string, err error) {
	f.RLock()
	defer f.RUnlock()
	var stubs []entryStub
	if _, err = find(f, path, &stubs); err != nil {
		return
	}

	for _, stub := range stubs {
		if stub.Mime != "" && stub.Mime != InodeDirectoryMimeType {
			paths = append(paths, beFs.PruneRootFrom(f.path, stub.Path))
		}
	}
	sort.Sort(natural.StringSlice(paths))
	return
}

func (f *DBFileSystem) ReadDir(path string) (paths []fs.DirEntry, err error) {
	f.RLock()
	defer f.RUnlock()

	var entries []*File
	var realpath string
	if realpath, err = find(f, path, &entries); err != nil {
		return
	}

	for _, entry := range entries {
		//if entry.Mime != InodeDirectoryMimeType {
		if isDirectChild(realpath, entry.Path) {
			paths = append(paths, entry)
		}
		//}
	}
	sort.Slice(entries, func(i, j int) (less bool) {
		a := entries[i]
		b := entries[j]
		less = natural.Less(a.Path, b.Path)
		return
	})
	return
}

func (f *DBFileSystem) ReadFile(path string) (content []byte, err error) {
	f.RLock()
	defer f.RUnlock()
	var entry *File
	if entry, err = f.getEntryUnsafe(path); err != nil {
		return
	} else if entry != nil && entry.ID > 0 {
		content = entry.Content
	} else {
		err = fs.ErrNotExist
	}
	return
}

func (f *DBFileSystem) MimeType(path string) (mime string, err error) {
	f.RLock()
	defer f.RUnlock()
	var stub *entryStub
	if stub, err = f.getStubUnsafe(path); err != nil {
		return
	} else if stub.Mime != "" && stub.Mime != InodeDirectoryMimeType {
		mime = stub.Mime
	} else {
		err = fs.ErrNotExist
	}
	return
}

func (f *DBFileSystem) Shasum(path string) (shasum string, err error) {
	f.RLock()
	defer f.RUnlock()
	var stub *entryStub
	if stub, err = f.getStubUnsafe(path); err != nil {
		return
	} else if stub.Mime != "" && stub.Mime != InodeDirectoryMimeType {
		shasum = stub.Shasum
	} else {
		err = fs.ErrNotExist
	}
	return
}

func (f *DBFileSystem) Sha256(path string) (shasum string, err error) {
	f.RLock()
	defer f.RUnlock()
	var file *File
	if file, err = f.getEntryUnsafe(path); err != nil {
		return
	} else if file.Mime != "" && file.Mime != InodeDirectoryMimeType {
		shasum, err = sha.Hash256(file.Content)
	} else {
		err = fs.ErrNotExist
	}
	return
}

func (f *DBFileSystem) FileCreated(path string) (created int64, err error) {
	f.RLock()
	defer f.RUnlock()
	var stamp *entryStamp
	if stamp, err = f.getStampUnsafe(path); err != nil {
		return
	} else if stamp != nil {
		created = stamp.CreatedAt.Unix()
	} else {
		err = fs.ErrNotExist
	}
	return
}

func (f *DBFileSystem) LastModified(path string) (modTime int64, err error) {
	f.RLock()
	defer f.RUnlock()
	var stamp *entryStamp
	if stamp, err = f.getStampUnsafe(path); err != nil {
		return
	} else if stamp != nil {
		modTime = stamp.UpdatedAt.Unix()
	} else {
		err = fs.ErrNotExist
	}
	return
}

func (f *DBFileSystem) FileStats(path string) (mime, shasum string, created, updated time.Time, err error) {
	f.RLock()
	defer f.RUnlock()
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

func (f *DBFileSystem) FindFilePath(prefix string, extensions ...string) (path string, err error) {
	f.RLock()
	defer f.RUnlock()

	realpath := f.realpath(prefix)

	sort.Sort(clStrings.SortByLength(extensions))

	realpath = strings.TrimSuffix(realpath, "/")
	paths := []string{realpath}
	for _, extension := range extensions {
		paths = append(paths, realpath+"."+extension)
	}

	var entry File

	if tx := f.tableScopedOrTx(); tx == nil {
		err = fmt.Errorf("transaction scope not found")
		return
	} else if stmt := tx.Where(`path IN (?)`, paths); stmt.Error != nil {
		err = stmt.Error
		return
	} else if stmt = stmt.Order("path DESC"); stmt.Error != nil {
		err = stmt.Error
		return
	} else if err = stmt.First(&entry).Error; err != nil {
		return
	}

	path = beFs.PruneRootFrom(f.path, entry.Path)
	return
}

func (f *DBFileSystem) ReadPageMatter(path string) (pm *matter.PageMatter, err error) {
	f.RLock()

	var entry *File
	if entry, err = f.getEntryUnsafe(path); err != nil {
		f.RUnlock()
		return
	}

	var entryCtxData []byte
	if entryCtxData, err = entry.Context.MarshalJSON(); err != nil {
		err = fmt.Errorf("error marshalling json from gorm.File.Context: %v - %v", path, err)
		f.RUnlock()
		return
	}
	var entryCtx beContext.Context
	if err = json.Unmarshal(entryCtxData, &entryCtx); err != nil {
		err = fmt.Errorf("error unmarshalling json from gorm.File.Context data: %v - %v", path, err)
		f.RUnlock()
		return
	}

	fmType := matter.JsonMatter
	if v, ok := entryCtx.Get("FrontMatterType").(matter.FrontMatterType); ok {
		fmType = v
	}

	contents := matter.MakeStanza(fmType, entryCtx)
	contents += "\n"
	contents += string(entry.Content)

	f.RUnlock()

	_, _, created, updated, _ := f.FileStats(path)
	pm, err = matter.ParsePageMatter(f.origin, path, created, updated, []byte(contents))
	return

}
