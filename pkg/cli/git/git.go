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

package git

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/go-enjin/be/pkg/cli/env"
	"github.com/go-enjin/be/pkg/cli/run"
	bePath "github.com/go-enjin/be/pkg/path"
)

func Which() (gitBin string) {
	paths := env.GetPaths()
	for _, path := range paths {
		if gitBin = path + "/git"; bePath.IsFile(gitBin) {
			return
		}
	}
	gitBin = ""
	return
}

func IsRepo() (within bool) {
	return FindDotGit() != ""
}

func FindDotGit() (dotGit string) {
	var err error
	if bePath.IsDir(".git") {
		if dotGit, err = bePath.Abs(".git"); err != nil {
			dotGit = ".git"
			return
		}
		return
	}
	wd := bePath.Pwd()
	parts := strings.Split(wd, "/")
	for idx := len(parts) - 1; idx >= 0; idx-- {
		if dotGit = strings.Join(parts[0:idx], "/") + "/.git"; bePath.IsDir(dotGit) {
			return
		}
	}
	dotGit = ""
	return
}

func Describe() (tag string, ok bool) {
	if out, _, _, err := Cmd("describe", "--tags"); err == nil {
		tag = strings.TrimSpace(out)
		ok = true
	}
	return
}

func DiffHash(prefix string) (hash string, ok bool) {
	var err error
	var out string
	if out, _, _, err = Cmd("diff"); err != nil {
		return
	}
	sum := fmt.Sprintf("%x", sha256.Sum256([]byte(out)))
	if len(sum) == 0 {
		return
	}
	if prefix != "" {
		prefix += "-"
	}
	hash = prefix + sum[:10]
	ok = true
	return
}

func RevParse(prefix string) (id string, ok bool) {
	var err error
	var rev string
	if rev, _, _, err = Cmd("rev-parse", "--short=10", "HEAD"); err != nil {
		return
	}
	if prefix != "" {
		prefix += "-"
	}
	id = prefix + strings.TrimSpace(rev)
	ok = true
	return
}

func Status() (status string, ok bool) {
	var err error
	if status, _, _, err = Cmd("status", "--porcelain"); err != nil {
		return
	}
	status = strings.TrimSpace(status)
	ok = true
	return
}

func Cmd(argv ...string) (stdout, stderr string, status int, err error) {
	return run.Cmd("git", argv...)
}

func Exe(argv ...string) (status int, err error) {
	return run.Exe("git", argv...)
}

func MakeReleaseVersion() (release string) {
	release = MakeCustomVersion("release", "c", "d")
	return
}

func MakeCustomVersion(r, c, d string) (version string) {
	if status, ok := Status(); ok {
		if status == "" {
			version, _ = RevParse(c)
			return
		}
		version, _ = DiffHash(d)
		return
	}
	version = r
	return
}