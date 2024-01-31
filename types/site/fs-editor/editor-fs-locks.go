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

package fs_editor

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-corelibs/path"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CEditorFeature[MakeTypedFeature]) LockEditorFile(eid, fsid, filePath string) (err error) {
	if !userbase.IsValidEID(eid) {
		err = fmt.Errorf("invalid EID value")
		return
	}
	checkPath := path.TrimSlashes(filePath)
	lockPath := checkPath + ".~lock"
	for _, efs := range f.EditingFileSystems {
		if efs.Tag().String() == fsid {
			var hasRWFS, exists bool
			for _, mountedPoints := range efs.GetMountedPoints() {
				for _, mountPoint := range mountedPoints {
					if !strings.HasPrefix("/"+lockPath, mountPoint.Mount) {
						continue
					} else if hasRWFS = mountPoint.RWFS != nil; hasRWFS {
						if exists = mountPoint.RWFS.Exists(checkPath); exists {
							err = mountPoint.RWFS.WriteFile(lockPath, []byte(eid), 0664)
							return
						}
					}
				}
			}
			if !hasRWFS {
				err = fmt.Errorf("%s filesystem is read-only", fsid)
				return
			}
			err = os.ErrNotExist
			return
		}
	}
	err = os.ErrNotExist
	return
}

func (f *CEditorFeature[MakeTypedFeature]) IsEditorFileLocked(fsid, filePath string) (eid string, locked bool) {
	checkPath := path.TrimSlashes(filePath)
	lockPath := checkPath + ".~lock"
	for _, efs := range f.EditingFileSystems {
		if efs.Tag().String() == fsid {
			for _, mountedPoints := range efs.GetMountedPoints() {
				for _, mountPoint := range mountedPoints {
					if !strings.HasPrefix("/"+lockPath, mountPoint.Mount) {
						continue
					} else if mountPoint.RWFS != nil {
						if mountPoint.RWFS.Exists(lockPath) {
							if data, err := mountPoint.RWFS.ReadFile(lockPath); err == nil {
								if value := string(data); userbase.IsValidEID(value) {
									eid = value
									locked = true
									return
								}
							}
							return
						}
					}
				}
			}
			return
		}
	}
	return
}

func (f *CEditorFeature[MakeTypedFeature]) UnLockEditorFile(fsid, filePath string) (err error) {
	if _, locked := f.IsEditorFileLocked(fsid, filePath); !locked {
		return
	}
	checkPath := path.TrimSlashes(filePath)
	lockPath := checkPath + ".~lock"
	for _, efs := range f.EditingFileSystems {
		if efs.Tag().String() == fsid {
			for _, mountedPoints := range efs.GetMountedPoints() {
				for _, mountPoint := range mountedPoints {
					if !strings.HasPrefix("/"+lockPath, mountPoint.Mount) {
						continue
					} else if mountPoint.RWFS != nil && mountPoint.RWFS.Exists(lockPath) {
						err = mountPoint.RWFS.Remove(lockPath)
						return
					}
				}
			}
			return
		}
	}
	return
}
