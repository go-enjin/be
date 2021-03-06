//go:build header_proxy || all

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

package proxy

import (
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
)

var _headerProxy *Feature

var (
	_ feature.Feature         = (*Feature)(nil)
	_ feature.RequestModifier = (*Feature)(nil)
)

const Tag feature.Tag = "HeaderProxy"

type Feature struct {
	feature.CFeature

	enabled bool
}

type MakeFeature interface {
	feature.MakeFeature

	Enable() MakeFeature
}

func New() MakeFeature {
	if _headerProxy == nil {
		_headerProxy = new(Feature)
		_headerProxy.Init(_headerProxy)
	}
	return _headerProxy
}

func (f *Feature) Enable() MakeFeature {
	f.enabled = true
	return f
}

func (f *Feature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *Feature) Build(b feature.Buildable) (err error) {
	b.AddFlags(&cli.BoolFlag{
		Name:    "header-proxy",
		Usage:   "rewrite req.RemoteAddr to take proxy headers into account",
		EnvVars: b.MakeEnvKeys("HEADER_PROXY"),
	})
	return
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	if ctx.IsSet("header-proxy") {
		f.enabled = ctx.Bool("header-proxy")
	}
	log.InfoF("header-proxy enabled: %v", f.enabled)
	return
}

func (f *Feature) ModifyRequest(w http.ResponseWriter, r *http.Request) {
	if ip, err := net.GetIpFromRequest(r); err == nil {
		if ip != r.RemoteAddr {
			log.DebugF("setting RemoteAddr to %v (was: %v)", ip, r.RemoteAddr)
			r.RemoteAddr = ip
		}
	} else {
		log.ErrorF("error getting ip from request: %v", err)
	}
}