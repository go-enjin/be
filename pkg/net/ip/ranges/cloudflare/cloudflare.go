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
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/slices"
)

var (
	knownRanges []string
	knownMutex  = &sync.RWMutex{}
)

// Init retrieves the latest IP range list from Cloudflare's service and enables
// developers to have control over when in their application lifecycle to
// perform network operations.
func Init() {
	knownMutex.Lock()
	defer knownMutex.Unlock()
	knownRanges = []string{}
	if v4ranges, err := GetIpRangesV4(); err == nil {
		for _, cidr := range v4ranges {
			if !slices.Present(cidr, knownRanges...) {
				knownRanges = append(
					knownRanges,
					cidr,
				)
			}
		}
		log.InfoF("received %d cloudflare ipv4 ranges", len(knownRanges))
	}
	if v6ranges, err := GetIpRangesV6(); err == nil {
		for _, cidr := range v6ranges {
			if !slices.Present(cidr, knownRanges...) {
				knownRanges = append(
					knownRanges,
					cidr,
				)
			}
		}
		log.InfoF("received %d cloudflare ipv6 ranges", len(knownRanges))
	}
}

// CheckRequestIP returns TRUE if the remote IP in an http.Request is from an
// Cloudflare IP range
func CheckRequestIP(req *http.Request) bool {
	if len(knownRanges) == 0 {
		Init()
	}
	return net.CheckRequestIpWithList(req, knownRanges)
}

func GetIpRanges() (ranges []string, err error) {
	if v4, err := GetIpRangesV4(); err == nil {
		ranges = append(ranges, v4...)
	}
	if v6, err := GetIpRangesV6(); err == nil {
		ranges = append(ranges, v6...)
	}
	return
}

// GetIpRangesV4 retrieves Cloudflare IPv4 ranges and returns the results
func GetIpRangesV4() (ranges []string, err error) {
	var resp *http.Response
	if resp, err = http.Get("https://www.cloudflare.com/ips-v4"); err != nil {
		log.ErrorF("error getting cloudflare IPv4 ranges: %v", err)
		return
	}
	var body []byte
	body, err = io.ReadAll(resp.Body)
	ranges = strings.Split(string(body), "\n")
	_ = resp.Body.Close()
	log.DebugF("received ipv6 ranges: [%d] total=%d", resp.StatusCode, len(ranges))
	return
}

// GetIpRangesV6 retrieves Cloudflare IPv6 ranges and returns the results
func GetIpRangesV6() (ranges []string, err error) {
	var resp *http.Response
	if resp, err = http.Get("https://www.cloudflare.com/ips-v6"); err != nil {
		log.ErrorF("error getting cloudflare IPv6 ranges: %v", err)
		return
	}
	var body []byte
	body, err = io.ReadAll(resp.Body)
	ranges = strings.Split(string(body), "\n")
	_ = resp.Body.Close()
	log.DebugF("received ipv6 ranges: [%d] total=%d", resp.StatusCode, len(ranges))
	return
}
