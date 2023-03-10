// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import "fmt"

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

func (s KeywordSource) Type() string {
	return "keyword-source"
}

func (s KeywordSource) Value() (value string) {
	value = fmt.Sprintf(`'%v'`, s)
	return
}