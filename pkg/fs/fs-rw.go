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

package fs

import (
	"os"
	"time"

	"github.com/go-enjin/be/types/page/matter"
)

type RWFileSystem interface {
	FileSystem

	CloneRWFS() (cloned RWFileSystem)
	BeginTransaction()
	RollbackTransaction()
	CommitTransaction()
	EndTransaction()

	MakeDir(path string, perm os.FileMode) (err error)
	MakeDirAll(path string, perm os.FileMode) (err error)

	Remove(path string) (err error)
	RemoveAll(path string) (err error)

	WriteFile(path string, data []byte, perm os.FileMode) (err error)
	ChangeTimes(path string, created, updated time.Time) (err error)

	WritePageMatter(pm *matter.PageMatter) (err error)
	RemovePageMatter(path string) (err error)
}