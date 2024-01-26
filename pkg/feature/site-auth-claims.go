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
	"github.com/golang-jwt/jwt/v4"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/hash/sha"
)

type CSiteAuthClaims struct {
	// RID is the Real ID of the user
	RID string
	// EID is the Enjin ID of the user
	EID string
	// Email is the email address of the user
	Email string
	// Context is the variable metadata related to the user
	Context beContext.Context

	jwt.RegisteredClaims
}

// ResetFactorValueTypes will check all factors and verifications for factors-as-maps and convert them to their
// proper concrete types
func (c *CSiteAuthClaims) ResetFactorValueTypes() {
	for _, topKey := range []string{"f", "v"} {
		if ctx := c.Context.Context(topKey); ctx.Len() > 0 {
			for k, v := range ctx {
				if m, ok := v.(map[string]interface{}); ok {
					if claim, ok := ParseSiteAuthClaimsFactor(m); ok {
						ctx[k] = claim
					}
				} else if l, ok := ParseSiteAuthClaimsFactors(v); ok {
					ctx[k] = l
				}
			}
		}
	}

}

// GetAudience returns the first name in the .Audience ClaimStrings list
func (c *CSiteAuthClaims) GetAudience() (name string) {
	if len(c.Audience) > 0 {
		name = c.Audience[0]
	}
	return
}

func (c *CSiteAuthClaims) SetFactor(factor *CSiteAuthClaimsFactor) {
	var ok bool
	var factors []*CSiteAuthClaimsFactor
	if v := c.Context.Get(".f." + factor.K); v != nil {
		if factors, ok = v.(SiteAuthClaimsFactors); !ok {
			factors, _ = ParseSiteAuthClaimsFactors(v)
		}
	}
	factors = append(factors, factor)
	_ = c.Context.SetKV(".f."+factor.K, factors)
	return
}

func (c *CSiteAuthClaims) GetFactor(kebab, name string) (factor *CSiteAuthClaimsFactor, ok bool) {
	var factors SiteAuthClaimsFactors
	if v := c.Context.Get(".f." + kebab); v != nil {
		if factors, ok = v.(SiteAuthClaimsFactors); !ok {
			factors, _ = ParseSiteAuthClaimsFactors(v)
		}
	}
	factor, ok = factors.Find(kebab + ";" + name)
	return
}

func (c *CSiteAuthClaims) GetFactors(kebab string) (factors SiteAuthClaimsFactors) {
	var ok bool
	if v := c.Context.Get(".f." + kebab); v != nil {
		if factors, ok = v.(SiteAuthClaimsFactors); !ok {
			factors, _ = ParseSiteAuthClaimsFactors(v)
		}
	}
	return
}

func (c *CSiteAuthClaims) GetAllFactors() (factors SiteAuthClaimsFactors) {
	if f := c.Context.Context(".f"); f != nil {
		for _, key := range f.Keys() {
			if v := f.Get(key); v != nil {
				if more, ok := v.(SiteAuthClaimsFactors); ok {
					factors = append(factors, more...)
				} else if more, ok = ParseSiteAuthClaimsFactors(v); ok {
					factors = append(factors, more...)
				}
			}
		}
	}
	return
}

func (c *CSiteAuthClaims) RevokeFactor(kebab, name string) {
	var ok bool
	var factors SiteAuthClaimsFactors
	if v := c.Context.Get(".f." + kebab); v != nil {
		if factors, ok = v.(SiteAuthClaimsFactors); !ok {
			factors, _ = ParseSiteAuthClaimsFactors(v)
		}
	}
	factors = factors.Filter(kebab + ";" + name)
	_ = c.Context.SetKV(".f."+kebab, factors)
	return
}

func (c *CSiteAuthClaims) RevokeFactors(kebab string) {
	_ = c.Context.Delete(".f." + kebab)
	return
}

func (c *CSiteAuthClaims) SetVerifiedFactor(target string, factor *CSiteAuthClaimsFactor) {
	shasum := sha.MustDataHash10(target)
	_ = c.Context.SetKV(".v."+shasum, factor)
	return
}

func (c *CSiteAuthClaims) GetVerifiedFactor(target string) (factor *CSiteAuthClaimsFactor, ok bool) {
	shasum := sha.MustDataHash10(target)
	if v := c.Context.Get(".v." + shasum); v != nil {
		if factor, ok = v.(*CSiteAuthClaimsFactor); ok {
			return
		}
		var m map[string]interface{}
		if m, ok = v.(map[string]interface{}); ok {
			factor, ok = ParseSiteAuthClaimsFactor(m)
		}
	}
	return
}

func (c *CSiteAuthClaims) RevokeVerifiedFactor(target string) {
	shasum := sha.MustDataHash10(target)
	c.Context.Delete(".v." + shasum)
	return
}
