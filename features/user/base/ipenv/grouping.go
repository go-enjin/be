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

	"github.com/go-enjin/be/pkg/feature"
)

type grouping struct {
	name      string
	addresses []net.IP
	networks  []*net.IPNet
	actions   feature.Actions
}

func (g *grouping) matches(ip net.IP) (matches bool) {
	for _, address := range g.addresses {
		if matches = address.Equal(ip); matches {
			return
		}
	}
	for _, network := range g.networks {
		if matches = network.Contains(ip); matches {
			return
		}
	}
	return
}

func (g *grouping) addAddress(addresses ...net.IP) {
	g.addresses = append(g.addresses, addresses...)
}

func (g *grouping) addNetwork(networks ...*net.IPNet) {
	g.networks = append(g.networks, networks...)
}

func (g *grouping) addActions(actions ...feature.Action) {
	for _, action := range actions {
		if g.actions.Has(action) {
			continue
		}
		g.actions = append(g.actions, action)
	}
}
