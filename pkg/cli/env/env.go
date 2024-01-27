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

package env

import (
	"os"
	"regexp"
	"strings"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-corelibs/slices"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var Cache = context.New()

var (
	rxFixDupeColons = regexp.MustCompile(`:+`)
	rxParseEnvLines = regexp.MustCompile(`^\s*([^=]+?)\s*=\s*(.+?)\s*$`)
)

func init() {
	for _, raw := range os.Environ() {
		if rxParseEnvLines.MatchString(raw) {
			m := rxParseEnvLines.FindAllStringSubmatch(raw, 1)
			Cache.Set(m[0][1], m[0][2])
		}
	}
}

func Get(key, def string) (value string) {
	value = Cache.String(key, def)
	return
}

func Set(key, value string) {
	Cache.Set(key, value)
	_ = os.Setenv(key, value)
}

func GetPaths() (paths []string) {
	if rawPath := Cache.String("PATH", ""); rawPath != "" {
		rawPath = rxFixDupeColons.ReplaceAllString(rawPath, ":")
		rawParts := strings.Split(rawPath, ":")
		paths = append(paths, rawParts...)
		return
	}
	return
}

func SetPathSuffixed(path string) string {
	if paths := GetPaths(); len(paths) > 0 {
		paths = append(paths, path)
		envPath := strings.Join(paths, ":")
		envPath = rxFixDupeColons.ReplaceAllString(envPath, ":")
		Set("PATH", envPath)
		return envPath
	}
	Set("PATH", path)
	return path
}

func SetPathPrefixed(path string) string {
	if paths := GetPaths(); len(paths) > 0 {
		paths = append([]string{path}, paths...)
		envPath := strings.Join(paths, ":")
		envPath = rxFixDupeColons.ReplaceAllString(envPath, ":")
		Set("PATH", envPath)
		return envPath
	}
	Set("PATH", path)
	return path
}

func SetPathRemoved(path string) string {
	if paths := GetPaths(); len(paths) > 0 {
		for slices.Present(path, paths...) {
			if idx := beStrings.StringIndexInStrings(path, paths...); idx >= 0 {
				paths = slices.Remove(paths, idx)
			}
		}
		envPath := strings.Join(paths, ":")
		envPath = rxFixDupeColons.ReplaceAllString(envPath, ":")
		Set("PATH", envPath)
		return envPath
	}
	Set("PATH", "")
	return ""
}
