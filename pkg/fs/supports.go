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
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	strings2 "github.com/go-enjin/be/pkg/strings"
)

func PruneRootPrefixes(path string) (pruned string) {
	pruned = strings.TrimPrefix(path, "./")
	pruned = strings.TrimPrefix(path, "/")
	if pruned == "." {
		pruned = ""
	} else {
		pruned = filepath.Clean(pruned)
	}
	return
}

func PruneRootFrom[T string | []string](root string, path T) (pruned T) {
	switch t := interface{}(&path).(type) {
	case *string:
		var modified string
		modified = PruneRootPrefixes(*t)
		modified = strings.TrimPrefix(modified, PruneRootPrefixes(root))
		modified = PruneRootPrefixes(modified)
		// modified = "/" + modified
		pruned, _ = interface{}(modified).(T)
	case *[]string:
		var modified []string
		for _, p := range *t {
			modified = append(modified, PruneRootFrom(root, p))
		}
		pruned, _ = interface{}(modified).(T)
	default:
		panic(fmt.Errorf("unsupported type union: (%T) %#+v", path, path))
	}
	return
}

func LookupFilePath(fs FileSystem, basePath string, extensions ...string) (path string, present bool) {
	sort.Sort(strings2.SortByLengthDesc(extensions))
	for _, extension := range extensions {
		p := basePath + "." + extension
		if present = fs.Exists(p); present {
			path = p
			return
		}
	}
	return
}