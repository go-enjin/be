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

package path

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	RxDupeSlashes   = regexp.MustCompile(`/+`)
	RxBothSlashes   = regexp.MustCompile(`^\s*/?(.+?)/?\s*$`)
	RxSlashDotSlash = regexp.MustCompile(`/\./`)
)

func CleanWithSlash(path string) (clean string) {
	clean = "/" + strings.Trim(path, "/\t ")
	clean = filepath.Clean(clean)
	return
}

func CleanWithSlashes(path string) (clean string) {
	clean = CleanWithSlash(path)
	clean += "/"
	return
}

func Join(parts ...string) (joined string) {
	joined = strings.Join(parts, string(os.PathSeparator))
	joined = RxDupeSlashes.ReplaceAllString(joined, "/")
	return
}

func JoinWithSlash(paths ...string) (joined string) {
	joined = strings.Join(paths, "/")
	joined = CleanWithSlash(joined)
	return
}

func JoinWithSlashes(paths ...string) (joined string) {
	joined = strings.Join(paths, "/")
	joined = CleanWithSlash(joined)
	return
}

func ToSlug(path string) (slug string) {
	if path == "" {
		slug = "/"
	} else if path != "/" && path[0] == '/' {
		slug = path[1:]
	} else {
		slug = path
	}
	return
}

func TrimSlash(path string) (clean string) {
	clean = "/" + TrimSlashes(path)
	clean = RxDupeSlashes.ReplaceAllString(clean, "/")
	return
}

func TrimSlashes(path string) (clean string) {
	if path == "" {
		return
	}
	clean = strings.TrimSpace(path)
	clean = RxBothSlashes.ReplaceAllString(clean, "$1")
	clean = RxDupeSlashes.ReplaceAllString(clean, "/")
	return
}

func SafeConcatRelPath(root string, paths ...string) (out string) {
	var outs []string
	for _, path := range paths {
		if v := TrimSlashes(path); v != "" {
			outs = append(outs, v)
		}
	}
	out = strings.Join(outs, "/")
	root = TrimSlashes(root)
	if rl := len(root); rl > 0 {
		if ol := len(out); ol > rl {
			if out[:rl] == root {
				out = out[rl+1:]
			}
		}
	}
	out = root + "/" + TrimSlashes(out)
	out = RxDupeSlashes.ReplaceAllString(out, "/")
	lout := len(out)
	if lout >= 2 {
		if out[:2] == "/." {
			out = out[2:]
		} else if out[lout-2:] == "/." {
			out = out[:lout-2]
		} else if out[lout-1:] == "/" {
			out = out[:lout-1]
		}
	}
	out = RxSlashDotSlash.ReplaceAllString(out, "/")
	return
}

func SafeConcatUrlPath(root string, paths ...string) (out string) {
	out = "/" + SafeConcatRelPath(root, paths...)
	return
}

func TrimPrefix(path, prefix string) (modified string) {
	prefix = TrimSlashes(prefix)
	modified = TrimSlashes(path)
	if pl := len(prefix); pl > 0 {
		if ml := len(modified); ml > pl {
			if modified[0:pl] == prefix {
				modified = modified[pl+1:]
			}
		} else {
			if modified == prefix {
				return ""
			}
		}
	}
	modified = TrimSlashes(modified)
	return
}

func TrimDotSlash(path string) (out string) {
	out = path
	if len(out) > 2 && out[0:2] == "./" {
		out = out[2:]
	}
	return
}

func GetSectionSlug(url string) (path, section, slug string) {
	if url == "" {
		return
	}
	if url[0] != '/' {
		url = "/" + url
	}
	slug = Base(url)
	section = TrimSlashes(Dir(url))
	path = "/" + section
	if section != "" {
		// section is the top of parent hierarchy, not the whole tree
		list := strings.Split(section, "/")
		section = list[0]
	}
	path = filepath.Clean(path)
	slug = filepath.Clean(slug)
	return
}

func TrimTrailingSlash(path string) (out string) {
	if out = path; out != "" {
		if last := len(out) - 1; out[last] == '/' {
			out = out[:last]
		}
	}
	return
}