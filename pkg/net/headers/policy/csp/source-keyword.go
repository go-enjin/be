// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import "fmt"

var _ Source = (*KeywordSource)(nil)

const KeywordSourceType string = "keyword-source"

type KeywordSource string

const (
	None           KeywordSource = `none`
	Self           KeywordSource = `self`
	UnsafeEval     KeywordSource = `unsafe-eval`
	UnsafeHashes   KeywordSource = `unsafe-hashes`
	UnsafeInline   KeywordSource = `unsafe-inline`
	RequireSample  KeywordSource = `require-sample`
	StrictDynamic  KeywordSource = `strict-dynamic`
	WasmUnsafeEval KeywordSource = `wasm-unsafe-eval`
)

func ParseKeywordSource(input string) (s KeywordSource, ok bool) {
	if l := len(input); l > 2 {
		if input[0] == '\'' && input[l-1] == '\'' {
			input = input[1 : l-1]
		}
	}
	v := KeywordSource(input)
	if ok = v == None || v == Self || v == UnsafeEval || v == UnsafeHashes ||
		v == UnsafeInline || v == RequireSample || v == StrictDynamic ||
		v == WasmUnsafeEval; ok {
		s = v
	}
	return
}

func (s KeywordSource) SourceType() string {
	return KeywordSourceType
}

func (s KeywordSource) Value() (value string) {
	value = fmt.Sprintf(`'%v'`, s)
	return
}
