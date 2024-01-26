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

package feature

import (
	"errors"
	"os"
	"sort"
	"strings"

	"github.com/maruel/natural"

	"github.com/go-enjin/be/pkg/fs"
	bePath "github.com/go-enjin/be/pkg/path"
)

type CMountPoint struct {
	// Path is the actual filesystem path
	Path string
	// Mount is the URL path prefix
	Mount string
	// ROFS is the read-only filesystem, always non-nil
	ROFS fs.FileSystem
	// RWFS is the write-only filesystem, nil when fs is read-only
	RWFS fs.RWFileSystem
}

type MountPoints []*CMountPoint

func (m MountPoints) Len() (count int) {
	count = len(m)
	return
}

func (m MountPoints) Append(mountPoint *CMountPoint) (modified MountPoints) {
	var found bool
	for _, mp := range m {
		found = mp.Mount == mountPoint.Mount && mp.Path == mountPoint.Path
		modified = append(modified, mp)
	}
	if !found {
		modified = append(modified, mountPoint)
	}
	return
}

func (m MountPoints) HasRWFS() (rw bool) {
	for _, mp := range m {
		if rw = mp.RWFS != nil; rw {
			return
		}
	}
	return
}

type MountedPoints map[string]MountPoints

func (m MountedPoints) HasRWFS() (rw bool) {
	for _, mps := range m {
		if rw = mps.HasRWFS(); !rw {
			return
		}
	}
	return
}

func (m MountedPoints) ListMounts() (mounts []string) {
	for point := range m {
		mounts = append(mounts, point)
	}
	// root points must go last
	sort.Slice(mounts, func(i, j int) (less bool) {
		a, b := mounts[i], mounts[j]
		aRoot, bRoot := a == "/", b == "/"
		if aRoot && bRoot {
			return false
		} else if aRoot || bRoot {
			return aRoot && !bRoot
		}
		aLen, bLen := len(a), len(b)
		if less = aLen < bLen; less {
			return
		} else if aLen > bLen {
			return
		}
		// equal lengths, sorted naturally
		less = natural.Less(a, b)
		return
	})
	return
}

func (m MountedPoints) HasRootOrAllOf(mounts ...string) (present bool) {
	unique := make(map[string]struct{})
	for _, mount := range m.ListMounts() {
		if present = mount == "/"; present {
			return
		}
		unique[mount] = struct{}{}
	}
	for _, point := range mounts {
		if _, present = unique[point]; !present {
			return
		}
	}
	return
}

// FindPathPoints returns a list of MountPoints which prefix match the given path
func (m MountedPoints) FindPathPoints(path string) (mountPoints MountPoints) {
	cleaned := bePath.CleanWithSlash(path)
	var roots MountPoints
	for _, mount := range m.ListMounts() {
		if mount == "/" {
			roots = append(roots, m[mount]...)
			continue
		}
		if mount == cleaned || strings.HasPrefix(cleaned, mount+"/") {
			mountPoints = append(mountPoints, m[mount]...)
		}
	}
	if mountPoints.Len() == 0 && roots.Len() > 0 {
		mountPoints = roots
	}
	return
}

// FindRWPathPoint finds the first read-write MountPoint matching the path given
func (m MountedPoints) FindRWPathPoint(path string) (readWrite *CMountPoint) {
	var mountPoints MountPoints
	cleaned := bePath.CleanWithSlash(path)
	var roots MountPoints
	for _, mount := range m.ListMounts() {
		if mount == "/" {
			roots = append(roots, m[mount]...)
			continue
		}
		if mount == cleaned || strings.HasPrefix(cleaned, mount+"/") {
			mountPoints = append(mountPoints, m[mount]...)
		}
	}
	if mountPoints.Len() == 0 && roots.Len() > 0 {
		mountPoints = roots
	}
	for _, mp := range mountPoints {
		if mp.RWFS != nil {
			readWrite = mp
			return
		}
	}
	return
}

func (m MountedPoints) Exists(path string) (present bool) {
	for _, mp := range m.FindPathPoints(path) {
		if present = mp.ROFS.Exists(path); present {
			return
		}
	}
	return
}

func (m MountedPoints) IsReadOnly(path string) (readOnly bool) {
	readOnly = m.FindRWPathPoint(path) == nil
	return
}

func (m MountedPoints) ReadFile(path string) (data []byte, err error) {
	for _, mp := range m.FindPathPoints(path) {
		if mp.ROFS.Exists(path) {
			data, err = mp.ROFS.ReadFile(path)
			return
		}
	}
	err = os.ErrNotExist
	return
}

func (m MountedPoints) WriteFile(path string, data []byte) (err error) {
	if mp := m.FindRWPathPoint(path); mp != nil {
		err = mp.RWFS.WriteFile(path, data, 0660)
		return
	}
	err = errors.New("read-only filesystem")
	return
}

func (m MountedPoints) RemoveFile(path string) (err error) {
	if mp := m.FindRWPathPoint(path); mp != nil {
		if mp.RWFS.Exists(path) {
			err = mp.RWFS.Remove(path)
		}
		return
	}
	err = errors.New("read-only filesystem")
	return
}

func (m MountedPoints) ListDirs(path string) (dirs []string) {
	unique := make(map[string]struct{})
	if mps := m.FindPathPoints(path); len(mps) > 0 {
		for _, mp := range mps {
			if list, err := mp.ROFS.ListDirs(path); err == nil {
				for _, file := range list {
					if _, present := unique[file]; !present {
						unique[file] = struct{}{}
						dirs = append(dirs, file)
					}
				}
			}
		}
	}
	return
}

func (m MountedPoints) ListFiles(path string) (files []string) {
	unique := make(map[string]struct{})
	if mps := m.FindPathPoints(path); len(mps) > 0 {
		for _, mp := range mps {
			if list, err := mp.ROFS.ListFiles(path); err == nil {
				for _, file := range list {
					if _, present := unique[file]; !present {
						unique[file] = struct{}{}
						files = append(files, file)
					}
				}
			}
		}
	}
	return
}
