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
	"net/http"
	"os"
	"path/filepath"
	"strings"

	clPath "github.com/go-corelibs/path"
	"github.com/go-corelibs/x-text/language"
	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	beMime "github.com/go-enjin/be/pkg/mime"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CEditorFeature[MakeTypedFeature]) FileExists(info *editor.File) (exists bool) {
	filePath := info.FilePath()
	for _, mpf := range f.EditingFileSystems {
		mpfTag := mpf.Tag().String()
		if mpfTag == info.FSID {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					if mountPoint.Mount != "/" && !strings.HasPrefix("/"+filePath, mountPoint.Mount) {
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

func (f *CEditorFeature[MakeTypedFeature]) ReadFile(info *editor.File) (data []byte, err error) {
	filePath := info.FilePath()
	for _, mpf := range f.EditingFileSystems {
		mpfTag := mpf.Tag().String()
		if mpfTag == info.FSID {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					if mountPoint.Mount != "/" && !strings.HasPrefix("/"+filePath, mountPoint.Mount) {
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

func (f *CEditorFeature[MakeTypedFeature]) WriteFile(info *editor.File, data []byte) (err error) {
	if info.ReadOnly {
		err = fmt.Errorf("file is read-only")
		return
	}
	filePath := info.FilePath()
	for _, mpf := range f.EditingFileSystems {
		mpfTag := mpf.Tag().String()
		if mpfTag == info.FSID {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					if mountPoint.Mount != "/" && !strings.HasPrefix("/"+filePath, mountPoint.Mount) {
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

func (f *CEditorFeature[MakeTypedFeature]) RemoveFile(info *editor.File) (err error) {
	if info.ReadOnly {
		err = fmt.Errorf("file is read-only")
		return
	}
	filePath := info.FilePath()
	for _, mpf := range f.EditingFileSystems {
		mpfTag := mpf.Tag().String()
		if mpfTag == info.FSID {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					if mountPoint.Mount != "/" && !strings.HasPrefix("/"+filePath, mountPoint.Mount) {
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

func (f *CEditorFeature[MakeTypedFeature]) RemoveDirectory(info *editor.File) (err error) {
	if info.ReadOnly {
		err = fmt.Errorf("directory is read-only")
		return
	}
	filePath := info.FilePath()
	for _, mpf := range f.EditingFileSystems {
		mpfTag := mpf.Tag().String()
		if mpfTag == info.FSID {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					if mountPoint.RWFS != nil && mountPoint.RWFS.Exists(filePath) {
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

func (f *CEditorFeature[MakeTypedFeature]) PrepareEditableFile(r *http.Request, info *editor.File) (editFile *editor.File) {
	eid := userbase.GetCurrentEID(r)
	printer := message.GetPrinter(r)

	for _, mpf := range f.EditingFileSystems {
		mpfTag := mpf.Tag().String()

		if info.FSID != "" && info.FSID != mpfTag {
			continue
		}

		mountedPoints := mpf.GetMountedPoints()
		for _, point := range maps.SortedKeys(mountedPoints) {
			for _, mountPoint := range mountedPoints[point] {
				if mountPoint.Mount != "/" && !strings.HasPrefix("/"+info.FilePath(), mountPoint.Mount) {
					continue
				} else if mountPoint.ROFS.Exists(info.FilePath()) {
					if ef, ignored := f.SelfEditor().ProcessMountPointFile(r, printer, eid, mpf.BaseTag().String(), mpfTag, info.Code, info.Path, info.FilePath(), mountPoint, true); ignored {
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

func (f *CEditorFeature[MakeTypedFeature]) UpdatePathInfo(info *editor.File, r *http.Request) {
	//eid := userbase.GetCurrentEID(r)
	//printer := message.GetPrinter(r)
	//if !info.ReadOnly {
	//	info.Actions = append(info.Actions, editor.MakeCreatePageAction(printer))
	//}
	info.Actions = info.Actions.Sort()
	return
}

func (f *CEditorFeature[MakeTypedFeature]) UpdateFileInfo(info *editor.File, r *http.Request) {
	eid := userbase.GetCurrentEID(r)
	printer := message.GetPrinter(r)

	info.HasDraft = f.SelfEditor().DraftExists(info)

	var isLockedBy bool
	if lockedBy, locked := f.IsEditorFileLocked(info.FSID, info.FilePath()); locked {
		isLockedBy = eid == lockedBy
		info.Locked = !isLockedBy
		info.LockedBy = lockedBy
	}

	info.Actions = editor.Actions{}
	mp, _ := info.MountPoint.(*feature.CMountPoint)

	if info.ReadOnly = mp.RWFS == nil; info.ReadOnly {

		info.Actions = append(info.Actions, editor.MakeViewFileAction(printer))

	} else if info.Locked {

		info.Actions = append(info.Actions, editor.MakeRetakeFileAction(printer, info.File))

	} else {

		info.Actions = append(info.Actions, editor.MakeEditFileAction(printer))

		if info.HasDraft {
			info.Actions = append(info.Actions, editor.MakePublishFileAction(printer, info.File))
		}

		if isLockedBy {
			info.Actions = append(info.Actions, editor.MakeUnlockFileAction(printer, info.File))
		}

		if info.HasDraft {
			info.Actions = append(info.Actions, editor.MakeDeleteDraftFileAction(printer, info.File))
		} else {
			info.Actions = append(info.Actions, editor.MakeDeleteFileAction(printer, info.File))
			info.Actions = append(info.Actions, editor.MakeMoveFileAction(printer, info.File))
		}

	}

	info.Actions = append(info.Actions, editor.MakeCopyFileAction(printer, info.File))
	info.Actions = info.Actions.Sort()

}

func (f *CEditorFeature[MakeTypedFeature]) UpdateFileInfoForEditing(info *editor.File, r *http.Request) {
	printer := message.GetPrinter(r)
	fileActions := info.Actions.Prune(editor.EditActionKey, editor.ViewActionKey, editor.UnlockActionKey)
	//for idx := 0; idx < len(fileActions); idx++ {
	//	fileActions[idx].Icon = ""
	//}

	if info.ReadOnly || info.Locked {
		info.Actions = fileActions
	} else {
		info.Actions = append(
			editor.Actions{
				editor.MakeCancelFileAction(printer),
				editor.MakeCommitFileAction(printer),
			},
			fileActions...,
		)
	}

	info.Actions = info.Actions.Sort()
}

func (f *CEditorFeature[MakeTypedFeature]) ProcessMountPointFile(r *http.Request, printer *message.Printer, eid, mpfBTag, mpfTag, code, dirs, file string, mountPoint *feature.CMountPoint, draftWork bool) (ef *editor.File, ignored bool) {
	ef = editor.ParseFile(mpfTag, file)
	ef.FSBT = mpfBTag
	ef.MountPoint = mountPoint
	if ignored = ef.Tilde != "" && (!draftWork || ef.Tilde != editor.DraftFile.String()); ignored {
		// ignore work-files
		return
	}
	if ignored = code != "" && ef.Locale.String() != code; ignored {
		// unwanted language
		return
	} else if trimmed := clPath.TrimSlashes(dirs); dirs != "" && ef.Path != trimmed && !strings.HasPrefix(ef.Path, trimmed+"/") {
		// unwanted directory
		ignored = true
		return
	} else if hasValidExtension := f.EditAnyFileExtension || clPath.HasAnyExt(ef.File, f.EditingFileExtensions...); !hasValidExtension {
		log.DebugRF(r, "ignoring file by extension: %v (not any of: %+v)", file, f.EditingFileExtensions)
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

func (f *CEditorFeature[MakeTypedFeature]) ListFileSystems() (list editor.Files) {
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
	list = list.Sort()
	return
}

func (f *CEditorFeature[MakeTypedFeature]) ListFileSystemLocales(fsid string) (list editor.Files) {
	for _, mpf := range f.EditingFileSystems {
		tag := mpf.Tag().String()
		if tag == fsid {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					if found, err := mountPoint.ROFS.ListDirs("."); err == nil {
						for _, dir := range found {
							if lt, ee := language.Parse(dir); ee == nil {
								list = append(list, &editor.File{
									FSBT:     mpf.BaseTag().String(),
									FSID:     tag,
									Code:     lt.String(),
									Name:     lt.String(),
									Locale:   &lt,
									MimeType: beMime.DirectoryMimeType,
									ReadOnly: mountPoint.RWFS == nil,
								})
							}
						}
						list = append(list, &editor.File{
							FSBT:     mpf.BaseTag().String(),
							FSID:     tag,
							Code:     language.Und.String(),
							Name:     language.Und.String(),
							Locale:   &language.Und,
							MimeType: beMime.DirectoryMimeType,
							ReadOnly: mountPoint.RWFS == nil,
						})
					}
				}
			}
		}
	}
	list = list.Sort()
	return
}

func (f *CEditorFeature[MakeTypedFeature]) ListFileSystemDirectories(r *http.Request, fsid, code, dirs string) (list editor.Files) {
	printer := message.GetPrinter(r)
	isUnd := code == language.Und.String()
	dirsPath := editor.MakeLangCodePath(code, dirs)
	lookup := make(map[string]struct{})
	for _, mpf := range f.EditingFileSystems {
		tag := mpf.Tag().String()
		if tag == fsid {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					if found, err := mountPoint.ROFS.ListDirs(dirsPath); err == nil {
						for _, dir := range found {
							lt := language.Und
							var topDir, dirPath string
							if td := clPath.TopDirectory(dir); td != "" {
								topDir = td
							} else {
								topDir = dir
							}
							if parsed, ee := language.Parse(topDir); ee == nil {
								lt = parsed
								if strings.HasPrefix(dir, lt.String()+"/") {
									dirPath = strings.TrimPrefix(dir, lt.String()+"/")
								}
							}
							if isUnd && lt.String() != language.Und.String() {
								continue
							} else if _, exists := lookup[dir]; exists {
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
								Code:     code,
								Path:     dirPath,
								Name:     filepath.Base(dirPath),
								Locale:   &lt,
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

func (f *CEditorFeature[MakeTypedFeature]) ListFileSystemFiles(r *http.Request, fsid, code, dirs string) (list editor.Files) {
	if fsid == "" {
		return
	}
	isUnd := code == language.Und.String()
	eid := userbase.GetCurrentEID(r)
	printer := message.GetPrinter(r)
	dirsPath := editor.MakeLangCodePath(code, dirs)
	for _, mpf := range f.EditingFileSystems {
		tag := mpf.Tag().String()
		if tag == fsid {
			for _, mountPoints := range mpf.GetMountedPoints() {
				for _, mountPoint := range mountPoints {
					if files, err := mountPoint.ROFS.ListFiles(dirsPath); err == nil {
						for _, file := range files {
							if ef, ignored := f.SelfEditor().ProcessMountPointFile(r, printer, eid, mpf.BaseTag().String(), tag, code, dirs, file, mountPoint, false); !ignored {
								if isUnd && ef.Locale.String() != language.Und.String() {
									continue
								}
								ef.Name = ef.File
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
