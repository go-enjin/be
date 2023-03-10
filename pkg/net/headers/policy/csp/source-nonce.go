// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import "fmt"

type NonceSource string

func NewNonceSource(nonce string) (value NonceSource) {
	value = NonceSource(nonce)
	return
}

func (s NonceSource) Type() string {
	return "nonce-source"
}

func (s NonceSource) Value() (value string) {
	value = fmt.Sprintf("nonce-%v", s)
	return
}