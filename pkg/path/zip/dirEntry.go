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

package embed

import "io/fs"

type dirEntry struct {
	info fs.FileInfo
}

func (z *dirEntry) Name() (name string) {
	name = z.info.Name()
	return
}

func (z *dirEntry) IsDir() (is bool) {
	is = z.info.IsDir()
	return
}

func (z *dirEntry) Type() (mode fs.FileMode) {
	mode = z.info.Mode()
	return
}

func (z *dirEntry) Info() (info fs.FileInfo, err error) {
	info = z.info
	return
}
