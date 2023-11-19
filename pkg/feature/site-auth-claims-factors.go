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
	"strings"
)

type SiteAuthClaimsFactors []*CSiteAuthClaimsFactor

func ParseSiteAuthClaimsFactors(v interface{}) (p SiteAuthClaimsFactors, ok bool) {
	var err error
	var data []byte
	if data, err = json.Marshal(v); err != nil {
		return
	}
	p = SiteAuthClaimsFactors{}
	err = json.Unmarshal(data, &p)
	ok = err == nil
	return
}

func (f SiteAuthClaimsFactors) Len() (count int) {
	count = len(f)
	return
}

// Find returns the first factor that matches the "kebab;name" provision string given
func (f SiteAuthClaimsFactors) Find(provision string) (found *CSiteAuthClaimsFactor, ok bool) {
	if f.Len() == 0 {
		return
	}
	var kebab, name string
	if kebab, name, ok = strings.Cut(provision, ";"); ok {
		for _, p := range f {
			if ok = p.K == kebab && p.N == name; ok {
				found = p
				return
			}
		}
	}
	return
}

// Filter returns the same list of factors, without the provision given
func (f SiteAuthClaimsFactors) Filter(provision string) (modified SiteAuthClaimsFactors) {
	if f.Len() == 0 {
		return
	}
	var ok bool
	var kebab, name string
	if kebab, name, ok = strings.Cut(provision, ";"); ok {
		for _, p := range f {
			if ok = p.K == kebab && p.N == name; !ok {
				modified = append(modified, p)
			}
		}
	}
	return
}