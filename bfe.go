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
	"fmt"
	"net/http"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/fastcgi"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

type FastcgiEnjin struct {
	port   int
	listen string

	target  string
	source  string
	network string

	prefix string

	production bool
	debug      bool

	feb *FastcgiEnjinBuilder
	cli *cli.App

	handler http.Handler
}

func newFastcgiEnjin(feb *FastcgiEnjinBuilder) (fe *FastcgiEnjin) {
	fe = &FastcgiEnjin{
		feb: feb,
		cli: &cli.App{
			Name:     globals.BinName,
			Usage:    globals.Summary,
			Version:  globals.BuildVersion(),
			Flags:    feb.flags,
			Commands: feb.commands,
		},
	}
	fe.cli.Action = fe.action
	return
}

func (fe *FastcgiEnjin) action(ctx *cli.Context) (err error) {
	log.DebugF("test")
	if fe == nil {
		log.FatalDF(1, "how is the cli context nil?!")
	}

	fe.port = ctx.Int("port")
	fe.listen = ctx.String("listen")

	fe.target = fe.feb.target

	fe.source = fe.feb.source
	if source := ctx.String("fastcgi-source"); source != "" {
		fe.source = source
	}
	if fe.source == "" {
		err = fmt.Errorf("missing --fastcgi-source")
		return
	}
	fe.network = fe.feb.network
	if network := ctx.String("fastcgi-network"); network != "" {
		network = strings.ToLower(network)
		switch network {
		case "tcp", "tcp4", "tcp6", "unix", "auto":
			fe.network = network
		default:
			err = fmt.Errorf("invalid --fastcgi-network")
			return
		}
	}

	fe.debug = ctx.Bool("debug")
	fe.prefix = ctx.String("prefix")
	fe.prefix = strings.ToLower(fe.prefix)
	fe.production = fe.prefix == "" || fe.prefix == "prd"
	if fe.production {
		fe.prefix = ""
		fe.debug = false
	}

	lvl := strings.ToLower(ctx.String("log-level"))
	if v, ok := log.Levels[lvl]; ok {
		log.Config.LogLevel = v
		log.Config.Apply()
	} else {
		if lvl != "" {
			log.FatalF("invalid log-level: %v", lvl)
		}
		if fe.debug {
			log.Config.LogLevel = log.LevelDebug
			log.Config.Apply()
		} else if ctx.Bool("quiet") {
			log.Config.LogLevel = log.LevelWarn
			log.Config.Apply()
		}
	}

	if domains := ctx.StringSlice("domain"); domains != nil && len(domains) > 0 {
		for _, domain := range domains {
			if domain != "" && !beStrings.StringInStrings(domain, fe.feb.domains...) {
				fe.feb.domains = append(fe.feb.domains, domain)
			}
		}
	}

	if fe.handler, err = fastcgi.New(fe.target, fe.network, fe.source); err != nil {
		return
	}
	if err = startupHandledHttpListener(fe.listen, fe.port, fe.handler, fe); err != nil && err.Error() == "http: Server closed" {
		err = nil
	}
	return
}

func (fe *FastcgiEnjin) Shutdown() {
	log.DebugF("fastcgi enjin shutting down")
}