// Copyright (c) 2023  The Go-Enjin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package feature

import (
	"encoding/json"
)

type CSiteAuthClaimsFactor struct {
	// K is a kebab-cased feature.Tag
	K string `json:"k"`
	// N is a user-provided label for this factor
	N string `json:"n"`
	// E is an expiration unix epoch
	E int64 `json:"e"`
	// T is a timestamp unix epoch
	T int64 `json:"t"`
	// C is the challenge provided by the user
	C string `json:"c"`
}

func NewSiteAuthClaimsFactor(k, n string, e, t int64, c string) (p *CSiteAuthClaimsFactor) {
	p = &CSiteAuthClaimsFactor{
		K: k,
		N: n,
		E: e,
		T: t,
		C: c,
	}
	return
}

func ParseSiteAuthClaimsFactor(m map[string]interface{}) (p *CSiteAuthClaimsFactor, ok bool) {
	var err error
	var data []byte
	if data, err = json.Marshal(m); err != nil {
		return
	}
	p = &CSiteAuthClaimsFactor{}
	err = json.Unmarshal(data, p)
	ok = err == nil
	return
}

func (p CSiteAuthClaimsFactor) ToMap() (m map[string]interface{}) {
	m = map[string]interface{}{
		"k": p.K,
		"n": p.N,
		"e": p.E,
		"t": p.T,
		"c": p.C,
	}
	return
}