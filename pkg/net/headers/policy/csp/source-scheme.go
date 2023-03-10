// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import (
	"fmt"
	"regexp"

	"github.com/go-enjin/be/pkg/log"
)

type SchemeSource string

var (
	rxSchemeSource = regexp.MustCompile(`^\s*([a-z0-9]+[-.a-z0-9]*):?\s*$`)
)

func NewSchemeSource(value string) (v SchemeSource) {
	if rxSchemeSource.MatchString(value) {
		m := rxSchemeSource.FindAllStringSubmatch(value, 1)
		v = SchemeSource(m[0][1])
		return
	}
	log.PanicF("invalid scheme-source value: %v", value)
	return
}

func (s SchemeSource) Type() string {
	return "scheme-source"
}

func (s SchemeSource) Value() (value string) {
	value = fmt.Sprintf("%s:", s)
	return
}