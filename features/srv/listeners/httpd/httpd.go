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
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
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
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *CFeature) Shutdown() {

}

func (f *CFeature) StopListening() (err error) {
	if f.srv != nil {
		err = f.srv.Shutdown(context.Background())
	}
	return
}

func (f *CFeature) StartListening(listen string, port int, router *chi.Mux, e feature.EnjinRunner) (err error) {
	e.Notify("http listener starting")
	log.DebugF("http listener info:\n%v", e.StartupString())

	// TODO: implement signal handler features for http listeners
	f.srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", listen, port),
		Handler: router,
	}

	idleConnectionsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint
		if f.srv != nil {
			if ee := f.srv.Shutdown(context.Background()); ee != nil {
				log.ErrorF("error shutting down http listener: %v", ee)
			}
			f.srv = nil
		}
		e.Shutdown()
		close(idleConnectionsClosed)
	}()

	if err = f.srv.ListenAndServe(); err != http.ErrServerClosed {
		log.ErrorF("unexpected error during http listener startup: %v", err)
		close(idleConnectionsClosed)
		return
	}

	<-idleConnectionsClosed
	return
}