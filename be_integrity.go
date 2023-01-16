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
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/slug"
)

func (e *Enjin) startupIntegrityChecks(ctx *cli.Context) (err error) {
	eicPrefix := "enjin integrity checks"
	eicLogMsg := func(status, format string, argv ...interface{}) (msg string) {
		msgFmt := fmt.Sprintf("%v %v: %v", eicPrefix, status, format)
		msg = fmt.Sprintf(msgFmt, argv...)
		return
	}
	eicFmtErr := func(format string, argv ...interface{}) (e error) {
		e = fmt.Errorf(eicLogMsg("failed", format, argv...))
		return
	}
	if e.eb.slugsums {
		if slug.SlugsumsPresent() {
			var slugMap slug.ShaMap
			var imposters, extraneous, validated []string
			if slugMap, _, imposters, extraneous, validated, err = slug.ValidateSlugsumsComplete(); err != nil {
				log.ErrorF(eicLogMsg("failed", err.Error()))
				return
			}
			il := len(imposters)
			el := len(extraneous)
			if il > 0 || (ctx.Bool("strict") && el > 0) {
				if il > 0 {
					if el > 0 {
						err = eicFmtErr("imposters: %v, extraneous: %v", imposters, extraneous)
						log.ErrorF(eicLogMsg("failed", "summary:\n\timposter files present: %v\n\textraneous files present: %v", imposters, extraneous))
					} else {
						err = eicFmtErr("imposters: %v", imposters)
						log.ErrorF(eicLogMsg("failed", "summary:\n\timposter files present: %v", imposters))
					}
				} else {
					err = eicFmtErr("extraneous: %v", extraneous)
					log.ErrorF(eicLogMsg("failed", "summary:\n\textraneous files present: %v", extraneous))
				}
				e.NotifyF("failed", err.Error())
				return
			}
			if ctx.Bool("strict") {
				if err = slugMap.CheckSlugIntegrity(); err != nil {
					err = eicFmtErr(err.Error())
					return
				}
				log.InfoF(eicLogMsg("partial-pass", "slug integrity validated successfully"))
				if ctx.IsSet("sums-integrity") {
					globals.SumsIntegrity = strings.ToLower(ctx.String("sums-integrity"))
					if !sha.RxShasum64.MatchString(globals.SumsIntegrity) {
						err = eicFmtErr("invalid --sums-integrity value, must be 64 characters of [a-f0-9]")
						return
					}
				} else {
					err = eicFmtErr("missing --sums-integrity value (--strict present)")
					return
				}
				if err = slugMap.CheckSumsIntegrity(); err != nil {
					err = eicFmtErr(err.Error())
					return
				}
				log.InfoF(eicLogMsg("partial-pass", "sums integrity validated successfully"))
			} else if sums, ee := slugMap.SumsIntegrity(); ee == nil {
				envKey := globals.MakeEnvKey("SUMS_INTEGRITY_" + strings.ToUpper(globals.BinHash))
				log.InfoF("export %v=\"%v\"", envKey, sums)
			}
			vl := len(validated)
			if el > 0 {
				log.WarnF(eicLogMsg("passed", "ignoring extraneous files present: %v", extraneous))
				e.NotifyF(eicPrefix+" passed", "successfully validated %d files (ignoring %d extraneous)", vl, el)
			} else {
				e.NotifyF(eicPrefix+" passed", "successfully validated %d files", vl)
			}
		} else if ctx.Bool("strict") {
			err = eicFmtErr("missing Slugsums file (--strict present)")
			e.NotifyF(eicPrefix+" failed", "missing Slugsums file (--strict present)")
			return
		} else {
			log.DebugF("Slugsums not found, enjin integrity checks skipped")
		}
	}

	return
}