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

package tar

import (
	"os"
	"os/exec"

	"github.com/go-enjin/be/pkg/path"
)

func UnTarGz(src, dst string) (out string, err error) {
	cmd := exec.Command(
		"tar",
		"--extract",
		"--directory", dst,
		"--file", src,
	)
	if !path.IsDir(dst) {
		if err = os.MkdirAll(dst, 0770); err != nil {
			return
		}
	}
	var data []byte
	if data, err = cmd.CombinedOutput(); err == nil {
		out = string(data)
	}
	return
}