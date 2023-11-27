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
	"strings"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
)

type EditorValidateFn = func(r *http.Request, pg Page, ctx, form beContext.Context, info *editor.File, eid string) (err error)
type EditorOperationFn = func(r *http.Request, pg Page, ctx, form beContext.Context, info *editor.File, eid string) (redirect string)

type EditorOperation struct {
	// Key is the submit button kebab-cased value
	Key string `json:"key"`
	// Confirm is the confirmation form parameter, leave empty for no confirmation step
	Confirm string `json:"confirm,omitempty"`
	// Action is the permission users are required to have in order to perform the operation
	Action Action
	// Validate is the handler called to sanitize the form ctx
	Validate EditorValidateFn
	// Operation is the handler called to perform the operation
	Operation EditorOperationFn
}

func ParseEditorOpKey(action string) (op, tgt string) {
	op, tgt, _ = strings.Cut(action, ".")
	return
}

func ParseEditorOpTargetValue(target string) (tgt, value string) {
	tgt, value, _ = strings.Cut(target, ":")
	return
}
