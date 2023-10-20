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

package editor

import (
	"github.com/go-enjin/golang-org-x-text/message"
)

const (
	ViewActionKey         = "view"
	PreviewDraftActionKey = "preview-draft"
	ViewErrorActionKey    = "preview-error"
	EditActionKey         = "edit"
	UnlockActionKey       = "unlock"
	RetakeActionKey       = "retake"
	DeleteActionKey       = "delete"
	DeleteDraftActionKey  = "delete-draft"
	DeletePathActionKey   = "delete-path"
	CommitActionKey       = "commit"
	PublishActionKey      = "publish"
	IndexPageActionKey    = "index-page"
	DeIndexPageActionKey  = "de-index-page"
	CancelActionKey       = "cancel"
	MoveActionKey         = "move"
	CopyActionKey         = "copy"
	CreatePageActionKey   = "create-page"
	CreateMenuActionKey   = "create-menu"
	TranslateActionKey    = "translate"
	ChangeActionKey       = "change"
	SearchActionKey       = "search"
)

func MakeViewErrorAction(printer *message.Printer) (action *Action) {
	return &Action{
		Key:    ViewErrorActionKey,
		Name:   printer.Sprintf("View error"),
		Icon:   "fa-solid fa-eye",
		Class:  "danger",
		Active: true,
		Method: GetFormMethod,
		Tilde:  DraftFile.String(),
		Order:  0,
	}
}

func MakeDeletePathAction(printer *message.Printer, dirPath string) (action *Action) {
	return &Action{
		Key:    DeletePathActionKey,
		Name:   printer.Sprintf("Delete path"),
		Icon:   "fa-solid fa-delete-left",
		Class:  "danger",
		Active: true,
		Method: PostFormMethod,
		Prompt: printer.Sprintf(`Really delete the directory "%[1]s"?`, dirPath),
		Order:  1,
	}
}

func MakeCreatePageAction(printer *message.Printer) (action *Action) {
	return &Action{
		Key:    CreatePageActionKey,
		Name:   printer.Sprintf("Create page"),
		Icon:   "fa-solid fa-plus",
		Class:  "important",
		Active: true,
		Method: PostFormMethod,
		Prompt: printer.Sprintf("Create a new page?"),
		Dialog: "create-page",
		Order:  1,
	}
}

func MakeCreateMenuAction(printer *message.Printer) (action *Action) {
	return &Action{
		Key:    CreateMenuActionKey,
		Name:   printer.Sprintf("Create menu"),
		Icon:   "fa-solid fa-plus",
		Class:  "important",
		Active: true,
		Method: PostFormMethod,
		Prompt: printer.Sprintf("Create a new menu?"),
		Dialog: "create-menu",
		Order:  1,
	}
}

func MakeViewFileAction(printer *message.Printer) (action *Action) {
	return &Action{
		Key:    ViewActionKey,
		Name:   printer.Sprintf("View (read-only)"),
		Icon:   "fa-solid fa-eye",
		Class:  "primary",
		Active: true,
		Method: GetFormMethod,
		Order:  4,
	}
}

func MakeEditFileAction(printer *message.Printer) (action *Action) {
	return &Action{
		Key:    EditActionKey,
		Name:   printer.Sprintf("Edit draft"),
		Icon:   "fa-solid fa-file-pen",
		Class:  "primary",
		Active: true,
		Method: GetFormMethod,
		Order:  1,
	}
}

func MakeCommitFileAction(printer *message.Printer) (action *Action) {
	return &Action{
		Key:    CommitActionKey,
		Name:   printer.Sprintf("Save draft"),
		Icon:   "fa-solid fa-floppy-disk",
		Class:  "primary",
		Active: true,
		Method: PostFormMethod,
		Order:  1,
	}
}

func MakePreviewDraftAction(printer *message.Printer) (action *Action) {
	return &Action{
		Key:    PreviewDraftActionKey,
		Name:   printer.Sprintf("Preview draft"),
		Icon:   "fa-solid fa-eye",
		Class:  "primary",
		Active: true,
		Method: GetFormMethod,
		Tilde:  DraftFile.String(),
		Order:  2,
	}
}

func MakePublishFileAction(printer *message.Printer, filename string) (action *Action) {
	return &Action{
		Key:    PublishActionKey,
		Name:   printer.Sprintf("Publish draft"),
		Icon:   "fa-solid fa-upload",
		Class:  "important",
		Active: true,
		Method: PostFormMethod,
		Prompt: printer.Sprintf(`Publish the draft of "%[1]s"?`, filename),
		Order:  5,
	}
}

