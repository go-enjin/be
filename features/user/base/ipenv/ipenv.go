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

package ipenv

import (
	"net"
	"net/http"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	site_environ "github.com/go-enjin/be/pkg/feature/site-environ"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	beNet "github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/userbase"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "user-base-ipenv"

type Feature interface {
	feature.Feature
	feature.RequestRewriter
}

type MakeFeature interface {
	AddAddresses(group string, cidrs ...string) MakeFeature
	AddPermissions(group string, actions ...feature.Action) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	env *site_environ.CSiteEnviron[MakeFeature]

	groupings groupings
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

func (f *CFeature) UsageNotes() (notes []string) {
	//tag := f.Tag().String()
	//notes = []string{
	//	"this feature looks for dynamic environment variable keys",
	//	"groups are defined by their permissions and one or more IP addresses or CIDR ranges",
	//	"in the following examples, \"PRIVATE_USERS\" is the group \"private-users\"",
	//	"add networks: " + globals.MakeFlagEnvKey(tag, "NETWORKS_PRIVATE_USERS") + "='<space separated CIDR>'",
	//	"add addresses: " + globals.MakeFlagEnvKey(tag, "ADDRESSES_PRIVATE_USERS") + "='<space separated IP>'",
	//	"add permissions: " + globals.MakeFlagEnvKey(tag, "PERMISSIONS_PRIVATE_USERS") + "='<space separated actions>'",
	//}
	notes = f.env.SiteEnvironUsageNotes()
	return
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.env = site_environ.New[MakeFeature](this,
		"networks", "space separated CIDR ranges",
		"addresses", "space separated IP addresses",
		"permissions", "space separated actions",
	)
	f.groupings = make(groupings)
	return
}

func (f *CFeature) AddNetworks(group string, cidrs ...string) MakeFeature {
	if len(cidrs) > 0 {
		group = strcase.ToKebab(group)
		for _, cidr := range cidrs {
			if _, network, err := net.ParseCIDR(cidr); err != nil {
				log.FatalDF(1, "error parsing %q CIDR: %v", err)
			} else {
				f.groupings.addNetwork(group, network)
			}
		}
	}
	return f
}

func (f *CFeature) AddAddresses(group string, ips ...string) MakeFeature {
	if len(ips) > 0 {
		group = strcase.ToKebab(group)
		for _, ip := range ips {
			if parsed := net.ParseIP(ip); parsed == nil {
				log.FatalDF(1, "%q is not an IP address", ip)
			} else {
				f.groupings.addAddress(group, parsed)
			}
		}
	}
	return f
}

func (f *CFeature) AddPermissions(group string, actions ...feature.Action) MakeFeature {
	if len(actions) > 0 {
		group = strcase.ToKebab(group)
		f.groupings.addActions(group, actions...)
	}
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	} else if err = f.env.StartupSiteEnviron(); err != nil {
		return
	}
	if err = f.loadEnvironment(); err != nil {
		return
	}

	log.InfoF("%v feature has %d IP permission groupings: %v", f.Tag(), len(f.groupings), f.groupings)
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) RewriteRequest(w http.ResponseWriter, r *http.Request) (modified *http.Request) {
	if address, err := beNet.ParseIpFromRequest(r); err == nil {
		modified = r
		for _, name := range maps.SortedKeys(f.groupings) {
			if group := f.groupings[name]; group != nil {
				if group.matches(address) {
					log.DebugF("[%v] current address (%v) matches %q group; adding permissions: %v", f.Tag(), address, name, group.actions)
					modified = userbase.AppendCurrentPermissions(modified, group.actions...)
				}
			}
		}
	}
	return
}
