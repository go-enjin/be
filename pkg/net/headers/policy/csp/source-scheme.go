// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-enjin/be/pkg/log"
)

var _ Source = (*SchemeSource)(nil)

const SchemeSourceType string = "scheme-source"

type SchemeSource string

var (
	rxSchemeSource = regexp.MustCompile(`^\s*([a-z0-9][-+.a-z0-9]*):\s*$`)
)

func ParseSchemeSource(input string) (s SchemeSource, ok bool) {
	if ok = rxSchemeSource.MatchString(input); ok {
		s = NewSchemeSource(input)
	}
	return
}

func NewSchemeSource(value string) (v SchemeSource) {
	if !strings.HasSuffix(value, ":") {
		value += ":"
	}
	if rxSchemeSource.MatchString(value) {
		m := rxSchemeSource.FindAllStringSubmatch(value, 1)
		v = SchemeSource(m[0][1])
		return
	}
	log.PanicF("invalid scheme-source value: %v", value)
	return
}

func (s SchemeSource) SourceType() string {
	return SchemeSourceType
}

func (s SchemeSource) Value() (value string) {
	value = fmt.Sprintf("%s:", s)
	return
}