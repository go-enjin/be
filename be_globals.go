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

package be

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func init() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print only the version",
	}
	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h", "usage"},
		Usage:   "display helpful information",
	}
	if globals.BinName == "" {
		globals.BinName = filepath.Base(os.Args[0])
	}
	if globals.EnvPrefix == "" {
		globals.EnvPrefix = strcase.ToScreamingSnake(globals.BinName)
	} else {
		globals.EnvPrefix = strcase.ToScreamingSnake(globals.EnvPrefix)
	}
	if globals.BinHash == "" {
		var err error
		var absPath string
		if absPath, err = filepath.Abs(os.Args[0]); err != nil {
			log.ErrorF("absolute path error %v: %v", os.Args[0], err)
		} else if globals.BinHash, err = sha.FileHash10(absPath); err != nil {
			log.ErrorF("sha hashing error %v: %v", os.Args[0], err)
			globals.BinHash = "0000000000"
		}
	}
	if globals.Version == "" {
		globals.Version = "v0.0.0"
	}
	if globals.Summary == "" {
		globals.Summary = "A back enjin service."
	}
	log.Config.AppName = globals.BinName

	earlyDebug := false
	check := strings.ToLower(os.Getenv(globals.EnvPrefix + "_DEBUG"))
	if check == "true" {
		earlyDebug = true
	} else {
		check = strings.ToLower(os.Getenv("DEBUG"))
		if check == "true" {
			earlyDebug = true
		} else {
			if idx := beStrings.StringIndexInStrings("--debug", os.Args...); idx >= 0 {
				if len(os.Args) > idx+1 {
					if strings.ToLower(os.Args[idx+1]) == "true" {
						earlyDebug = true
					}
				} else {
					earlyDebug = true
				}
			} else if beStrings.StringInStrings("--debug=true", os.Args...) {
				earlyDebug = true
			}
		}
	}
	if earlyDebug {
		log.Config.LogLevel = log.LevelDebug
	}

	log.Config.Apply()
}