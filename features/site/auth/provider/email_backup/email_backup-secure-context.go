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

package email_backup

import (
	"net/http"

	"github.com/go-enjin/be/pkg/context"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) userLock(eid string, r *http.Request) {
	eid = f.parseEID(eid, r)
	f.Site().SiteUsers().LockUser(r, eid)
	return
}

func (f *CFeature) userUnlock(eid string, r *http.Request) {
	eid = f.parseEID(eid, r)
	f.Site().SiteUsers().UnlockUser(r, eid)
	return
}

func (f *CFeature) userRLock(eid string, r *http.Request) {
	eid = f.parseEID(eid, r)
	f.Site().SiteUsers().RLockUser(r, eid)
	return
}

func (f *CFeature) userRUnlock(eid string, r *http.Request) {
	eid = f.parseEID(eid, r)
	f.Site().SiteUsers().RUnlockUser(r, eid)
	return
}

func (f *CFeature) parseEID(eid string, r *http.Request) (parsed string) {
	if parsed = eid; parsed == "" {
		parsed = userbase.GetCurrentEID(r)
	}
	return
}

func (f *CFeature) getSecureContext(eid string, r *http.Request) (ctx context.Context, err error) {
	eid = f.parseEID(eid, r)
	ctx, err = f.ssc.Get(eid, r, f.Site().SiteUsers())
	return
}

func (f *CFeature) setSecureContext(eid string, r *http.Request, ctx context.Context) (err error) {
	eid = f.parseEID(eid, r)
	err = f.ssc.Set(eid, r, f.Site().SiteUsers(), ctx)
	return
}

func (f *CFeature) getSecureContextUnsafe(eid string, r *http.Request) (ctx context.Context, err error) {
	eid = f.parseEID(eid, r)
	ctx, err = f.ssc.GetUnsafe(eid, r, f.Site().SiteUsers())
	return
}

func (f *CFeature) setSecureContextUnsafe(eid string, r *http.Request, ctx context.Context) (err error) {
	eid = f.parseEID(eid, r)
	err = f.ssc.SetUnsafe(eid, r, f.Site().SiteUsers(), ctx)
	return
}

func (f *CFeature) listSecureProvisions(eid string, r *http.Request) (names []string) {
	eid = f.parseEID(eid, r)
	var err error
	var secure, provisions context.Context
	if secure, err = f.getSecureContext(eid, r); err != nil {
		return
	} else if provisions = secure.Context("provisions"); provisions != nil {
		names = provisions.Keys()
	}
	return
}

func (f *CFeature) countSecureProvisions(eid string, r *http.Request) (count int) {
	eid = f.parseEID(eid, r)
	var err error
	var secure, provisions context.Context
	if secure, err = f.getSecureContext(eid, r); err != nil {
		return
	} else if provisions = secure.Context("provisions"); provisions != nil {
		count = provisions.Len()
	}
	return
}

type provisionedFactor struct {
	M string `json:"m"`
}

func parseProvision(v interface{}) (p *provisionedFactor) {
	switch t := v.(type) {
	case map[string]interface{}:
		ctx := context.Context(t)
		if email := ctx.String("m", ""); email != "" {
			p = &provisionedFactor{M: email}
		}
	case *provisionedFactor:
		p = t
	}
	return
}

func (f *CFeature) getProvisionByEmail(eid, email string, r *http.Request) (name string, present bool) {
	eid = f.parseEID(eid, r)
	var err error
	var secure, provisions context.Context
	if secure, err = f.getSecureContext(eid, r); err != nil {
		return
	} else if provisions = secure.Context("provisions"); provisions != nil {
		for key, v := range provisions {
			if p := parseProvision(v); p != nil {
				if present = email == p.M; present {
					name = key
					return
				}
			}
		}
	}
	return
}

func (f *CFeature) hasSecureProvision(eid, key string, r *http.Request) (present bool) {
	eid = f.parseEID(eid, r)
	var err error
	var secure, provisions context.Context
	if secure, err = f.getSecureContext(eid, r); err != nil {
		return
	} else if provisions = secure.Context("provisions"); provisions == nil {
		return
	}
	present = parseProvision(provisions.Get(key)) != nil
	return
}

func (f *CFeature) getSecureProvision(eid, key string, r *http.Request) (email string, err error) {
	eid = f.parseEID(eid, r)
	var secure, provisions context.Context
	if secure, err = f.getSecureContext(eid, r); err != nil {
		return
	} else if provisions = secure.Context("provisions"); provisions == nil {
	} else if provision := parseProvision(provisions.Get(key)); provision != nil {
		email = provision.M
		return
	}
	err = berrs.ErrSecretNotFound
	return
}

func (f *CFeature) setSecureProvision(eid, key, email string, r *http.Request) (err error) {
	eid = f.parseEID(eid, r)
	f.userLock(eid, r)
	defer f.userUnlock(eid, r)
	var secure, provisions context.Context
	if secure, err = f.getSecureContextUnsafe(eid, r); err != nil {
		return
	} else if provisions = secure.Context("provisions"); provisions == nil {
		provisions = context.Context{}
	}
	provisions.SetSpecific(key, &provisionedFactor{M: email})
	secure.SetSpecific("provisions", provisions)
	err = f.setSecureContextUnsafe(eid, r, secure)
	return
}

func (f *CFeature) revokeSecureProvision(eid, key string, r *http.Request) (err error) {
	eid = f.parseEID(eid, r)
	f.userLock(eid, r)
	defer f.userUnlock(eid, r)
	var secure, provisions context.Context
	if secure, err = f.getSecureContextUnsafe(eid, r); err != nil {
		return
	} else if provisions = secure.Context("provisions"); provisions == nil {
		provisions = context.Context{}
	}
	provisions.Delete(key)
	secure.SetSpecific("provisions", provisions)
	err = f.setSecureContextUnsafe(eid, r, secure)
	return
}