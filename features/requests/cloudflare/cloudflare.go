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
	"github.com/go-enjin/be/pkg/net/ip/ranges/cloudflare"
	"github.com/go-enjin/be/pkg/utils"
)

var _cloudflare *Feature

var _ feature.Feature = (*Feature)(nil)

var _ feature.RequestFilter = (*Feature)(nil)

const Tag feature.Tag = "RequestCloudflare"

type Feature struct {
	feature.CFeature

	allowDirect bool
	ipRanges    []string
}

type MakeFeature interface {
	feature.MakeFeature

	AllowDirect() MakeFeature
}

func New() MakeFeature {
	if _cloudflare == nil {
		_cloudflare = new(Feature)
		_cloudflare.Init(_cloudflare)
	}
	return _cloudflare
}

func (f *Feature) AllowDirect() MakeFeature {
	f.allowDirect = true
	return f
}

func (f *Feature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *Feature) Build(b feature.Buildable) (err error) {
	return
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	if f.ipRanges, err = cloudflare.GetIpRanges(); err != nil {
		log.FatalF("error getting cloudflare ip ranges: %v", err)
	}
	log.DebugF("%v known cloudflare ip ranges", len(f.ipRanges))
	return
}

func (f *Feature) FilterRequest(r *http.Request) (err error) {
	address, _ := utils.GetIpFromRequest(r)
	var ip string
	if ip, err = utils.GetProxyIpFromRequest(r); err != nil {
		if f.allowDirect {
			err = nil
			log.DebugF("cloudflare allowing direct request from: %v", address)
		}
		return
	}
	found := false
	for _, ipRange := range f.ipRanges {
		if utils.IsIpInRange(ip, ipRange) {
			found = true
			break
		}
	}
	if !found {
		log.DebugF("all headers: %v", r.Header)
		err = fmt.Errorf("request denied - not from a known cloudflare ip range: %v (%v)", ip, address)
	}
	return
}