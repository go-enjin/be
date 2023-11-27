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

package slug

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/maruel/natural"

	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/hash/sha"
	bePath "github.com/go-enjin/be/pkg/path"
)

var (
	FileName = "Slugfile"
	SumsName = "Slugsums"
)

var (
	rxSlugsumsLine = regexp.MustCompile(`(?ms)^\s*([0-9a-f]{64})\s*([^/].+?)\s*$`)
)

type ShaMap map[string]string

func (m ShaMap) Keys() (keys []string) {
	for key := range m {
		keys = append(keys, key)
	}
	sort.Sort(natural.StringSlice(keys))
	return
}

func (m ShaMap) SlugIntegrity() (shasum string, err error) {
	var keys []string
	for _, key := range m.Keys() {
		if key != globals.BinName {
			keys = append(keys, key)
		}
	}
	content := ""
	for _, key := range keys {
		content += fmt.Sprintf("%v %v\n", m[key], key)
	}
	shasum, err = sha.DataHash64([]byte(content))
	return
}

func (m ShaMap) CheckSlugIntegrity() (err error) {
	var shasum string
	if shasum, err = m.SlugIntegrity(); err != nil {
		return
	}
	if globals.SlugIntegrity == "" {
		err = fmt.Errorf("integrity value not set at build-time")
		return
	}
	if shasum == globals.SlugIntegrity {
		return
	}
	err = fmt.Errorf("integrity check failed, expected: \"%v\"", globals.SlugIntegrity)
	return
}

func (m ShaMap) SumsIntegrity() (shasum string, err error) {
	content := ""
	for _, key := range m.Keys() {
		content += fmt.Sprintf("%v %v\n", m[key], key)
	}
	shasum, err = sha.DataHash64([]byte(content))
	return
}

func (m ShaMap) CheckSumsIntegrity() (err error) {
	var shasum string
	if shasum, err = m.SumsIntegrity(); err != nil {
		return
	}
	if globals.SumsIntegrity == "" {
		err = fmt.Errorf("integrity value not set at run-time")
		return
	}
	if shasum == globals.SumsIntegrity {
		return
	}
	err = fmt.Errorf("integrity check failed, expected: \"%v\"", globals.SumsIntegrity)
	return
}

func SlugsumsPresent() (ok bool) {
	slugsums := bePath.FindFileRelativeToPwd(SumsName)
	return slugsums != ""
}

func SlugfilePresent() (ok bool) {
	slugfile := bePath.FindFileRelativeToPwd(FileName)
	return slugfile != ""
}

func ReadSlugsums() (slugMap ShaMap, err error) {
	slugMap = make(ShaMap)
	slugsums := bePath.FindFileRelativeToPwd(SumsName)
	if slugsums != "" {
		var data []byte
		if data, err = bePath.ReadFile(slugsums); err != nil {
			return
		}
		content := string(data)
		m := rxSlugsumsLine.FindAllStringSubmatch(content, -1)
		for _, mm := range m {
			slugMap[mm[2]] = mm[1]
		}
	}
	return
}

func ReadSlugfile() (paths []string, err error) {
	slugfile := bePath.FindFileRelativeToPwd(FileName)
	if slugfile != "" {
		var data []byte
		if data, err = bePath.ReadFile(slugfile); err != nil {
			return
		}
		content := string(data)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				paths = append(paths, trimmed)
			}
		}
	}
	return
}

func BuildSlugMap() (slugMap ShaMap, err error) {
	slugMap, err = BuildSlugMapIgnoring()
	return
}

var rxRestrictedPaths = regexp.MustCompile(`(?:/|^)\.(git|gpg)(?:/|$)`)

func isRestrictedPath(path string) (restricted bool) {
	restricted = rxRestrictedPaths.MatchString(path)
	return
}

func BuildSlugMapIgnoring(files ...string) (slugMap ShaMap, err error) {
	var paths []string
	if paths, err = ReadSlugfile(); err != nil {
		return
	}
	isIgnored := func(name string) (ignored bool) {
		for _, file := range files {
			if name == file {
				ignored = true
				return
			}
		}
		return
	}
	var topFiles, topDirs []string
	if topFiles, err = bePath.ListFiles("."); err != nil {
		return
	}
	if topDirs, err = bePath.ListDirs("."); err != nil {
		return
	}
	slugMap = make(ShaMap)
	for _, path := range paths {
		if isIgnored(path) {
			continue
		} else if isRestrictedPath(path) {
			continue
		}
		if strings.HasPrefix(path, "!") {
			pattern := strings.TrimSpace(path[1:])
			var rx *regexp.Regexp
			if rx, err = regexp.Compile(pattern); err != nil {
				err = fmt.Errorf("error compiling Slugfile regular expression: %v - %v", pattern, err)
				return
			}
			for _, file := range topFiles {
				if strings.HasPrefix(file, "./") {
					file = file[2:]
				}
				if rx.MatchString(file) {
					if slugMap[file], err = sha.FileHash64(file); err != nil {
						return
					}
				}
			}
			for _, dir := range topDirs {
				if strings.HasPrefix(dir, "./") {
					dir = dir[2:]
				}
				if rx.MatchString(dir) {
					var subPaths []string
					if subPaths, err = bePath.FindAllFiles(dir, false); err != nil {
						return
					}
					for _, subPath := range subPaths {
						if slugMap[subPath], err = sha.FileHash64(subPath); err != nil {
							return
						}
					}
				}
			}
			continue
		}
		if bePath.IsFile(path) {
			if slugMap[path], err = sha.FileHash64(path); err != nil {
				return
			}
			continue
		}
		if bePath.IsDir(path) {
			var subPaths []string
			if subPaths, err = bePath.FindAllFiles(path, false); err != nil {
				return
			}
			for _, subPath := range subPaths {
				if slugMap[subPath], err = sha.FileHash64(subPath); err != nil {
					return
				}
			}
			continue
		}
		err = fmt.Errorf("missing path from Slugfile: %v", path)
		return
	}
	return
}

