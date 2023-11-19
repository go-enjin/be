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

package email_hotp

import (
	"net/http"

	"github.com/go-enjin/be/pkg/context"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) getSecureContext(r *http.Request) (ctx context.Context, err error) {
	eid := userbase.GetCurrentEID(r)
	ctx, err = f.ssc.Get(eid, r, f.Site().SiteUsers())
	return
}

func (f *CFeature) setSecureContext(r *http.Request, ctx context.Context) (err error) {
	eid := userbase.GetCurrentEID(r)
	err = f.ssc.Set(eid, r, f.Site().SiteUsers(), ctx)
	return
}

func (f *CFeature) getSecureContextUnsafe(r *http.Request) (ctx context.Context, err error) {
	eid := userbase.GetCurrentEID(r)
	ctx, err = f.ssc.GetUnsafe(eid, r, f.Site().SiteUsers())
	return
}

func (f *CFeature) setSecureContextUnsafe(r *http.Request, ctx context.Context) (err error) {
	eid := userbase.GetCurrentEID(r)
	err = f.ssc.SetUnsafe(eid, r, f.Site().SiteUsers(), ctx)
	return
}

func (f *CFeature) userLock(r *http.Request) {
	eid := userbase.GetCurrentEID(r)
	f.Site().SiteUsers().LockUser(r, eid)
	return
}

func (f *CFeature) userUnlock(r *http.Request) {
	eid := userbase.GetCurrentEID(r)
	f.Site().SiteUsers().UnlockUser(r, eid)
	return
}

func (f *CFeature) userRLock(r *http.Request) {
	eid := userbase.GetCurrentEID(r)
	f.Site().SiteUsers().RLockUser(r, eid)
	return
}

func (f *CFeature) userRUnlock(r *http.Request) {
	eid := userbase.GetCurrentEID(r)
	f.Site().SiteUsers().RUnlockUser(r, eid)
	return
}

func (f *CFeature) setNewSecretKey(secret string, r *http.Request) (err error) {
	f.userLock(r)
	defer f.userUnlock(r)
	var ctx context.Context
	if ctx, err = f.getSecureContextUnsafe(r); err != nil {
		return
	}
	ctx.SetSpecific(gSecureNewSecretKey, secret)
	err = f.setSecureContextUnsafe(r, ctx)
	return
}

func (f *CFeature) getNewSecretKey(r *http.Request) (secret string) {
	var err error
	var ctx context.Context
	if ctx, err = f.getSecureContext(r); err != nil {
		return
	}
	secret = ctx.String(gSecureNewSecretKey, "")
	return
}

func (f *CFeature) removeNewSecretKey(r *http.Request) (err error) {
	f.userLock(r)
	defer f.userUnlock(r)
	var ctx context.Context
	if ctx, err = f.getSecureContextUnsafe(r); err != nil {
		return
	}
	ctx.Delete(gSecureNewSecretKey)
	err = f.setSecureContextUnsafe(r, ctx)
	return
}

func (f *CFeature) listSecureProvisions(r *http.Request) (names []string) {
	var err error
	var secure, provisions context.Context
	if secure, err = f.getSecureContext(r); err != nil {
		return
	} else if provisions = secure.Context("provisions"); provisions != nil {
		names = provisions.Keys()
	}
	return
}

func (f *CFeature) countSecureProvisions(r *http.Request) (count int) {
	var err error
	var secure, provisions context.Context
	if secure, err = f.getSecureContext(r); err != nil {
		return
	} else if provisions = secure.Context("provisions"); provisions != nil {
		count = provisions.Len()
	}
	return
}

type provisionedFactor struct {
	S string `json:"s"`
	C int64  `json:"c"`
	M string `json:"m"`
}

func parseProvision(v interface{}) (p *provisionedFactor) {
	switch t := v.(type) {
	case map[string]interface{}:
		ctx := context.Context(t)
		if s := ctx.String("s", ""); s != "" {
			if c := ctx.Int64("c", -1); c > -1 {
				if m := ctx.String("m", ""); m != "" {
					p = &provisionedFactor{S: s, C: c, M: m}
				}
			}
		}
	case *provisionedFactor:
		p = t
	}
	return
}

func (f *CFeature) hasSecureProvision(key string, r *http.Request) (present bool) {
	var err error
	var secure, provisions context.Context
	if secure, err = f.getSecureContext(r); err != nil {
		return
	} else if provisions = secure.Context("provisions"); provisions == nil {
		return
	}
	present = parseProvision(provisions.Get(key)) != nil
	return
}

func (f *CFeature) getSecureProvision(key string, r *http.Request) (email, secret string, count int64, err error) {
	var secure, provisions context.Context
	if secure, err = f.getSecureContext(r); err != nil {
		return
	} else if provisions = secure.Context("provisions"); provisions == nil {
	} else if provision := parseProvision(provisions.Get(key)); provision != nil {
		secret = provision.S
		count = provision.C
		email = provision.M
		return
	}
	err = berrs.ErrSecretNotFound
	return
}

func (f *CFeature) setSecureProvision(key, email, secret string, count int64, r *http.Request) (err error) {
	f.userLock(r)
	defer f.userUnlock(r)
	var secure, provisions context.Context
	if secure, err = f.getSecureContextUnsafe(r); err != nil {
		return
	} else if provisions = secure.Context("provisions"); provisions == nil {
		provisions = context.Context{}
	}
	provisions.SetSpecific(key, &provisionedFactor{S: secret, C: count, M: email})
	secure.SetSpecific("provisions", provisions)
	err = f.setSecureContextUnsafe(r, secure)
	return
}

func (f *CFeature) revokeSecureProvision(key string, r *http.Request) (err error) {
	f.userLock(r)
	defer f.userUnlock(r)
	var secure, provisions context.Context
	if secure, err = f.getSecureContextUnsafe(r); err != nil {
		return
	} else if provisions = secure.Context("provisions"); provisions == nil {
		provisions = context.Context{}
	}
	provisions.Delete(key)
	secure.SetSpecific("provisions", provisions)
	err = f.setSecureContextUnsafe(r, secure)
	return
}