func MakeCancelFileAction(printer *message.Printer) (action *Action) {
	return &Action{
		Key:    CancelActionKey,
		Name:   printer.Sprintf("Cancel editing"),
		Icon:   "fa-solid fa-ban",
		Class:  "caution",
		Active: true,
		Method: PostFormMethod,
		Order:  9,
	}
}

func MakeRetakeFileAction(printer *message.Printer, filename string) (action *Action) {
	return &Action{
		Key:    RetakeActionKey,
		Name:   printer.Sprintf("Take over"),
		Icon:   "fa-solid fa-user-lock",
		Class:  "danger",
		Active: true,
		Method: PostFormMethod,
		Prompt: printer.Sprintf(`Take over editing of "%[1]s"?`, filename),
		Order:  10,
	}
}

func MakeUnlockFileAction(printer *message.Printer, filename string) (action *Action) {
	return &Action{
		Key:    UnlockActionKey,
		Name:   printer.Sprintf("Unlock editing"),
		Icon:   "fa-solid fa-user-lock",
		Class:  "caution",
		Active: true,
		Method: PostFormMethod,
		Prompt: printer.Sprintf(`Unlock "%[1]s" for editing by others?`, filename),
		Order:  10,
	}
}

func MakeMoveFileAction(printer *message.Printer, filename string) (action *Action) {
	return &Action{
		Key:    MoveActionKey,
		Name:   printer.Sprintf("Move/rename"),
		Icon:   "fa-solid fa-arrow-right-to-bracket",
		Class:  "caution",
		Active: true,
		Method: PostFormMethod,
		Prompt: printer.Sprintf(`Move/rename "%[1]s"?`, filename),
		Dialog: "move-file",
		Order:  25,
	}
}

func MakeCopyFileAction(printer *message.Printer, filename string) (action *Action) {
	return &Action{
		Key:    CopyActionKey,
		Name:   printer.Sprintf("Copy"),
		Icon:   "fa-solid fa-copy",
		Class:  "important",
		Active: true,
		Method: PostFormMethod,
		Prompt: printer.Sprintf(`Copy "%[1]s" to another location?`, filename),
		Dialog: "copy-file",
		Order:  50,
	}
}

func MakeDeleteDraftFileAction(printer *message.Printer, filename string) (action *Action) {
	return &Action{
		Key:    DeleteDraftActionKey,
		Name:   printer.Sprintf("Delete draft"),
		Icon:   "fa-solid fa-delete-left",
		Class:  "danger",
		Active: true,
		Method: PostFormMethod,
		Prompt: printer.Sprintf(`Really delete the draft of "%[1]s"?`, filename),
		Order:  75,
	}
}

func MakeDeleteFileAction(printer *message.Printer, filename string) (action *Action) {
	return &Action{
		Key:    DeleteActionKey,
		Name:   printer.Sprintf("Delete"),
		Icon:   "fa-solid fa-delete-left",
		Class:  "danger",
		Active: true,
		Method: PostFormMethod,
		Prompt: printer.Sprintf(`Really delete "%[1]s"?`, filename),
		Order:  75,
	}
}

func MakeTranslateAction(printer *message.Printer, filename string) (action *Action) {
	return &Action{
		Key:    TranslateActionKey,
		Name:   printer.Sprintf("Translate"),
		Icon:   "fa-solid fa-plus",
		Class:  "important",
		Active: true,
		Method: PostFormMethod,
		Prompt: printer.Sprintf(`Start a new translation of "%[1]s"?`, filename),
		Dialog: "translate-file",
		Order:  90,
	}
}

func MakeIndexPageAction(printer *message.Printer, filename string) (action *Action) {
	return &Action{
		Key:    IndexPageActionKey,
		Name:   printer.Sprintf("Add indexing"),
		Icon:   "fa-solid fa-arrow-rotate-left",
		Class:  "primary",
		Active: true,
		Method: PostFormMethod,
		Prompt: printer.Sprintf(`Add page indexing on "%[1]s"?`, filename),
		Order:  100,
	}
}

func MakeDeIndexPageAction(printer *message.Printer, filename string) (action *Action) {
	return &Action{
		Key:    DeIndexPageActionKey,
		Name:   printer.Sprintf("Remove indexing"),
		Icon:   "fa-solid fa-arrow-rotate-right",
		Class:  "primary",
		Active: true,
		Method: PostFormMethod,
		Prompt: printer.Sprintf(`Remove page indexing on "%[1]s"?`, filename),
		Order:  100,
	}
}