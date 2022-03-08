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
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/headers"
	"github.com/go-enjin/be/pkg/net/ip/deny"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/theme"
)

var _ feature.Builder = &EnjinBuilder{}

type EnjinBuilder struct {
	flags        []cli.Flag
	pages        map[string]*page.Page
	context      context.Context
	theme        string
	theming      map[string]*theme.Theme
	features     []feature.Feature
	headers      []headers.ModifyHeadersFn
	domains      []string
	processors   map[string]feature.ReqProcessFn
	translators  map[string]feature.TranslateOutputFn
	transformers map[string]feature.TransformOutputFn
	slugsums     bool
}

func New() (be *EnjinBuilder) {
	be = new(EnjinBuilder)
	be.theme = globals.DefaultTheme
	be.flags = make([]cli.Flag, 0)
	be.pages = make(map[string]*page.Page)
	be.context = context.New()
	be.theming = make(map[string]*theme.Theme)
	be.features = make([]feature.Feature, 0)
	be.headers = make([]headers.ModifyHeadersFn, 0)
	be.domains = make([]string, 0)
	be.processors = make(map[string]feature.ReqProcessFn)
	be.translators = make(map[string]feature.TranslateOutputFn)
	be.transformers = make(map[string]feature.TransformOutputFn)
	be.slugsums = true
	return be
}

func (eb *EnjinBuilder) IgnoreSlugsums() *EnjinBuilder {
	eb.slugsums = false
	return eb
}

func (eb *EnjinBuilder) ExcludeDefaultTheme() *EnjinBuilder {
	eb.context.Set("ExcludeDefaultTheme", true)
	return eb
}

func (eb *EnjinBuilder) Build() feature.Runner {
	if err := eb.resolveFeatureDeps(); err != nil {
		log.FatalF("error resolving feature dependencies: %v", err)
		return nil
	}

	if !eb.context.Bool("ExcludeDefaultTheme", false) {
		if t, err := theme.NewDefault(); err != nil {
			log.FatalF("error enabling default theme: %v", err)
		} else {
			eb.theming["_default"] = t
		}
	}
	delete(eb.context, "ExcludeDefaultTheme")

	if eb.theme == "" {
		if len(eb.theming) == 1 {
			if _, ok := eb.theming["_default"]; ok {
				eb.theme = "_default"
			} else {
				for k, _ := range eb.theming {
					eb.theme = k
					break
				}
			}
		}
	}

	for _, f := range eb.features {
		if err := f.Self().Build(eb); err != nil {
			log.FatalF("%v", err)
			return nil
		}
	}

	eb.flags = append(
		eb.flags,
		&cli.StringFlag{
			Name:    "listen",
			Usage:   "the address to listen on",
			Value:   globals.DefaultListen,
			Aliases: []string{"L"},
			EnvVars: eb.MakeEnvKeys("LISTEN"),
		},
		&cli.IntFlag{
			Name:    "port",
			Usage:   "the port to listen on",
			Value:   globals.DefaultPort,
			Aliases: []string{"p"},
			EnvVars: append(eb.MakeEnvKeys("PORT"), "PORT"),
		},
		&cli.StringFlag{
			Name:    "prefix",
			Usage:   "for dev and stg sites to prefix labels",
			Value:   os.Getenv("USER"),
			Aliases: []string{"P"},
			EnvVars: eb.MakeEnvKeys("PREFIX"),
		},
		&cli.BoolFlag{
			Name:    "quiet",
			Usage:   "set log level to WARN",
			Aliases: []string{"q"},
			EnvVars: eb.MakeEnvKeys("QUIET"),
		},
		&cli.BoolFlag{
			Name:    "debug",
			Usage:   "enable verbose logging for debugging purposes",
			Aliases: []string{"D"},
			EnvVars: eb.MakeEnvKeys("DEBUG"),
		},
		&cli.Int64Flag{
			Name:    "deny-duration",
			Usage:   "number of seconds to block denied ip addresses",
			EnvVars: eb.MakeEnvKeys("DENY_DURATION"),
			Value:   deny.DenyDuration,
		},
		&cli.StringSliceFlag{
			Name:    "domain",
			Usage:   "restrict inbound requests to only the domain names given",
			EnvVars: eb.MakeEnvKeys("DOMAIN"),
		},
		&cli.BoolFlag{
			Name:    "strict",
			Usage:   "use strict Slugsums validation (extraneous files are errors)",
			Aliases: []string{"s"},
			EnvVars: eb.MakeEnvKeys("STRICT"),
		},
		&cli.StringFlag{
			Name:    "sums-integrity",
			Usage:   "specify the sha256sum of the Shasums file for --strict validations",
			EnvVars: eb.MakeEnvKeys("SUMS_INTEGRITY_" + strings.ToUpper(globals.BinHash)),
		},
	)
	return newEnjin(eb)
}