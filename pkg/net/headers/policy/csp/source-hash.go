// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import (
	"fmt"

	"github.com/go-enjin/be/pkg/log"
)

type HashSource struct {
	algo string
	hash string
}

func NewHashSource(algo, hash string) (value HashSource) {
	switch algo {
	case "sha256", "sha384", "sha512":
	default:
		log.PanicF("invalid hash algorithm: %v (%v)", algo, hash)
		return
	}
	value = HashSource{
		algo: algo,
		hash: hash,
	}
	return
}

func (s HashSource) Type() string {
	return "hash-source"
}

func (s HashSource) Value() (value string) {
	value = fmt.Sprintf("%v-%v", s.algo, s.hash)
	return
}