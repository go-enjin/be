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
	"github.com/go-enjin/be/pkg/fs"
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