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

// MustParseCIDR and the list of PrivateNetworks borrowed from:
//   https://github.com/wader/filtertransport/blob/master/filter.go
//   Thanks Mattias Wadman!

package net

import "net"

// MustParseCIDR parses string into net.IPNet
func MustParseCIDR(s string) net.IPNet {
	if _, ipnet, err := net.ParseCIDR(s); err != nil {
		panic(err)
	} else {
		return *ipnet
	}
}

// PrivateNetworks net.IPNets that are loopback, private, link local, default unicast
// based on https://github.com/letsencrypt/boulder/blob/master/bdns/dns.go
var PrivateNetworks = []net.IPNet{
	MustParseCIDR("10.0.0.0/8"),         // RFC1918
	MustParseCIDR("172.16.0.0/12"),      // private
	MustParseCIDR("192.168.0.0/16"),     // private
	MustParseCIDR("127.0.0.0/8"),        // RFC5735
	MustParseCIDR("0.0.0.0/8"),          // RFC1122 Section 3.2.1.3
	MustParseCIDR("169.254.0.0/16"),     // RFC3927
	MustParseCIDR("192.0.0.0/24"),       // RFC 5736
	MustParseCIDR("192.0.2.0/24"),       // RFC 5737
	MustParseCIDR("198.51.100.0/24"),    // Assigned as TEST-NET-2
	MustParseCIDR("203.0.113.0/24"),     // Assigned as TEST-NET-3
	MustParseCIDR("192.88.99.0/24"),     // RFC 3068
	MustParseCIDR("192.18.0.0/15"),      // RFC 2544
	MustParseCIDR("224.0.0.0/4"),        // RFC 3171
	MustParseCIDR("240.0.0.0/4"),        // RFC 1112
	MustParseCIDR("255.255.255.255/32"), // RFC 919 Section 7
	MustParseCIDR("100.64.0.0/10"),      // RFC 6598
	MustParseCIDR("::/128"),             // RFC 4291: Unspecified Address
	MustParseCIDR("::1/128"),            // RFC 4291: Loopback Address
	MustParseCIDR("100::/64"),           // RFC 6666: Discard Address Block
	MustParseCIDR("2001::/23"),          // RFC 2928: IETF Protocol Assignments
	MustParseCIDR("2001:2::/48"),        // RFC 5180: Benchmarking
	MustParseCIDR("2001:db8::/32"),      // RFC 3849: Documentation
	MustParseCIDR("2001::/32"),          // RFC 4380: TEREDO
	MustParseCIDR("fc00::/7"),           // RFC 4193: Unique-Local
	MustParseCIDR("fe80::/10"),          // RFC 4291: Section 2.5.6 Link-Scoped Unicast
	MustParseCIDR("ff00::/8"),           // RFC 4291: Section 2.7
	MustParseCIDR("2002::/16"),          // RFC 7526: 6to4 anycast prefix deprecated
}
