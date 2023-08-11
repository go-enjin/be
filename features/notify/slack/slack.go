//go:build notify_slack || all

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

package slack

import (
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/notify"
	"github.com/go-enjin/be/pkg/slices"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "notify-slack"

type Feature interface {
	feature.Feature
}

type MakeFeature interface {
	Make() Feature

	Add(channel string) MakeFeature
}

type CFeature struct {
	feature.CFeature

	channels []string
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = Tag
	return f
}

func (f *CFeature) Add(channel string) MakeFeature {
	if !slices.Present(channel, f.channels...) {
		f.channels = append(f.channels, channel)
	}
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	b.AddFlags(&cli.StringFlag{
		Name:     "slack",
		Usage:    "the unique part of a slack channel webhook URL",
		Aliases:  []string{"S"},
		EnvVars:  b.MakeEnvKeys("SLACK"),
		Value:    "",
		Category: f.Tag().String(),
	})
	b.AddNotifyHook("slack", f.notifyHook)
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	if channel := ctx.String("slack"); channel != "" {
		f.Add(channel)
	}
	return
}

func (f *CFeature) notifyHook(format string, argv ...interface{}) {
	for _, channel := range f.channels {
		if err := notify.SlackF(channel, format, argv...); err != nil {
			log.ErrorF("error notifying slack: %v", err)
		}
	}
}