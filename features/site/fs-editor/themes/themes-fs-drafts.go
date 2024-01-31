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
	"strings"

	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-corelibs/path"
	"github.com/go-enjin/be/types/page/matter"
)

func (f *CFeature) DraftExists(info *editor.File) (present bool) {
	checkPath := path.TrimSlashes(info.EditPath())
	draftPath := checkPath + ".~draft"
	for _, efs := range f.EditingFileSystems {
		if efs.Tag().String() == info.FSID {
			for _, mountedPoints := range efs.GetMountedPoints() {
				for _, mountPoint := range mountedPoints {
					if mount := path.TrimSlashes(mountPoint.Mount); mount != info.Code {
						continue
					} else if present = mountPoint.ROFS.Exists(draftPath); present {
						return
					}
				}
			}
			return
		}
	}
	return
}

func (f *CFeature) ReadDraft(info *editor.File) (contents []byte, err error) {
	checkPath := path.TrimSlashes(info.EditPath())
	var draftPath string
	if strings.HasSuffix(checkPath, ".~draft") {
		draftPath = checkPath
	} else {
		draftPath = checkPath + ".~draft"
	}
	for _, efs := range f.EditingFileSystems {
		if efs.Tag().String() == info.FSID {
			for _, mountedPoints := range efs.GetMountedPoints() {
				for _, mountPoint := range mountedPoints {
					if mount := path.TrimSlashes(mountPoint.Mount); mount != info.Code {
						continue
					} else if mountPoint.ROFS != nil {
						var data []byte
						if data, err = mountPoint.ROFS.ReadFile(draftPath); err == nil {
							contents = data
						}
						return
					}
				}
			}
			err = fmt.Errorf("read/write mount point not found")
			return
		}
	}
	err = fmt.Errorf("fileystem not found")
	return
}

func (f *CFeature) ReadDraftMatter(info *editor.File) (pm *matter.PageMatter, err error) {
	checkPath := path.TrimSlashes(info.EditPath())
	var draftPath string
	if strings.HasSuffix(checkPath, ".~draft") {
		draftPath = checkPath
	} else {
		draftPath = checkPath + ".~draft"
	}
	for _, efs := range f.EditingFileSystems {
		if efs.Tag().String() == info.FSID {
			for _, mountedPoints := range efs.GetMountedPoints() {
				for _, mountPoint := range mountedPoints {
					if mount := path.TrimSlashes(mountPoint.Mount); mount != info.Code {
						continue
					} else if mountPoint.ROFS != nil {
						pm, err = mountPoint.ROFS.ReadPageMatter(draftPath)
						return
					}
				}
			}
			err = fmt.Errorf("read/write mount point not found")
			return
		}
	}
	err = fmt.Errorf("fileystem not found")
	return
}

func (f *CFeature) WriteDraft(info *editor.File, contents []byte) (err error) {
	checkPath := path.TrimSlashes(info.EditPath())
	draftPath := checkPath + ".~draft"
	for _, efs := range f.EditingFileSystems {
		if efs.Tag().String() == info.FSID {
			for _, mountedPoints := range efs.GetMountedPoints() {
				for _, mountPoint := range mountedPoints {
					if mount := path.TrimSlashes(mountPoint.Mount); mount != info.Code {
						continue
					} else if mountPoint.RWFS != nil {
						err = mountPoint.RWFS.WriteFile(draftPath, contents, 0664)
						return
					}
				}
			}
			err = fmt.Errorf("read/write mount point not found")
			return
		}
	}
	err = fmt.Errorf("fileystem not found")
	return
}

func (f *CFeature) RemoveDraft(info *editor.File) (err error) {
	if f.SelfEditor().DraftExists(info) {
		checkPath := path.TrimSlashes(info.EditPath())
		draftPath := checkPath + ".~draft"
		for _, efs := range f.EditingFileSystems {
			if efs.Tag().String() == info.FSID {
				for _, mountedPoints := range efs.GetMountedPoints() {
					for _, mountPoint := range mountedPoints {
						if mount := path.TrimSlashes(mountPoint.Mount); mount != info.Code {
							continue
						} else if mountPoint.RWFS != nil {
							if present := mountPoint.RWFS.Exists(draftPath); present {
								err = mountPoint.RWFS.Remove(draftPath)
								return
							}
						}
					}
				}
				return
			}
		}
	}
	return
}

func (f *CFeature) PublishDraft(info *editor.File) (err error) {
	if f.SelfEditor().DraftExists(info) {
		checkPath := path.TrimSlashes(info.EditPath())
		draftPath := checkPath + ".~draft"
		for _, efs := range f.EditingFileSystems {
			if efs.Tag().String() == info.FSID {
				for _, mountedPoints := range efs.GetMountedPoints() {
					for _, mountPoint := range mountedPoints {
						if mount := path.TrimSlashes(mountPoint.Mount); mount != info.Code {
							continue
						} else if mountPoint.RWFS != nil {
							if present := mountPoint.RWFS.Exists(draftPath); present {
								var data []byte
								if data, err = mountPoint.RWFS.ReadFile(draftPath); err != nil {
									return
								}
								if err = mountPoint.RWFS.WriteFile(checkPath, data, 0664); err != nil {
									return
								}
								if err = mountPoint.RWFS.Remove(draftPath); err != nil {
									return
								}
								return
							}
						}
					}
				}
				return
			}
		}
	}
	// no draft to publish, nop
	return
}
