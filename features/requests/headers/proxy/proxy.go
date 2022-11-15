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

	"github.com/go-chi/chi/v5/middleware"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
)

var (
	_ Feature            = (*CFeature)(nil)
	_ MakeFeature        = (*CFeature)(nil)
	_ feature.Middleware = (*CFeature)(nil)
)

const Tag feature.Tag = "HeaderProxy"

type Feature interface {
	feature.Middleware
}

type CFeature struct {
	feature.CMiddleware

	enabled bool
}

type MakeFeature interface {
	Enable() MakeFeature

	Make() Feature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	return f
}

// Enable sets the default state to enabled, overridden by --header-proxy
func (f *CFeature) Enable() MakeFeature {
	f.enabled = true
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	b.AddFlags(&cli.BoolFlag{
		Name:    "header-proxy",
		Usage:   "rewrite req.RemoteAddr to take proxy headers into account",
		EnvVars: b.MakeEnvKeys("HEADER_PROXY"),
	})
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if ctx.IsSet("header-proxy") {
		f.enabled = ctx.Bool("header-proxy")
	}
	return
}

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	return func(next http.Handler) (this http.Handler) {
		if f.enabled {
			return middleware.RealIP(next)
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			return
		})
	}
}