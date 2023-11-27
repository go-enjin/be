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
	beNet "github.com/go-enjin/be/pkg/net"
)

func (f *CFeature) loadEnvironment() (err error) {

	for _, key := range []string{"networks", "addresses", "permissions"} {
		if named, ok := f.env.GetSiteEnviron(key); ok {
			for group, value := range named {
				switch key {

				case "networks":
					var networks []*net.IPNet
					if networks, err = beNet.ParseCIDR(value); err != nil {
						return
					}
					f.groupings.addNetwork(group, networks...)

				case "addresses":
					var addresses []net.IP
					if addresses, err = beNet.ParseIP(value); err != nil {
						return
					}
					f.groupings.addAddress(group, addresses...)

				case "permissions":
					actions := feature.ParseActions(value)
					f.groupings.addActions(group, actions...)
				}
			}

		}
	}

	return
}
