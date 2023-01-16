//go:build fastcgi || all

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

package be

import (
	"os"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

type FastcgiEnjinBuilder struct {
	tag string

	listen string
	port   int

	target  string
	source  string
	network string

	domains     []string
	statusPages map[int]string

	flags    []cli.Flag
	commands cli.Commands
}

func NewFastcgi() (feb *FastcgiEnjinBuilder) {
	feb = &FastcgiEnjinBuilder{
		listen:      globals.DefaultListen,
		port:        globals.DefaultPort,
		target:      "./docroot",
		source:      "",
		network:     "auto",
		statusPages: make(map[int]string),
	}
	return
}

func (feb *FastcgiEnjinBuilder) SetTag(tag string) *FastcgiEnjinBuilder {
	feb.tag = strings.ToLower(tag)
	return feb
}

func (feb *FastcgiEnjinBuilder) SetTarget(target string) *FastcgiEnjinBuilder {
	if bePath.IsFile(target) || bePath.IsDir(target) {
		feb.target = target
	} else {
		log.FatalDF(1, "not a file or directory: %v", target)
	}
	return feb
}

func (feb *FastcgiEnjinBuilder) SetSource(source string) *FastcgiEnjinBuilder {
	feb.source = source
	return feb
}

func (feb *FastcgiEnjinBuilder) SetNetwork(network string) *FastcgiEnjinBuilder {
	switch strings.ToLower(network) {
	case "tcp", "tcp4", "tcp6", "unix", "auto":
		feb.network = network
	default:
		log.FatalDF(1, "invalid network type: %v", network)
	}
	return feb
}

func (feb *FastcgiEnjinBuilder) AddStatusPage(status int, redirect string) *FastcgiEnjinBuilder {
	feb.statusPages[status] = redirect
	return feb
}

func (feb *FastcgiEnjinBuilder) AddFlags(flags ...cli.Flag) *FastcgiEnjinBuilder {
	feb.flags = append(feb.flags, flags...)
	return feb
}

func (feb *FastcgiEnjinBuilder) AddCommands(commands ...*cli.Command) *FastcgiEnjinBuilder {
	feb.commands = append(feb.commands, commands...)
	return feb
}

func (feb *FastcgiEnjinBuilder) AddDomains(domains ...string) *FastcgiEnjinBuilder {
	for _, domain := range domains {
		if domain != "" && !beStrings.StringInSlices(domain, feb.domains) {
			feb.domains = append(feb.domains, domain)
		}
	}
	return feb
}

func (feb *FastcgiEnjinBuilder) Build() (runner feature.Runner) {
	if feb.tag == "" {
		log.FatalDF(1, "missing .SetTag")
	}

	feb.flags = append(
		feb.flags,
		&cli.StringFlag{
			Name:    "listen",
			Usage:   "the address to listen on",
			Value:   globals.DefaultListen,
			Aliases: []string{"L"},
			EnvVars: globals.MakeEnvKeys("LISTEN"),
		},
		&cli.IntFlag{
			Name:    "port",
			Usage:   "the port to listen on",
			Value:   globals.DefaultPort,
			Aliases: []string{"p"},
			EnvVars: append(globals.MakeEnvKeys("PORT"), "PORT"),
		},
		&cli.StringFlag{
			Name:    "fastcgi-source",
			Usage:   "path to local unix socket or an address:port",
			EnvVars: globals.MakeEnvKeys("FCGI_SRC"),
		},
		&cli.StringFlag{
			Name:    "fastcgi-network",
			Usage:   "specify fast cgi network type: tcp, tcp4, tcp6, unix or auto",
			Value:   "auto",
			EnvVars: globals.MakeEnvKeys("FCGI_NET"),
		},
		&cli.StringFlag{
			Name:    "prefix",
			Usage:   "for dev and stg sites to prefix labels",
			Value:   os.Getenv("USER"),
			Aliases: []string{"P"},
			EnvVars: globals.MakeEnvKeys("PREFIX"),
		},
		&cli.BoolFlag{
			Name:    "quiet",
			Usage:   "set log level to WARN",
			Aliases: []string{"q"},
			EnvVars: globals.MakeEnvKeys("QUIET"),
		},
		&cli.BoolFlag{
			Name:    "debug",
			Usage:   "enable verbose logging for debugging purposes",
			EnvVars: globals.MakeEnvKeys("DEBUG"),
		},
		&cli.StringFlag{
			Name:    "log-level",
			Usage:   "set logging level: error, warn, info, debug or trace",
			EnvVars: globals.MakeEnvKeys("LOG_LEVEL"),
		},
		&cli.StringSliceFlag{
			Name:    "domain",
			Usage:   "restrict inbound requests to only the domain names given",
			EnvVars: globals.MakeEnvKeys("DOMAIN"),
		},
		&cli.BoolFlag{
			Name:    "strict",
			Usage:   "use strict Slugsums validation (extraneous files are errors)",
			Aliases: []string{"s"},
			EnvVars: globals.MakeEnvKeys("STRICT"),
		},
		&cli.StringFlag{
			Name:    "sums-integrity",
			Usage:   "specify the sha256sum of the Shasums file for --strict validations",
			Value:   "",
			EnvVars: globals.MakeEnvKeys("SUMS_INTEGRITY_" + strcase.ToScreamingSnake(globals.BinHash)),
		},
	)

	runner = newFastcgiEnjin(feb)
	return
}