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

package log

import (
	"fmt"
	"net/http"
	"regexp"
	"runtime"
	"strings"

	"github.com/go-enjin/be/pkg/request"
)

var (
	rxGoModuleVersion = regexp.MustCompile(`\@(.+?)/`)
	rxInvalidFuncName = regexp.MustCompile(`^\s*(\d+|func\d+)\s*$`)
)

func getLogPrefix(depth int, r *http.Request) string {
	depth += 1
	var file, name string
	var line int
	var ok bool
	if _, file, line, ok = runtime.Caller(depth); ok {
		file = rxGoModuleVersion.ReplaceAllString(file, "/")
		for i := depth; i < 20; i++ {
			if pc, _, _, ok := runtime.Caller(i); ok {
				fn := runtime.FuncForPC(pc).Name()
				if i := strings.LastIndex(fn, "."); i > -1 {
					fn = fn[i+1:]
				}
				if rxInvalidFuncName.MatchString(fn) {
					continue
				}
				name = fn
			}
			break
		}
	}
	if Config.LoggingFormat == FormatText {
		return "[" + name + "]"
	}
	if rid := request.GetRequestID(r); rid != "" {
		return fmt.Sprintf("[%s] %s:%d [%s]", rid, file, line, name)
	}
	return fmt.Sprintf("%s:%d [%s]", file, line, name)
}

func prefixLogEntry(depth int, format string, r *http.Request) string {
	depth += 1
	return fmt.Sprintf("%v %v", getLogPrefix(depth, r), format)
}
