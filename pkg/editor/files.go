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
	"sort"

	"github.com/maruel/natural"
)

type Files []*File

func (l Files) Sort() (sorted Files) {
	sorted = append(sorted, l...)
	sort.Slice(sorted, func(i, j int) (less bool) {
		switch {
		case sorted[i].FSID != sorted[j].FSID:
			less = natural.Less(sorted[i].FSID, sorted[j].FSID)

		case sorted[i].Code != sorted[j].Code:
			less = natural.Less(sorted[i].Code, sorted[j].Code)

		case sorted[i].Path != sorted[j].Path:
			less = natural.Less(sorted[i].Path, sorted[j].Path)

		case sorted[i].File != sorted[j].File:
			less = natural.Less(sorted[i].File, sorted[j].File)
		}
		return
	})
	return
}

func (l Files) Find(fsid, filePath string) (f *File) {
	for _, i := range l {
		if i.FSID == fsid && i.FilePath() == filePath {
			f = i
			return
		}
	}
	return
}
