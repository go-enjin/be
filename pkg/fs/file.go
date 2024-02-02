// Copyright (c) 2022  The Go-Enjin Authors
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

	clPath "github.com/go-corelibs/path"
	beMime "github.com/go-enjin/be/pkg/mime"
)

type File struct {
	Path string
	Name string
	Extn string
	Mime string
	Data []byte
}

func New(path string) (file *File, err error) {
	var data []byte
	if data, err = os.ReadFile(path); err != nil {
		return
	}
	file = &File{
		Path: path,
		Name: clPath.Base(path),
		Extn: clPath.Ext(path),
		Mime: beMime.Mime(path),
		Data: data,
	}
	return
}

func (f File) String() string {
	return string(f.Data)
}
