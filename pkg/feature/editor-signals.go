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
	"net/http"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature/signaling"
)

const (
	PreLockFileSignal            signaling.Signal = "pre-lock-file"
	LockFileSignal               signaling.Signal = "lock-file"
	PreUnlockFileSignal          signaling.Signal = "pre-unlock-file"
	UnlockFileSignal             signaling.Signal = "unlock-file"
	PreMoveFileSignal            signaling.Signal = "pre-move-file"
	MoveFileSignal               signaling.Signal = "move-file"
	PreCopyFileSignal            signaling.Signal = "pre-copy-file"
	CopyFileSignal               signaling.Signal = "copy-file"
	PrePublishFileSignal         signaling.Signal = "pre-publish-file"
	PublishFileSignal            signaling.Signal = "publish-file"
	PreRepublishFileSignal       signaling.Signal = "pre-republish-file"
	RepublishFileSignal          signaling.Signal = "republish-file"
	PreDeleteFileSignal          signaling.Signal = "pre-delete-file"
	DeleteFileSignal             signaling.Signal = "delete-file"
	PreDeletePathSignal          signaling.Signal = "pre-delete-path"
	DeletePathSignal             signaling.Signal = "delete-path"
	PreRetakeFileSignal          signaling.Signal = "pre-retake-file"
	RetakeFileSignal             signaling.Signal = "retake-file"
	PreCommitFileSignal          signaling.Signal = "pre-commit-file"
	CommitFileSignal             signaling.Signal = "commit-file"
	FileNameRequiredSignal       signaling.Signal = "file-name-required"
	PreTranslateFileActionSignal signaling.Signal = "pre-translate-file"
	TranslateFileActionSignal    signaling.Signal = "translate-file"
	PreChangeActionSignal        signaling.Signal = "pre-change-action"
	ChangeActionSignal           signaling.Signal = "change-action"
)

// ParseSignalArgv is a helper function for translating the emitted signal argv into concrete types
//
//	`r, pg, ctx, form, info, eid, file, redirect, ok := bePkgEditor.ParseSignalArgv(argv)`
func ParseSignalArgv(argv []interface{}) (r *http.Request, pg Page, ctx, form beContext.Context, info *editor.File, eid string, redirect *string, ok bool) {
	if ok = len(argv) == 7; !ok {
	} else if r, ok = argv[0].(*http.Request); !ok {
	} else if pg, ok = argv[1].(Page); !ok {
	} else if ctx, ok = argv[2].(beContext.Context); !ok {
	} else if form, ok = argv[3].(beContext.Context); !ok {
	} else if info, ok = argv[4].(*editor.File); !ok {
	} else if eid, ok = argv[5].(string); !ok {
	} else if redirect, ok = argv[6].(*string); !ok {
	}
	return
}
