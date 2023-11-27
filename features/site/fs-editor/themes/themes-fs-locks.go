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

package themes

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) LockEditorFile(eid, fsid, filePath string) (err error) {
	if !userbase.IsValidEID(eid) {
		err = fmt.Errorf("invalid EID value")
		return
	}
	var ok bool
	var code string
	if code, filePath, ok = strings.Cut(filePath, "/"); !ok {
		err = fmt.Errorf("not a file path")
		return
	}
	checkPath := path.TrimSlashes(filePath)
	lockPath := checkPath + ".~lock"
	for _, efs := range f.EditingFileSystems {
		if efs.Tag().String() == fsid {
			var hasRWFS, exists bool
			for _, mps := range efs.GetMountedPoints() {
				for _, mp := range mps {
					if mount := path.TrimSlashes(mp.Mount); code != mount {
						continue
					} else if hasRWFS = mp.RWFS != nil; hasRWFS {
						if exists = mp.RWFS.Exists(checkPath); exists {
							err = mp.RWFS.WriteFile(lockPath, []byte(eid), 0664)
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

func (f *CFeature) IsEditorFileLocked(fsid, filePath string) (eid string, locked bool) {
	var ok bool
	var code string
	if code, filePath, ok = strings.Cut(filePath, "/"); !ok {
		log.ErrorF("not a file path")
		return
	}
	checkPath := path.TrimSlashes(filePath)
	lockPath := checkPath + ".~lock"
	for _, efs := range f.EditingFileSystems {
		if efs.Tag().String() == fsid {
			for _, mps := range efs.GetMountedPoints() {
				for _, mp := range mps {
					if mount := path.TrimSlashes(mp.Mount); code != mount {
						continue
					} else if mp.RWFS != nil {
						if mp.RWFS.Exists(lockPath) {
							if data, err := mp.RWFS.ReadFile(lockPath); err == nil {
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

func (f *CFeature) UnLockEditorFile(fsid, filePath string) (err error) {
	if _, locked := f.IsEditorFileLocked(fsid, filePath); !locked {
		return
	}
	var ok bool
	var code string
	if code, filePath, ok = strings.Cut(filePath, "/"); !ok {
		log.ErrorF("not a file path")
		return
	}
	checkPath := path.TrimSlashes(filePath)
	lockPath := checkPath + ".~lock"
	for _, efs := range f.EditingFileSystems {
		if efs.Tag().String() == fsid {
			for _, mps := range efs.GetMountedPoints() {
				for _, mp := range mps {
					if mount := path.TrimSlashes(mp.Mount); code != mount {
						continue
					} else if mp.RWFS != nil && mp.RWFS.Exists(lockPath) {
						err = mp.RWFS.Remove(lockPath)
						return
					}
				}
			}
			return
		}
	}
	return
}
