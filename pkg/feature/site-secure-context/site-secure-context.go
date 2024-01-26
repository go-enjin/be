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

package site_secure_context

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/crypto"
	"github.com/go-enjin/be/pkg/feature"
)

/*

	Note: do not embed this feature component! Only use as a private member!

*/

type CSecureContext struct {
	this interface{}
	_tag feature.Tag
	_key []byte
}

func New(this interface{}) (c *CSecureContext) {
	c = new(CSecureContext)
	c.Construct(this)
	return
}

func (c *CSecureContext) Construct(this interface{}) {
	if f, ok := this.(feature.Feature); ok {
		c.this = this
		c._tag = f.Tag()
	} else {
		panic(fmt.Sprintf("%T is not a feature.Feature", this))
	}
}

func (c *CSecureContext) Build(b feature.Buildable) (err error) {
	kebab := c._tag.Kebab()
	flagName := kebab + "-secret-key"
	b.AddFlags(&cli.StringFlag{
		Name:     flagName,
		Usage:    "specify the secret key for this feature",
		EnvVars:  b.MakeEnvKeys(flagName),
		Category: kebab,
	})
	return
}

func (c *CSecureContext) Startup(ctx *cli.Context) (err error) {
	flagName := c._tag.Kebab() + "-secret-key"

	if ctx.IsSet(flagName) {
		if v := ctx.String(flagName); v != "" {
			c._key = []byte(v)
		} else {
			err = fmt.Errorf("--%s is required", flagName)
			return
		}
	} else {
		err = fmt.Errorf("--%s is required", flagName)
		return
	}

	if len(c._key) < 32 {
		err = fmt.Errorf("secret key is less than 32 bytes")
		return
	}

	return
}

func (c *CSecureContext) Get(eid string, r *http.Request, sup feature.SiteUsersProvider) (ctx context.Context, err error) {
	sup.RLockUser(r, eid)
	defer sup.RUnlockUser(r, eid)
	ctx, err = c.GetUnsafe(eid, r, sup)
	return
}

func (c *CSecureContext) GetUnsafe(eid string, r *http.Request, sup feature.SiteUsersProvider) (ctx context.Context, err error) {

	var au feature.User
	if au, err = sup.RetrieveUser(r, eid); err != nil {
		return
	}

	ctx = make(context.Context)

	if _, v := au.UnsafeContext().GetKV(".secure." + c._tag.Kebab()); v != nil {
		var encoded string
		switch t := v.(type) {
		case []byte:
			encoded = string(t)
		case string:
			encoded = t
		}
		if len(encoded) > 0 {
			var data []byte
			if data, err = crypto.Decrypt(c._key, encoded); err != nil {
				return
			} else if err = json.Unmarshal(data, &ctx); err != nil {
				return
			}
		}
	}

	return
}

func (c *CSecureContext) Set(eid string, r *http.Request, sup feature.SiteUsersProvider, ctx context.Context) (err error) {
	sup.RLockUser(r, eid)
	defer sup.RUnlockUser(r, eid)
	err = c.SetUnsafe(eid, r, sup, ctx)
	return
}

func (c *CSecureContext) SetUnsafe(eid string, r *http.Request, sup feature.SiteUsersProvider, ctx context.Context) (err error) {
	sup.LockUser(r, eid)
	defer sup.UnlockUser(r, eid)

	var au feature.User
	if au, err = sup.RetrieveUser(r, eid); err != nil {
		return
	}

	var data []byte
	var encoded string
	if data, err = json.Marshal(ctx); err != nil {
		return
	} else if encoded, err = crypto.Encrypt(c._key, data); err != nil {
		return
	}

	auCtx := au.UnsafeContext()

	if err = auCtx.SetKV(".secure."+c._tag.Kebab(), encoded); err != nil {
		return
	}

	err = sup.SetUserContext(r, eid, auCtx)
	return
}

func (c *CSecureContext) Delete(eid string, r *http.Request, sup feature.SiteUsersProvider) (err error) {
	sup.LockUser(r, eid)
	defer sup.UnlockUser(r, eid)

	var au feature.User
	if au, err = sup.RetrieveUser(r, eid); err != nil {
		return
	}

	auCtx := au.UnsafeContext()
	auCtx.Delete(".secure." + c._tag.Kebab())

	err = sup.SetUserContext(r, eid, auCtx)
	return
}
