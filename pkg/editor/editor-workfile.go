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
	"strings"
)

const (
	NilFile   WorkFile = ""
	LockFile  WorkFile = "lock"
	DraftFile WorkFile = "draft"
)

type WorkFile string

func ParseEditorWorkFile(filename string) (modified string, wf WorkFile, ok bool) {
	if before, after, found := strings.Cut(filename, ".~"); found && after != "" {
		wf = EditorWorkFiles.Lookup(after)
		if ok = wf != NilFile; ok {
			modified = before
			return
		}
	}
	modified = filename
	return
}

func (wf WorkFile) String() (name string) {
	name = string(wf)
	return
}

func (wf WorkFile) Is(name string) (is bool) {
	is = wf.String() == name
	return
}
