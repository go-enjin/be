//go:build cloudflare || requests || all

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

package cloudflare

import (
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/net/ip/ranges/cloudflare"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "request-cloudflare"

type Feature interface {
	feature.Feature
	feature.RequestFilter
}

type MakeFeature interface {
	Make() Feature

	AllowDirect() MakeFeature
}

type CFeature struct {
	feature.CFeature

	allowDirect bool
	ipRanges    []string
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

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
}

func (f *CFeature) AllowDirect() MakeFeature {
	f.allowDirect = true
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	if f.ipRanges, err = cloudflare.GetIpRanges(); err != nil {
		log.FatalF("error getting cloudflare ip ranges: %v", err)
	}
	log.DebugF("%v known cloudflare ip ranges", len(f.ipRanges))
	return
}

func (f *CFeature) FilterRequest(r *http.Request) (err error) {
	address, _ := net.GetIpFromRequest(r)
	var ip string
	if ip, err = net.GetProxyIpFromRequest(r); err != nil {
		if f.allowDirect {
			err = nil
			log.DebugRF(r, "cloudflare allowing direct request from: %v", address)
		}
		return
	}
	found := false
	for _, ipRange := range f.ipRanges {
		if net.IsIpInRange(ip, ipRange) {
			found = true
			break
		}
	}
	if !found {
		log.DebugRF(r, "all headers: %v", r.Header)
		err = fmt.Errorf("request denied - not from a known cloudflare ip range: %v (%v)", ip, address)
	}
	return
}