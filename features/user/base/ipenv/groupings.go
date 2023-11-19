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
	"github.com/go-enjin/be/pkg/maps"
)

type groupings map[string]*grouping

func (g groupings) addNetwork(group string, networks ...*net.IPNet) {
	if _, present := g[group]; !present {
		g[group] = &grouping{name: group}
	}
	g[group].addNetwork(networks...)
}

func (g groupings) addAddress(group string, ips ...net.IP) {
	if _, present := g[group]; !present {
		g[group] = &grouping{name: group}
	}
	g[group].addAddress(ips...)
}

func (g groupings) addActions(group string, actions ...feature.Action) {
	if _, present := g[group]; !present {
		g[group] = &grouping{name: group}
	}
	g[group].addActions(actions...)
}

func (g groupings) String() (text string) {
	text += "{"
	for idx, group := range maps.SortedKeys(g) {
		if idx > 0 {
			text += ","
		}
		text += `"` + group + `":{`

		text += `"networks":[`
		for jdx, network := range g[group].networks {
			if jdx > 0 {
				text += ","
			}
			text += network.String()
		}
		text += "],"

		text += `"addresses":[`
		for jdx, address := range g[group].addresses {
			if jdx > 0 {
				text += ","
			}
			text += address.String()
		}
		text += "],"

		text += `"actions":[` + g[group].actions.String() + `]`
		text += "}"
	}
	text += "}"
	return
}