// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import (
	"fmt"
	"strings"
)

var _ Source = (*NonceSource)(nil)

const NonceSourceType string = "nonce-source"

type NonceSource string

func ParseNonceSource(input string) (s NonceSource, ok bool) {
	if l := len(input); l > 2 {
		if input[0] == '\'' && input[l-1] == '\'' {
			input = input[1 : l-1]
		}
	}
	if ok = strings.HasPrefix(input, "nonce-"); ok {
		s = NewNonceSource(input[7:])
	}
	return
}

func NewNonceSource(nonce string) (value NonceSource) {
	value = NonceSource(nonce)
	return
}

func (s NonceSource) SourceType() string {
	return NonceSourceType
}

func (s NonceSource) Value() (value string) {
	value = fmt.Sprintf("'nonce-%v'", s)
	return
}
