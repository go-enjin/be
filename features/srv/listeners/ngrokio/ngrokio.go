//go:build srv_listener_ngrokio || ngrokio || srv_listeners || srv || all

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

package ngrokio

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v2"
	"golang.ngrok.com/ngrok"
	ngrokConfig "golang.ngrok.com/ngrok/config"

	//ngrokLog "golang.ngrok.com/ngrok/log"
	ngrokLog "golang.ngrok.com/ngrok/log/logrus"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "srv-listener-ngrokio"

type Feature interface {
	feature.Feature
	feature.ServiceListener
}

type MakeFeature interface {
	Make() Feature

	// WithDomain specifies the ngrok/config.WithDomain setting
	WithDomain(name string) MakeFeature
	// WithRegion specifies the ngrok/config.WithRegion setting
	WithRegion(code string) MakeFeature
	// WithLogging attaches the ngrok tunnel to the enjin logger
	WithLogging(enabled bool) MakeFeature

	// IncludeNgrokEnv includes the "NGROK_AUTHTOKEN" environment key
	IncludeNgrokEnv(enabled bool) MakeFeature
}

type CFeature struct {
	feature.CFeature

	token       string
	region      string
	domain      string
	tunnel      ngrok.Tunnel
	background  context.Context
	logging     bool
	ngrokEnvKey bool
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

func (f *CFeature) IncludeNgrokEnv(enabled bool) MakeFeature {
	f.ngrokEnvKey = enabled
	return f
}

func (f *CFeature) WithDomain(name string) MakeFeature {
	f.domain = name
	return f
}

func (f *CFeature) WithRegion(code string) MakeFeature {
	f.region = code
	return f
}

func (f *CFeature) WithLogging(enabled bool) MakeFeature {
	f.logging = enabled
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) makeCliFlags() (authToken, withDomain, withRegion, withLogging string) {
	authToken = f.KebabTag + "-authtoken"
	withDomain = f.KebabTag + "-domain"
	withRegion = f.KebabTag + "-region"
	withLogging = f.KebabTag + "-logging"
	return
}

func (f *CFeature) makeCliEnvKeys() (authToken, withDomain, withRegion, withLogging string) {
	authToken = f.KebabTag + "_AUTHTOKEN"
	withDomain = f.KebabTag + "_DOMAIN"
	withRegion = f.KebabTag + "_REGION"
	withLogging = f.KebabTag + "_LOGGING"
	return
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}

	authToken, withDomain, withRegion, withLogging := f.makeCliFlags()
	authTokenEnv, withDomainEnv, withRegionEnv, withLoggingEnv := f.makeCliEnvKeys()
	authTokenEnvKeys := b.MakeEnvKeys(authTokenEnv)
	if f.ngrokEnvKey {
		authTokenEnvKeys = append(authTokenEnvKeys, "NGROK_AUTHTOKEN")
	}

	b.AddFlags(
		&cli.StringFlag{
			Name:     authToken,
			Category: f.KebabTag,
			EnvVars:  authTokenEnvKeys,
		},
		&cli.StringFlag{
			Name:     withDomain,
			Category: f.KebabTag,
			EnvVars:  b.MakeEnvKeys(withDomainEnv),
		},
		&cli.StringFlag{
			Name:     withRegion,
			Category: f.KebabTag,
			EnvVars:  b.MakeEnvKeys(withRegionEnv),
		},
		&cli.BoolFlag{
			Name:     withLogging,
			Category: f.KebabTag,
			EnvVars:  b.MakeEnvKeys(withLoggingEnv),
		},
	)
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	authToken, withDomain, withRegion, withLogging := f.makeCliFlags()

	if ctx.IsSet(authToken) {
		f.token = ctx.String(authToken)
	}
	if f.token != "" {
		log.DebugF("ngrok auth-token present")
	} else {
		err = fmt.Errorf("%v feature requires --%s", f.Tag(), authToken)
		return
	}

	if ctx.IsSet(withDomain) {
		f.domain = ctx.String(withDomain)
		log.DebugF("using ngrok-domain: %v", f.domain)
	}

	if ctx.IsSet(withRegion) {
		f.region = ctx.String(withRegion)
		log.DebugF("using ngrok-region: %v", f.region)
	}

	if ctx.IsSet(withLogging) {
		f.logging = ctx.Bool(withLogging)
		log.DebugF("using ngrok-logging: %v", f.logging)
	}

	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) ServiceInfo() (scheme, listen string, port int) {
	port = -1
	if f.tunnel != nil {
		if v := f.tunnel.URL(); v != "" {
			if u, err := url.Parse(v); err == nil {
				scheme = u.Scheme
				if h, p, e := net.SplitHostPort(u.Host); e == nil {
					port, _ = strconv.Atoi(p)
					listen = h
					return
				}
				if u.Scheme == "https" {
					port = 443
				} else if u.Scheme == "http" {
					port = 80
				}
				listen = u.Host
			}
		}
	}
	return
}

func (f *CFeature) StartListening(router *chi.Mux, e feature.EnjinRunner) (err error) {
	e.Notify("ngrok listener starting")

	f.background = context.Background()

	var conOpts []ngrok.ConnectOption
	var tunOpts []ngrokConfig.HTTPEndpointOption

	if f.domain != "" {
		tunOpts = append(tunOpts, ngrokConfig.WithDomain(f.domain))
		log.InfoF("using ngrok.io domain: %v", f.domain)
	}

	conOpts = append(conOpts, ngrok.WithAuthtoken(f.token))

	if f.region != "" {
		conOpts = append(conOpts, ngrok.WithRegion(f.region))
		log.InfoF("using ngrok.io region: %v", f.region)
	}

	if f.logging {
		conOpts = append(conOpts, ngrok.WithLogger(ngrokLog.NewLogger(log.Logrus())))
		log.InfoF("using ngrok.io with logging")
	}

	if f.tunnel, err = ngrok.Listen(f.background, ngrokConfig.HTTPEndpoint(tunOpts...), conOpts...); err != nil {
		return
	}
	log.DebugF("ngrok listener info:\n%v", e.StartupString())

	if err = http.Serve(f.tunnel, router); err != nil {
		if !strings.Contains(err.Error(), "Listener closed") {
			log.ErrorF("unexpected error during ngrok.io tunnel startup: %v", err)
		}
		return
	}

	return
}

func (f *CFeature) StopListening() (err error) {
	if f.tunnel != nil {
		err = f.tunnel.Close()
	}
	return
}
