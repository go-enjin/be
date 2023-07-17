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

package globals

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/go-enjin/github-com-djherbis-times"
)

var (
	// BinName is the name of the actual binary compiled
	BinName string = ""
	// BinHash is set at runtime with the shasum of the compiled binary
	BinHash string = ""
	// Release is a revision indicator such as a short git commit id
	Release string = ""
	// Version is the standard semantic versioning of this release
	Version string = ""
	// Summary is used on the command line and other cosmetic places
	Summary string = ""
	// EnvPrefix is used as a prefix to all CLI environment variables
	EnvPrefix string = ""
	// DefaultPort is the fallback port to open for connections
	DefaultPort = 3334
	// DefaultListen is the fallback address to listen on
	DefaultListen = ""
	// SlugIntegrity is the expected hash of a Shasums file, without BinName
	// present (set by enjenv)
	SlugIntegrity = ""
	// SumsIntegrity is the expected hash of a Shasums file (set by enjenv)
	SumsIntegrity = ""
	// Hostname is set at runtime with the output of os.Hostname
	Hostname = ""
)

var rxHerokuHostname = regexp.MustCompile(`^\s*([a-f0-9]{8})-([a-f0-9]{4})-([a-f0-9]{4})-([a-f0-9]{4})-([a-z0-9]{12})\s*$`)

func init() {
	if Hostname, _ = os.Hostname(); Hostname == "" {
		Hostname = "os-Hostname-error"
		return
	}
	//  2be2ef99-65c1-4a47-a9d0-4bd9a1f203b9 buildpack build_23a520f7
	if rxHerokuHostname.MatchString(Hostname) {
		m := rxHerokuHostname.FindStringSubmatch(Hostname)
		Hostname = "heroku-" + m[1]
	}
}

func BuildVersion() (version string) {
	return fmt.Sprintf(
		"%v [%v] (%v)",
		Version,
		Release,
		BinHash,
	)
}

func BuildInfoString() string {
	return fmt.Sprintf(
		"%v [r=%v go=%v] (%v)",
		Version,
		Release,
		BinHash,
		Hostname,
	)
}

var (
	buildFileInfo times.Timespec
)

func BuildFileInfo() (info times.Timespec, err error) {
	if buildFileInfo != nil {
		info = buildFileInfo
		return
	}
	var name, tgt string
	if name, err = os.Executable(); err == nil {
		if tgt, err = filepath.EvalSymlinks(name); err == nil {
			if info, err = times.Stat(tgt); err == nil {
				buildFileInfo = info
			}
		}
	}
	return
}