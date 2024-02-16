//go:build srv_listener_httpd || srv_listeners || srv || all

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

package httpd

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "srv-listener-httpd"

type Feature interface {
	feature.Feature
	feature.ServiceListener
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature

	port   int
	listen string

	srv *http.Server
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.CFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	category := f.Tag().String()
	b.AddFlags(
		&cli.StringFlag{
			Name:     "listen",
			Usage:    "the address to listen on",
			Value:    globals.DefaultListen,
			Aliases:  []string{"L"},
			EnvVars:  b.MakeEnvKeys("LISTEN"),
			Category: category,
		},
		&cli.IntFlag{
			Name:     "port",
			Usage:    "the port to listen on",
			Value:    globals.DefaultPort,
			Aliases:  []string{"p"},
			EnvVars:  append(b.MakeEnvKeys("PORT"), "PORT"),
			Category: category,
		},
	)
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	f.port = ctx.Int("port")
	f.listen = ctx.String("listen")
	return
}

func (f *CFeature) ServiceInfo() (scheme, listen string, port int) {
	port = f.port
	listen = f.listen
	scheme = "http"
	return
}

func (f *CFeature) StopListening() (err error) {
	if f.srv != nil {
		err = f.srv.Shutdown(context.Background())
	}
	return
}

func (f *CFeature) StartListening(router *chi.Mux, e feature.EnjinRunner) (err error) {
	e.Notify("http listener starting")
	log.DebugF("http listener info:\n%v", e.StartupString())

	f.srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", f.listen, f.port),
		Handler: router,
	}

	if err = f.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.ErrorF("unexpected error during http listener startup/shutdown: %v", err)
		return
	}
	return
}