func BuildFileMap() (fileMap map[string]string, err error) {
	var wd string
	if wd, err = os.Getwd(); err != nil {
		return
	}
	var paths []string
	if paths, err = bePath.FindAllFiles(wd, true); err != nil {
		return
	}
	fileMap = make(ShaMap)
	for _, path := range paths {
		if isRestrictedPath(path) {
			continue
		}
		path = bePath.TrimRelativeToRoot(path, wd)
		if bePath.IsFile(path) {
			if fileMap[path], err = sha.FileHash64(path); err != nil {
				return
			}
			continue
		}
		if bePath.IsDir(path) {
			var subPaths []string
			if subPaths, err = bePath.FindAllFiles(path, true); err != nil {
				return
			}
			for _, subPath := range subPaths {
				if fileMap[subPath], err = sha.FileHash64(subPath); err != nil {
					return
				}
			}
			continue
		}
		err = fmt.Errorf("not a file or directory: %v", path)
		return
	}
	return
}

func FinalizeSlugfile(force bool) (slugsums string, removed []string, err error) {
	if !force {
		err = fmt.Errorf("unintentionally finalizing a slug prevented")
		return
	}

	// read Slugfile, build slug map
	var slugMap ShaMap
	if slugMap, err = BuildSlugMap(); err != nil {
		err = fmt.Errorf("error building slug map: %v", err)
		return
	}

	var fileMap ShaMap
	if fileMap, err = BuildFileMap(); err != nil {
		err = fmt.Errorf("error building file map: %v", err)
		return
	}

	for _, file := range fileMap.Keys() {
		// for each file present
		if _, ok := slugMap[file]; !ok {
			// removing those not present in the slug map
			bePath.ChmodAll(file)
			if err = os.Remove(file); err != nil {
				err = fmt.Errorf("error removing file: %v - %v", file, err)
				return
			}
			removed = append(removed, file)
		}
	}

	for _, file := range slugMap.Keys() {
		slugsums += fmt.Sprintf("%v %v\n", slugMap[file], file)
	}
	if err = os.WriteFile(SumsName, []byte(slugsums), 0660); err != nil {
		err = fmt.Errorf("error writing %v: %v", SumsName, err)
		return
	}

	var all []string
	if all, err = bePath.FindAllDirs(".", true); err != nil {
		err = fmt.Errorf("error finding all dirs: %v", err)
		return
	}
	sort.Slice(all, func(i, j int) bool {
		return len(all[i]) > len(all[j])
	})

	var unaccounted []string
	for _, dir := range all {
		dir = bePath.TrimDotSlash(dir)
		dl := len(dir)
		accounted := false
		for file, _ := range slugMap {
			fl := len(file)
			if dl < fl {
				if file[:dl] == dir {
					accounted = true
					break
				}
			}
		}
		if !accounted {
			unaccounted = append(unaccounted, dir)
		}
	}

	for _, dir := range unaccounted {
		if bePath.IsDir(dir) {
			bePath.ChmodAll(dir)
			if err = os.RemoveAll(dir); err != nil {
				err = fmt.Errorf("error removing dir: %v - %v", dir, err)
				return
			}
			removed = append(removed, dir+"/")
		}
	}
	return
}

func ValidateSlugsums() (imposters, extraneous, validated []string, err error) {
	_, _, imposters, extraneous, validated, err = ValidateSlugsumsComplete()
	return
}

func ValidateSlugsumsComplete() (slugMap, fileMap ShaMap, imposters, extraneous, validated []string, err error) {
	if slugMap, err = ReadSlugsums(); err != nil {
		return
	}
	if fileMap, err = BuildFileMap(); err != nil {
		return
	}
	for file, sum := range slugMap {
		if v, ok := fileMap[file]; ok {
			if sum == v {
				validated = append(validated, file)
			} else {
				imposters = append(imposters, file)
			}
		}
	}
	for file, _ := range fileMap {
		if file == SumsName {
			continue
		}
		if _, ok := slugMap[file]; !ok {
			extraneous = append(extraneous, file)
		}
	}
	return
}

func WriteSlugfile(argv ...string) (slugfile string, err error) {
	var files []string
	wd, _ := bePath.Abs(bePath.Pwd())
	for _, arg := range argv {
		absArg, _ := bePath.Abs(arg)
		relArg := bePath.TrimRelativeToRoot(absArg, wd)
		if relArg == "" || relArg[0] == '/' {
			err = fmt.Errorf("external file paths not allowed: %v", arg)
			return
		}
		files = append(files, strings.TrimSpace(relArg))
	}
	sort.Sort(natural.StringSlice(files))
	contents := strings.Join(files, "\n")
	if err = os.WriteFile(FileName, []byte(contents), 0660); err == nil {
		slugfile = contents
	}
	return
}
