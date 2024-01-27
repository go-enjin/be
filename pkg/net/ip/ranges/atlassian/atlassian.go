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

package atlassian

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/go-corelibs/slices"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
)

type IpRangeItem struct {
	Network   string   `json:"network"`
	MaskLen   int      `json:"mask_len"`
	Cidr      string   `json:"cidr"`
	Mask      string   `json:"mask"`
	Region    []string `json:"region"`
	Product   []string `json:"product"`
	Direction []string `json:"direction"`
}

type IpRangeResponse struct {
	CreationDate string        `json:"creationDate"`
	SyncToken    int           `json:"syncToken"`
	Items        []IpRangeItem `json:"items"`
}

var (
	knownRanges []string
	knownMutex  = &sync.RWMutex{}
)

// Init retrieves the latest IP range list from Atlassian's service and enables
// developers to have control over when in their application lifecycle to
// perform network operations.
func Init() {
	knownMutex.Lock()
	defer knownMutex.Unlock()
	knownRanges = []string{}
	if ranges, err := GetIpRanges(); err == nil {
		for _, cidr := range ranges {
			if !slices.Present(cidr, knownRanges...) {
				knownRanges = append(
					knownRanges,
					cidr,
				)
			}
		}
		log.InfoF("received %d atlassian ip ranges", len(knownRanges))
	}
}

// CheckRequestIP returns TRUE if the remote IP in an http.Request is from an
// Atlassian IP range
func CheckRequestIP(req *http.Request) bool {
	if len(knownRanges) == 0 {
		Init()
	}
	return net.CheckRequestIpWithList(req, knownRanges)
}

// GetIpRanges retrieves Atlassian IP ranges and returns the results
func GetIpRanges() (ranges []string, err error) {
	data := &IpRangeResponse{
		CreationDate: "",
		SyncToken:    0,
		Items:        []IpRangeItem{},
	}
	var resp *http.Response
	if resp, err = http.Get("https://ip-ranges.atlassian.com/"); err != nil {
		log.ErrorF("error getting atlassian IP ranges: %v", err)
		return
	}
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		_ = resp.Body.Close()
		log.ErrorF("error parsing atlassian IP ranges: %v", err)
		return
	}
	_ = resp.Body.Close()
	for _, item := range data.Items {
		ranges = append(ranges, item.Cidr)
	}
	log.DebugF("received ip ranges: [%d] total=%d", resp.StatusCode, len(ranges))
	return
}
