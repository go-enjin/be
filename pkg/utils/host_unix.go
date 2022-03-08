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

package utils

import (
	"bytes"
	"os"
	"os/exec"
)

var _hostname string = ""

func Hostname() string {
	if _hostname == "" {
		if _, err := os.Stat("/bin/hostname"); err == nil {
			cmd := exec.Command("/bin/hostname", "--fqdn")
			var out bytes.Buffer
			cmd.Stdout = &out
			if err := cmd.Run(); err != nil {
				_hostname, _ = os.Hostname()
			} else {
				fqdn := out.String()
				_hostname = fqdn[:len(fqdn)-1] // removing EOL
			}
		}
	}
	return _hostname
}