//go:build driver_fs_db_gorm || drivers_fs_db || drivers_fs || drivers || all

// Copyright (c) 2023  The Go-Enjin Authors
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
	"io"
	"io/fs"
	"path/filepath"
	"sync"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	_ fs.File     = (*File)(nil)
	_ fs.FileInfo = (*File)(nil)
)

const InodeDirectoryMimeType = "inode/directory"

type entryStub struct {
	Path   string
	Mime   string
	Shasum string
}

type entryStamp struct {
	CreatedAt time.Time      `gorm:"column:created,index"`
	UpdatedAt time.Time      `gorm:"column:updated,index"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted,index"`
}

type File struct {
	ID uint `gorm:"primaryKey"`

	// Path is the full URL path to this filesystem entry, including File name
	Path string `gorm:"not null,unique,uniqueIndex"`

	// Mime is the mimetype of this filesystem entry
	Mime string `gorm:"not null"`

	// Shasum is the SHA-256 sum of the content
	Shasum string

	// Content is the binary data of the File
	Content []byte

	// Context is arbitrary metadata associated with this filesystem entry
	Context datatypes.JSON `gorm:"index,default:{}"`

	CreatedAt time.Time      `gorm:"column:created,index"`
	UpdatedAt time.Time      `gorm:"column:updated,index"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted,index"`

	readIndex int
	closed    bool
	rwlock    sync.RWMutex
}

func (e *File) clone() (cloned *File, err error) {
	e.rwlock.RLock()
	defer e.rwlock.RUnlock()
	if e.closed {
		err = fs.ErrClosed
		return
	}
	cloned = &File{
		ID:        e.ID,
		Path:      e.Path,
		Mime:      e.Mime,
		Shasum:    e.Shasum,
		Content:   e.Content,
		Context:   e.Context,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		DeletedAt: e.DeletedAt,
		readIndex: 0,
		closed:    false,
		rwlock:    sync.RWMutex{},
	}
	return
}

func (e *File) Stat() (info fs.FileInfo, err error) {
	info, err = e.clone()
	return
}

func (e *File) Read(v []byte) (read int, err error) {
	e.rwlock.Lock()
	defer e.rwlock.Unlock()
	if e.closed {
		err = fs.ErrClosed
		return
	}
	capacity := cap(v)
	contentLen := len(e.Content)
	if e.readIndex < contentLen {
		for idx, b := range e.Content[e.readIndex:] {
			if idx >= capacity {
				break
			}
			v[idx] = b
			read += 1
		}
		e.readIndex += read
		return
	}
	err = io.EOF
	return
}

func (e *File) Reset() (err error) {
	e.rwlock.Lock()
	defer e.rwlock.Unlock()
	if e.closed {
		err = fs.ErrClosed
		return
	}
	e.readIndex = 0
	return
}

func (e *File) Close() (err error) {
	e.rwlock.Lock()
	defer e.rwlock.Unlock()
	if e.closed {
		err = fs.ErrClosed
		return
	}
	e.closed = true
	return
}

func (e *File) Name() (name string) {
	e.rwlock.RLock()
	defer e.rwlock.RUnlock()
	if e.closed {
		return
	}
	name = filepath.Base(e.Path)
	return
}

func (e *File) Size() (size int64) {
	if e.IsDir() {
		size = 0
		return
	}
	e.rwlock.RLock()
	defer e.rwlock.RUnlock()
	if e.closed {
		return
	}
	size = int64(len(e.Content))
	return
}

func (e *File) Mode() (mode fs.FileMode) {
	e.rwlock.RLock()
	defer e.rwlock.RUnlock()
	if e.closed {
		return
	}
	if e.IsDir() {
		mode = 0700
	} else {
		mode = 0600
	}
	return
}

func (e *File) ModTime() (modTime time.Time) {
	e.rwlock.RLock()
	defer e.rwlock.RUnlock()
	if e.closed {
		return
	}
	modTime = e.UpdatedAt
	return
}

func (e *File) IsDir() (isDir bool) {
	e.rwlock.RLock()
	defer e.rwlock.RUnlock()
	if e.closed {
		return
	}
	isDir = e.Mime == InodeDirectoryMimeType
	return
}

func (e *File) Sys() (sys interface{}) {
	return
}

func (e *File) Type() (mode fs.FileMode) {
	mode = e.Mode()
	return
}

func (e *File) Info() (info fs.FileInfo, err error) {
	info, err = e.Stat()
	return
}