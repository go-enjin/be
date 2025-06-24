// Copyright (c) 2025  The Go-Enjin Authors
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

package log

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/go-corelibs/context"
	"github.com/go-corelibs/values"
)

type SyslogDSN struct {
	network string
	address string
	port    string
	Options context.Context
}

func ParseSyslogDSN(input string) (dsn *SyslogDSN, err error) {
	if len(input) == 0 || input == "local" {
		dsn = &SyslogDSN{network: "", port: ""}
		return
	}

	var parsed *url.URL
	if parsed, err = url.Parse(input); err != nil {
		return
	}

	var network string
	switch parsed.Scheme {
	case "local":
	case "udp", "tcp", "tcp+tls":
		network = parsed.Scheme
	default:
		err = fmt.Errorf("invalid syslog dsn network: %q", parsed.Scheme)
		return
	}

	var host, port string
	if parsed.Scheme != "local" {
		if strings.Contains(parsed.Host, ":") {
			if host, port, err = net.SplitHostPort(parsed.Host); err != nil {
				return
			}
		} else {
			host = parsed.Host
			if parsed.Scheme == "tcp+tls" {
				port = "6514"
			} else {
				port = "514"
			}
		}
	}

	var params url.Values
	var options context.Context
	if params, err = url.ParseQuery(parsed.RawQuery); err != nil {
		return
	}
	if len(params) > 0 {
		options = context.Context{}
		for key, val := range params {
			switch strings.ToLower(key) {
			case "skipverifytls":
				options["SkipVerifyTLS"] = values.IsBoolTrue(val[0])
			case "format":
				lv := strings.ToLower(val[0])
				if f, ok := formatLookup[lv]; ok {
					options["Format"] = f
				} else {
					err = fmt.Errorf("syslog dsn format option is not one of: \"pretty\", \"json\" or \"text\"")
					return
				}
			default:
			}
		}
	}

	dsn = &SyslogDSN{
		network: network,
		address: host,
		port:    port,
		Options: options,
	}
	return
}

// String returns the parsed DSN as a string
func (s *SyslogDSN) String() string {
	if s.network == "local" {
		return "local://"
	}
	var options string
	if len(s.Options) > 0 {
		for idx, key := range s.Options.Keys() {
			if idx == 0 {
				options += "?"
			} else {
				options += "&"
			}
			options += key + "=" + fmt.Sprintf("%v", s.Options.Get(key))
		}
	}
	return s.network + "://" + s.address + ":" + s.port + options
}

func (s *SyslogDSN) Network() (network string) {
	if s.network == "local" {
		return
	}
	return s.network
}

func (s *SyslogDSN) Raddr() (raddr string) {
	if s.network == "local" {
		return
	}
	return s.address + ":" + s.port
}
