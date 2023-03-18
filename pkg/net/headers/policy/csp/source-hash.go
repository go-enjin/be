// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import (
	"fmt"
	"strings"

	"github.com/go-enjin/be/pkg/log"
)

var _ Source = (*HashSource)(nil)

const HashSourceType string = "hash-source"

type HashSource struct {
	algo string
	hash string
}

func ParseHashSource(input string) (s HashSource, ok bool) {
	if l := len(input); l > 2 {
		if input[0] == '\'' && input[l-1] == '\'' {
			input = input[1 : l-1]
		}
	}
	if ok = strings.HasPrefix(input, "sha256-"); ok {
		s = HashSource{algo: "sha256", hash: input[7:]}
	} else if ok = strings.HasPrefix(input, "sha384-"); ok {
		s = HashSource{algo: "sha384", hash: input[7:]}
	} else if ok = strings.HasPrefix(input, "sha512-"); ok {
		s = HashSource{algo: "sha512", hash: input[7:]}
	}
	return
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

func (s HashSource) SourceType() string {
	return HashSourceType
}

func (s HashSource) Value() (value string) {
	value = fmt.Sprintf("'%v-%v'", s.algo, s.hash)
	return
}