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

package profiling

import (
	"strings"

	"github.com/pkg/profile"

	"github.com/go-enjin/be/pkg/cli/env"
	"github.com/go-enjin/be/pkg/log"
)

var profiler interface {
	Stop()
}

func Start() {
	if mode := env.Get("BE_PROFILE_MODE", ""); mode != "" {
		path := env.Get("BE_PROFILE_PATH", ".")
		options := []func(*profile.Profile){
			profile.ProfilePath(path),
			profile.NoShutdownHook,
		}
		switch strings.ToLower(mode) {
		case "cpu":
			options = append(options, profile.CPUProfile)
		case "mem":
			options = append(options, profile.MemProfile)
		default:
			log.FatalF("unsupported BE_PROFILE_MODE: %v", mode)
		}
		profiler = profile.Start(options...)
	}
}

func Stop() {
	if profiler != nil {
		profiler.Stop()
	}
}
