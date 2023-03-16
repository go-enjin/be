// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import (
	"regexp"
	"strings"

	"github.com/go-enjin/be/pkg/log"
)

var _ Source = (*HostSource)(nil)

const HostSourceType string = "host-source"

type HostSource struct {
	scheme string
	host   string
	port   string
	path   string
}

var (
	rxHostSource       = regexp.MustCompile(`^\s*([a-z0-9]+[-.a-z0-9]*:)?(?://)?([^:/\s]+)?(:\d+)?(/.+?)?\s*$`)
	rxHostSourceDomain = regexp.MustCompile(`^\s*([a-zA-Z0-9][-a-zA-Z0-9]*\\.)+[A-Za-z]{2, 6}\s*$`)
)

func ParseHostSource(input string) (s HostSource, ok bool) {
	if ok = rxHostSourceDomain.MatchString(input); ok {
		s = HostSource{
			host: strings.TrimSpace(input),
		}
	} else if ok = rxHostSource.MatchString(input); ok {
		s = NewHostSource(input)
	}
	return
}

func NewHostSource(value string) (v HostSource) {
	if rxHostSource.MatchString(value) {
		m := rxHostSource.FindAllStringSubmatch(value, 1)
		v = HostSource{
			scheme: m[0][1],
			host:   m[0][2],
			port:   m[0][3],
			path:   m[0][4],
		}
		if end := len(v.scheme); end > 0 {
			// trim trailing colon
			v.scheme = v.scheme[:end-1]
		}
		return
	}
	log.PanicF("invalid host-source value: %v", value)
	return
}

func (s HostSource) SourceType() string {
	return HostSourceType
}

func (s HostSource) Value() (value string) {
	if s.scheme != "" {
		value = s.scheme + ":"
	}
	if s.host != "" {
		if s.scheme != "" {
			value += "//" + s.host
		} else {
			value += s.host
		}
		if s.port != "" {
			value += s.port
		}
		if s.path != "" {
			value += s.path
		}
	}
	return
}