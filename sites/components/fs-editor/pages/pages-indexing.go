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

package pages

import (
	"github.com/go-enjin/be/pkg/editor"
)

func (f *CFeature) HasIndexing(info *editor.File) (indexed bool) {
	indexed = f.Enjin.FindPageStub(info.Shasum) != nil
	return
}

func (f *CFeature) AddIndexing(info *editor.File) {
	for _, pfs := range f.pageFileSystems {
		if pfs.Tag().String() == info.FSID {
			pfs.AddIndexing(info.FilePath())
			return
		}
	}
}

func (f *CFeature) RemoveIndexing(info *editor.File) {
	for _, pfs := range f.pageFileSystems {
		if pfs.Tag().String() == info.FSID {
			pfs.RemoveIndexing(info.FilePath())
			return
		}
	}
}