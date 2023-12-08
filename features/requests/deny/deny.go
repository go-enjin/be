//go:build requests_deny || requests || all

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

package deny

import (
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
)

// TODO: validate .Block(addresses)
// TODO: support CIDR blocking

const Tag feature.Tag = "requests-deny"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type MakeFeature interface {
	Make() Feature

	SetDenyDuration(seconds int64) MakeFeature

	Block(address string) MakeFeature
	Restrict(path string) MakeFeature

	Defaults() MakeFeature

	RestrictEnv() MakeFeature
	RestrictGit() MakeFeature
	RestrictWordPress() MakeFeature
}

type Feature interface {
	feature.Feature
	feature.UseMiddleware
}

type CFeature struct {
	feature.CFeature

	manager *manager
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
	f.manager = newManager(DefaultDuration)
}

func (f *CFeature) SetDenyDuration(seconds int64) MakeFeature {
	f.manager.SetPeriod(seconds)
	return f
}

func (f *CFeature) Block(address string) MakeFeature {
	f.manager.Block(address)
	return f
}

func (f *CFeature) Restrict(path string) MakeFeature {
	if err := f.manager.Restrict(path); err != nil {
		log.FatalDF(1, "error setting path restriction: %v - %v", path, err)
	}
	return f
}

func (f *CFeature) RestrictEnv() MakeFeature {
	_ = f.manager.Restrict(`/\.env\b`)
	return f
}

func (f *CFeature) RestrictGit() MakeFeature {
	_ = f.manager.Restrict(`/\.git\b/?`)
	return f
}

func (f *CFeature) RestrictWordPress() MakeFeature {
	_ = f.manager.Restrict(`/wp-(admin|login\.php|login|includes|content)`)
	_ = f.manager.Restrict(`/xmlrpc\.php`)
	return f
}

func (f *CFeature) Defaults() MakeFeature {
	f.RestrictEnv()
	f.RestrictGit()
	f.RestrictWordPress()
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	envPrefix := f.Tag().ScreamingSnake()
	b.AddFlags(
		&cli.Int64Flag{
			Name:     f.KebabTag + "-deny-duration",
			Usage:    "number of seconds to block denied ip addresses",
			EnvVars:  b.MakeEnvKeys(envPrefix, "DENY_DURATION"),
			Value:    DefaultDuration,
			Category: f.KebabTag,
		},
		&cli.StringSliceFlag{
			Name:     f.KebabTag + "-deny-addresses",
			Usage:    "space separated list of IP addresses to always block",
			EnvVars:  b.MakeEnvKeys(envPrefix, "DENY_ADDRESSES"),
			Category: f.KebabTag,
		},
	)
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	err = f.CFeature.Startup(ctx)

	if denyDurationKey := f.KebabTag + "-deny-duration"; ctx.IsSet(denyDurationKey) {
		duration := ctx.Int64(denyDurationKey)
		f.manager.SetPeriod(duration)
		log.DebugF("%v - deny duration set to: %v", f.Tag(), duration)
	}

	if denyAddressesKey := f.KebabTag + "-deny-addresses"; ctx.IsSet(denyAddressesKey) {
		addresses := ctx.StringSlice(denyAddressesKey)
		for _, address := range addresses {
			f.manager.Block(address)
		}
		log.DebugF("%v - deny addresses set to: %+v", f.Tag(), addresses)
	}

	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) Use(s feature.System) (fn feature.MiddlewareFn) {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if address, denied := f.CheckRequestDenied(r); denied {
				log.DebugF(`request denied: "%v" (%v)`, address, r.URL.Path)
				f.Enjin.Serve404(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (f *CFeature) CheckRequestDenied(req *http.Request) (address string, denied bool) {
	var err error
	var addr string
	if addr, err = net.GetIpFromRequest(req); err != nil {
		return err.Error(), true
	}
	if f.manager.Denied(addr) {
		return addr, true
	} else if f.manager.Restricted(req.URL.Path) {
		f.manager.Deny(addr)
		return addr, true
	}
	return addr, false
}