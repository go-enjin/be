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
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/maruel/natural"

	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	beMime "github.com/go-enjin/be/pkg/mime"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/userbase"
	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/go-enjin/golang-org-x-text/message"
)

func (f *CFeature) FileExists(info *editor.File) (exists bool) {
	filePath := info.EditPath()
	for _, mpf := range f.EditingFileSystems {
		mpfTag := mpf.Tag().String()
		if mpfTag == info.FSID {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					if mount := bePath.TrimSlashes(mountPoint.Mount); mount != info.Code {
						continue
					} else if mountPoint.ROFS.Exists(filePath) {
						if mimeType, err := mountPoint.ROFS.MimeType(filePath); err == nil {
							if exists = mimeType != beMime.DirectoryMimeType; exists {
								return
							}
						}
					}
				}
			}
		}
	}
	return
}

func (f *CFeature) ReadFile(info *editor.File) (data []byte, err error) {
	filePath := info.EditPath()
	for _, mpf := range f.EditingFileSystems {
		mpfTag := mpf.Tag().String()
		if mpfTag == info.FSID {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					if mount := bePath.TrimSlashes(mountPoint.Mount); mount != info.Code {
						continue
					} else if mountPoint.ROFS.Exists(filePath) {
						if data, err = mountPoint.ROFS.ReadFile(filePath); err == nil {
							return
						}
					}
				}
			}
		}
	}
	err = os.ErrNotExist
	return
}

func (f *CFeature) WriteFile(info *editor.File, data []byte) (err error) {
	if info.ReadOnly {
		err = fmt.Errorf("file is read-only")
		return
	}
	filePath := info.EditPath()
	for _, mpf := range f.EditingFileSystems {
		mpfTag := mpf.Tag().String()
		if mpfTag == info.FSID {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					if mount := bePath.TrimSlashes(mountPoint.Mount); mount != info.Code {
						continue
					} else if mountPoint.RWFS != nil {
						err = mountPoint.RWFS.WriteFile(filePath, data, 0664)
						return
					}
				}
			}
		}
	}
	err = fmt.Errorf("filesystem is read-only")
	return
}

func (f *CFeature) RemoveFile(info *editor.File) (err error) {
	if info.ReadOnly {
		err = fmt.Errorf("file is read-only")
		return
	}
	filePath := info.EditPath()
	for _, mpf := range f.EditingFileSystems {
		mpfTag := mpf.Tag().String()
		if mpfTag == info.FSID {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					if mount := bePath.TrimSlashes(mountPoint.Mount); mount != info.Code {
						continue
					} else if mountPoint.ROFS.Exists(filePath) && mountPoint.RWFS != nil {
						err = mountPoint.RWFS.Remove(filePath)
						return
					}
				}
			}
		}
	}
	err = fmt.Errorf("filesystem is read-only")
	return
}

func (f *CFeature) RemoveDirectory(info *editor.File) (err error) {
	if info.ReadOnly {
		err = fmt.Errorf("directory is read-only")
		return
	}
	filePath := info.EditPath()
	for _, mpf := range f.EditingFileSystems {
		mpfTag := mpf.Tag().String()
		if mpfTag == info.FSID {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					if mount := bePath.TrimSlashes(mountPoint.Mount); mount != info.Code {
						continue
					} else if mountPoint.RWFS != nil && mountPoint.RWFS.Exists(filePath) {
						err = mountPoint.RWFS.Remove(filePath)
						return
					}
				}
			}
		}
	}
	err = fmt.Errorf("filesystem is read-only")
	return
}

func (f *CFeature) PrepareEditableFile(r *http.Request, info *editor.File) (editFile *editor.File) {
	eid := userbase.GetCurrentUserEID(r)
	printer := lang.GetPrinterFromRequest(r)

	for _, mpf := range f.EditingFileSystems {
		mpfTag := mpf.Tag().String()

		if info.FSID != "" && info.FSID != mpfTag {
			continue
		}

		mountedPoints := mpf.GetMountedPoints()
		for _, point := range maps.SortedKeys(mountedPoints) {
			for _, mountPoint := range mountedPoints[point] {
				if mount := bePath.TrimSlashes(mountPoint.Mount); mount != info.Code {
					continue
				} else if mountPoint.ROFS.Exists(info.EditPath()) {
					if ef, ignored := f.SelfEditor().ProcessMountPointFile(r, printer, eid, mpf.BaseTag().String(), mpfTag, info.Code, info.Path, info.EditPath(), mountPoint, true); ignored {
						continue
					} else {
						editFile = ef
						editFile.FSBT = mpf.BaseTag().String()
						editFile.Tilde = info.Tilde
						return
					}
				}
			}
		}
	}
	return
}

func (f *CFeature) ListFileSystems() (list editor.Files) {
	for _, mpf := range f.EditingFileSystems {
		var readWrite bool
		for _, mps := range mpf.GetMountedPoints() {
			for _, mp := range mps {
				if readWrite = mp.RWFS != nil; readWrite {
					break
				}
			}
		}
		list = append(list, &editor.File{
			FSBT:     mpf.BaseTag().String(),
			FSID:     mpf.Tag().String(),
			Name:     mpf.Tag().String(),
			MimeType: beMime.DirectoryMimeType,
			ReadOnly: !readWrite,
		})
	}
	return
}

