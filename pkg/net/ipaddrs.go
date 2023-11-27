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

package net

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var rxSplitForwardedFor = regexp.MustCompile(`\s*,\s*`)

func GetProxyIpFromRequest(r *http.Request) (ip string, err error) {
	xForwardedFor := r.Header.Get("X-Forwarded-FOR")
	switch {
	case xForwardedFor != "":
		if parts := rxSplitForwardedFor.Split(xForwardedFor, -1); len(parts) > 0 {
			// the first non-remote-addr IP address
			if len(parts) == 1 {
				ip = parts[0]
			} else {
				ip = parts[1]
			}
		}
		return
	}
	err = fmt.Errorf("request not proxied")
	return
}

func GetIpFromRequest(r *http.Request) (ip string, err error) {
	// See: https://golangbyexample.com/golang-ip-address-http-request/

	// Get IP from CF-Connecting-IP header
	ip = r.Header.Get("Cf-Connecting-Ip")
	if netIP := net.ParseIP(ip); netIP != nil {
		ip = netIP.String()
		// log.DebugF("cf-connecting-ip: %v", ip)
		return
	}

	// Get IP from X-FORWARDED-FOR header
	xff := r.Header.Get("X-Forwarded-For")
	ips := rxSplitForwardedFor.Split(xff, -1)
	if len(ips) > 0 {
		if netIP := net.ParseIP(ips[0]); netIP != nil {
			ip = netIP.String()
			// log.DebugF("x-forward-for: %v", ip)
			return
		}
	}

	// Get IP from the X-REAL-IP header
	ip = r.Header.Get("X-Real-IP")
	if netIP := net.ParseIP(ip); netIP != nil {
		ip = netIP.String()
		// log.DebugF("x-real-ip: %v", ip)
		return
	}

	// Get IP from RemoteAddr
	if ip = GetRemoteAddr(r); ip != "" {
		// log.DebugF("remoteAddr: %v", ip)
		return
	}

	err = fmt.Errorf("remote address not found")
	return
}

func GetRemoteAddr(r *http.Request) (ip string) {
	if sip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if netIP := net.ParseIP(sip); netIP != nil {
			ip = netIP.String()
			return
		}
	}
	if netIP := net.ParseIP(r.RemoteAddr); netIP != nil {
		ip = netIP.String()
	}
	return
}

func ParseIpFromRequest(r *http.Request) (ip net.IP, err error) {
	var address string
	if address, err = GetIpFromRequest(r); err == nil {
		if host, _, ee := net.SplitHostPort(address); ee == nil {
			address = host
		}
		return net.ParseIP(address), nil
	}
	return
}

func IsIpInRange(ip string, cidr string) bool {
	if netIp := net.ParseIP(ip); netIp != nil {
		if _, subnet, err := net.ParseCIDR(cidr); err == nil {
			return subnet.Contains(netIp)
		}
	}
	return false
}

func IsNetIpInRange(ip net.IP, cidr string) bool {
	if _, subnet, err := net.ParseCIDR(cidr); err == nil {
		return subnet.Contains(ip)
	}
	return false
}

// CheckRequestIpWithList returns TRUE if the remote IP in an http.Request is
// from an IP in the given list of CIDR ranges
func CheckRequestIpWithList(req *http.Request, list []string) bool {
	if ip, err := ParseIpFromRequest(req); err == nil {
		for _, cidr := range list {
			if IsNetIpInRange(ip, cidr) {
				return true
			}
		}
	}
	return false
}

// IsNetIpPrivate returns true if the IP address given is within any of the
// local or private network CIDR ranges (not public IP address space)
func IsNetIpPrivate(ip net.IP) (private bool) {
	for _, cidr := range PrivateNetworks {
		if private = cidr.Contains(ip); private {
			break
		}
	}
	return
}

func ParseIP(input ...string) (ips []net.IP, err error) {
	for _, lines := range input {
		for _, line := range strings.Split(lines, "\n") {
			for _, part := range strings.Split(line, " ") {
				if cleaned := strings.TrimSpace(part); cleaned != "" {
					if parsed := net.ParseIP(cleaned); parsed == nil {
						err = fmt.Errorf("%q is not an address", cleaned)
						return
					} else {
						ips = append(ips, parsed)
					}
				}
			}
		}
	}
	return
}

func ParseCIDR(input ...string) (cidrs []*net.IPNet, err error) {
	for _, lines := range input {
		for _, line := range strings.Split(lines, "\n") {
			for _, part := range strings.Split(line, " ") {
				if cleaned := strings.TrimSpace(part); cleaned != "" {
					if _, network, ee := net.ParseCIDR(cleaned); ee != nil {
						err = fmt.Errorf("error parsing CIDR %q: %w", cleaned, ee)
						return
					} else {
						cidrs = append(cidrs, network)
					}
				}
			}
		}
	}
	return
}
