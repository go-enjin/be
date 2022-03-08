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
	"github.com/go-enjin/be/pkg/strings"
)

var _notifySlack *Feature

var _ feature.Feature = (*Feature)(nil)

const Tag feature.Tag = "NotifySlack"

type Feature struct {
	feature.CFeature

	channels []string
}

type MakeFeature interface {
	feature.MakeFeature

	Add(channel string) MakeFeature
}

func New() MakeFeature {
	if _notifySlack == nil {
		_notifySlack = new(Feature)
		_notifySlack.Init(_notifySlack)
	}
	return _notifySlack
}

func (f *Feature) Add(channel string) MakeFeature {
	if !strings.StringInStrings(channel, f.channels...) {
		f.channels = append(f.channels, channel)
	}
	return f
}

func (f *Feature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *Feature) Build(b feature.Buildable) (err error) {
	b.AddFlags(&cli.StringFlag{
		Name:    "slack",
		Usage:   "the unique part of a slack channel webhook URL",
		Aliases: []string{"S"},
		EnvVars: b.MakeEnvKeys("SLACK"),
		Value:   "",
	})
	b.AddNotifyHook("slack", f.notifyHook)
	return
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	if channel := ctx.String("slack"); channel != "" {
		f.Add(channel)
	}
	return
}

func (f *Feature) notifyHook(format string, argv ...interface{}) {
	for _, channel := range f.channels {
		if err := notify.SlackF(channel, format, argv...); err != nil {
			log.ErrorF("error notifying slack: %v", err)
		}
	}
}