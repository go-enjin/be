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

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "log-papertrail"

type Feature interface {
	feature.Feature
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
}

func Make() Feature {
	return New().Make()
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	category := f.Tag().String()
	b.AddFlags(
		&cli.StringFlag{
			Name:     "papertrail-host",
			Usage:    "custom papertrail hostname",
			EnvVars:  b.MakeEnvKeys("PAPERTRAIL_HOST"),
			Value:    "",
			Category: category,
		},
		&cli.IntFlag{
			Name:     "papertrail-port",
			Usage:    "custom papertrail port",
			EnvVars:  b.MakeEnvKeys("PAPERTRAIL_PORT"),
			Value:    -1,
			Category: category,
		},
	)
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
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