// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import (
	"sort"

	"github.com/fvbommel/sortorder"
)

type Sources []Source

func (s Sources) Append(sources ...Source) (modified Sources) {
	modified = append(modified, s...)
	for _, source := range sources {
		var found bool
		for _, check := range modified {
			if found = source.SourceType() == check.SourceType() && source.Value() == check.Value(); found {
				break
			}
		}
		if !found {
			modified = append(modified, source)
		}
	}
	return
}

func (s Sources) Sort() (sorted Sources) {
	sorted = append(sorted, s...)
	sort.Slice(sorted, func(i, j int) (less bool) {
		a, b := sorted[i], sorted[j]
		if aType, bType := a.SourceType(), b.SourceType(); aType != bType {
			switch aType {
			case KeywordSourceType:
				return true
			case NonceSourceType:
				return bType != KeywordSourceType
			case HashSourceType:
				switch bType {
				case KeywordSourceType, NonceSourceType:
					return false
				}
			case SchemeSourceType:
				switch bType {
				case KeywordSourceType, NonceSourceType, HashSourceType:
					return false
				}
			case HostSourceType:
				return false
			}
			return true
		}
		return sortorder.NaturalLess(a.Value(), b.Value())
	})
	return
}

type Source interface {
	SourceType() string
	Value() string
}

func ParseSource(input string) (s Source, ok bool) {
	if s, ok = ParseKeywordSource(input); ok {
	} else if s, ok = ParseNonceSource(input); ok {
	} else if s, ok = ParseHashSource(input); ok {
	} else if s, ok = ParseSchemeSource(input); ok {
	} else if s, ok = ParseHostSource(input); ok {
	}
	return
}