func (f *CFeature) ListFileSystemLocales(fsid string) (list editor.Files) {
	unique := map[string]struct{}{}
	for _, mpf := range f.EditingFileSystems {
		tag := mpf.Tag().String()
		if tag == fsid {
			for _, mps := range mpf.GetMountedPoints() {
				for _, mp := range mps {
					if mp.Mount != "/" {
						mount := strings.TrimSuffix(bePath.TrimSlashes(mp.Mount), "/static")
						if _, present := unique[mount]; present {
							continue
						}
						unique[mount] = struct{}{}
						list = append(list, &editor.File{
							FSBT:     mpf.BaseTag().String(),
							FSID:     tag,
							Code:     mount,
							Name:     mount,
							Locale:   &language.Und, // above locales in path-space is Und
							MimeType: beMime.DirectoryMimeType,
						})
					}
				}
			}
		}
	}
	sort.Slice(list, func(i, j int) (less bool) {
		less = natural.Less(list[i].Name, list[j].Name)
		return
	})
	return
}

func (f *CFeature) ListFileSystemDirectories(r *http.Request, fsid, code, dirs string) (list editor.Files) {
	printer := lang.GetPrinterFromRequest(r)
	//dirsPath := editor.MakeLangCodePath(code, dirs)
	dirsPath := dirs
	lookup := make(map[string]struct{})
	for _, mpf := range f.EditingFileSystems {
		tag := mpf.Tag().String()
		if tag == fsid {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					mount := bePath.TrimSlashes(mountPoint.Mount)
					if code != mount {
						continue
					}
					if found, err := mountPoint.ROFS.ListDirs(dirsPath); err == nil {

						for _, dir := range found {
							if _, exists := lookup[dir]; exists {
								continue
							} else {
								lookup[dir] = struct{}{}
							}
							var actions editor.Actions
							if mountPoint.RWFS != nil {
								if foundFiles, _ := mountPoint.ROFS.ListFiles(dir); len(foundFiles) == 0 {
									if foundPaths, _ := mountPoint.ROFS.ListDirs(dir); len(foundPaths) == 0 {
										actions = append(actions, editor.MakeDeletePathAction(printer, dir))
									}
								}
							}
							list = append(list, &editor.File{
								FSBT:     mpf.BaseTag().String(),
								FSID:     tag,
								Code:     mount,
								Path:     dir,
								Name:     filepath.Base(dir),
								Locale:   &language.Und,
								MimeType: beMime.DirectoryMimeType,
								ReadOnly: mountPoint.RWFS == nil,
								Actions:  actions,
							})
						}
					}
				}
			}
		}
	}
	list = list.Sort()
	return
}

func (f *CFeature) ListFileSystemFiles(r *http.Request, fsid, code, dirs string) (list editor.Files) {
	if fsid == "" {
		return
	}
	//isUnd := code == language.Und.String()
	eid := userbase.GetCurrentUserEID(r)
	printer := lang.GetPrinterFromRequest(r)
	//dirsPath := editor.MakeLangCodePath(code, dirs)
	dirsPath := dirs
	unique := map[string]struct{}{}
	for _, mpf := range f.EditingFileSystems {
		tag := mpf.Tag().String()
		if tag == fsid {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					mount := bePath.TrimSlashes(mountPoint.Mount)
					if code != mount {
						continue
					} else if _, present := unique[mount]; present {
						continue
					}
					unique[mount] = struct{}{}
					if files, err := mountPoint.ROFS.ListFiles(dirsPath); err == nil {
						for _, file := range files {
							if ef, ignored := f.SelfEditor().ProcessMountPointFile(r, printer, eid, mpf.BaseTag().String(), tag, code, dirs, file, mountPoint, false); !ignored {
								//if isUnd && ef.Locale.String() != language.Und.String() {
								//	continue
								//}
								ef.Name = ef.File
								ef.ReadOnly = mountPoint.RWFS == nil
								list = append(list, ef)
							}
						}
					}
				}
			}
		}
	}
	list = list.Sort()
	return
}

func (f *CFeature) ProcessMountPointFile(r *http.Request, printer *message.Printer, eid, mpfBTag, mpfTag, code, dirs, file string, mountPoint *feature.CMountPoint, draftWork bool) (ef *editor.File, ignored bool) {
	ef = editor.ParseFile(mpfTag, file)
	ef.FSBT = mpfBTag
	ef.Code = code
	ef.MountPoint = mountPoint
	if ignored = ef.Tilde != "" && (!draftWork || ef.Tilde != editor.DraftFile.String()); ignored {
		// ignore work-files
		return
	}
	if trimmed := bePath.TrimSlashes(dirs); dirs != "" && ef.Path != trimmed && !strings.HasPrefix(ef.Path, trimmed+"/") {
		// unwanted directory
		ignored = true
		return
	} else if mimeType, shasum, created, updated, err := mountPoint.ROFS.FileStats(file); err != nil {
		log.ErrorRF(r, "error getting file stats: %v - %v", file, err)
		ignored = true
		return
	} else if !beMime.IsPlainText(mimeType) {
		log.DebugRF(r, "ignoring binary file: %v - %v", file, mimeType)
		ignored = true
		return
	} else {
		ef.Shasum = shasum
		ef.MimeType = mimeType
		ef.Created = created
		ef.Updated = updated
	}

	f.SelfEditor().UpdateFileInfo(ef, r)
	return
}