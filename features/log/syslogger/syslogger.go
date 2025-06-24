//go:build log_syslogger || syslogger || loggers || all

// Copyright (c) 2025  The Go-Enjin Authors
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

package syslogger

import (
	"fmt"

	syslog "github.com/dmachard/go-clientsyslog"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "log-syslogger"

type Feature interface {
	feature.Feature
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature

	network string
	address string
	port    int
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
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.Construct(f)
	return f
}

func (f *CFeature) SetNetwork(network string) MakeFeature {

	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	category := f.Tag().String()
	snake := f.Tag().Snake()
	b.AddFlags(
		&cli.StringFlag{
			Name:     "syslogger-dsn",
			Usage:    "custom syslog DSN",
			EnvVars:  b.MakeEnvKeys(snake + "_DSN"),
			Value:    "",
			Category: category,
		},
	)
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	if dsnArg := ctx.String("syslogger-dsn"); dsnArg != "" {
		// empty dsn means do nothing

		var dsn *log.SyslogDSN
		if dsn, err = log.ParseSyslogDSN(dsnArg); err != nil {
			return
		}

		var hook logrus.Hook
		if hook, err = log.NewSyslogHook(dsn, syslog.LOG_INFO, log.Config.AppName); err != nil {
			return fmt.Errorf("failed to setup syslogger (%v): %w", dsn, err)
		} else {
			log.Config.LogHooks = append(log.Config.LogHooks, hook)
		}

		if dsn.Options.Has("Format") {
			if lf, ok := dsn.Options.Get("Format").(log.Format); ok {
				log.Config.LoggingFormat = lf
			}
		}

		log.Config.Apply()
		log.DebugF("configured syslogger: %v", dsn)
	}
	return
}
