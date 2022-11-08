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

package cli

import (
	"github.com/urfave/cli/v2"

	beStrings "github.com/go-enjin/be/pkg/strings"
)

func FlagInFlags(name string, flags []cli.Flag) (ok bool) {
	for _, f := range flags {
		if ok = beStrings.StringInSlices(name, f.Names()); ok {
			return
		}
	}
	return
}

func CommandInCommands(name string, commands cli.Commands) (ok bool) {
	for _, c := range commands {
		if ok = beStrings.StringInSlices(name, c.Names()); ok {
			return
		}
	}
	return
}