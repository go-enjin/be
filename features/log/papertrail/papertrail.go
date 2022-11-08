//go:build papertrail || all

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

package papertrail

import (
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
)

var _ feature.Feature = (*Feature)(nil)

const Tag feature.Tag = "Papertrail"

type Feature struct {
	feature.CFeature
}

type MakeFeature interface {
	feature.MakeFeature
}

func Make() feature.Feature {
	f := new(Feature)
	f.Init(f)
	return f
}

func (f *Feature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *Feature) Build(b feature.Buildable) (err error) {
	b.AddFlags(
		&cli.StringFlag{
			Name:    "papertrail-host",
			Usage:   "custom papertrail hostname",
			EnvVars: b.MakeEnvKeys("PAPERTRAIL_HOST"),
			Value:   "",
		},
		&cli.IntFlag{
			Name:    "papertrail-port",
			Usage:   "custom papertrail port",
			EnvVars: b.MakeEnvKeys("PAPERTRAIL_PORT"),
			Value:   -1,
		},
	)
	return
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	ptHost := ctx.String("papertrail-host")
	ptPort := ctx.Int("papertrail-port")
	if ptHost == "" || ptPort <= 0 {
		return
	}
	log.Config.LogHook = "papertrail"
	log.Config.PapertrailHost = ptHost
	log.Config.PapertrailPort = ptPort
	log.Config.PapertrailTag = globals.BinName
	log.DebugF("configuring papertrail: %v:%v (tag=%v)", ptHost, ptPort, globals.BinName)
	log.Config.Apply()
	return
}