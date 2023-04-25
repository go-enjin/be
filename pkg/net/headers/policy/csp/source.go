// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import (
	"sort"
	"strings"

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

func (s Sources) FilterAllowedTypes(allowed ...string) (filtered Sources) {
	isAllowed := func(st string) bool {
		for _, at := range allowed {
			if at == st {
				return true
			}
		}
		return false
	}
	for _, src := range s {
		if isAllowed(src.SourceType()) {
			filtered = append(filtered, src)
		}
	}
	return
}

func (s Sources) FilterAllowedKeywords(allowed ...KeywordSource) (filtered Sources) {
	isAllowed := func(st string) bool {
		st = strings.ReplaceAll(st, "'", "")
		for _, at := range allowed {
			if at == KeywordSource(st) {
				return true
			}
		}
		return false
	}
	for _, src := range s {
		if src.SourceType() == KeywordSourceType {
			if isAllowed(src.Value()) {
				filtered = append(filtered, src)
			}
		} else {
			filtered = append(filtered, src)
		}
	}
	return
}

func (s Sources) FilterUnsafeInline() (filtered Sources) {
	var hasNonceHash bool
	for _, src := range s {
		srcType := src.SourceType()
		if hasNonceHash = srcType == NonceSourceType || srcType == HashSourceType; hasNonceHash {
			break
		}
	}
	for _, src := range s {
		if hasNonceHash && src.Value() == UnsafeInline.Value() {
			continue
		}
		filtered = append(filtered, src)
	}
	return
}

func (s Sources) Collapse() (collapsed Sources) {
	check := make(map[string]Source)
	for _, src := range s {
		check[src.Value()] = src
	}
	for _, src := range check {
		collapsed = append(collapsed, src)
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