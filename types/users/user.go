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

package users

import (
	"encoding/json"
	"strings"

	sha "github.com/go-corelibs/shasum"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/userbase"
)

var (
	_ feature.User = (*User)(nil)
)

type User struct {
	RID         string          `json:"real-id"`
	EID         string          `json:"enjin-id"`
	Name        string          `json:"name"`
	Email       string          `json:"email"`
	Image       string          `json:"image"`
	Origin      string          `json:"origin"`
	Groups      feature.Groups  `json:"groups"`
	Actions     feature.Actions `json:"actions"`
	Context     context.Context `json:"context"`
	Active      bool            `json:"active"`
	AdminLocked bool            `json:"admin-locked"`
}

func NewUser(id, name, email, image string, ctx context.Context) (user *User) {
	eid, _ := sha.BriefSum([]byte(id))
	user = &User{
		RID:     id,
		EID:     eid,
		Name:    name,
		Email:   email,
		Image:   image,
		Context: ctx,
		Active:  true,
	}
	return
}

func (u *User) GetRID() (rid string) {
	rid = u.RID
	return
}

func (u *User) GetEID() (eid string) {
	eid = u.EID
	return
}

func (u *User) GetName() (name string) {
	name = u.Name
	return
}

func (u *User) GetEmail() (email string) {
	email = u.Email
	return
}

func (u *User) GetImage() (image string) {
	image = u.Image
	return
}

func (u *User) GetOrigin() (origin string) {
	origin = u.Origin
	return
}

func (u *User) UnsafeContext() (ctx context.Context) {
	ctx = u.Context
	return
}

func (u *User) GetActive() (active bool) {
	active = u.Active
	return
}

func (u *User) GetAdminLocked() (locked bool) {
	locked = u.AdminLocked
	return
}

func (u *User) SafeContext(includeKeys ...string) (ctx context.Context) {
	ctx = context.Context{
		"EID":          u.EID,
		"Name":         u.Name,
		"Email":        u.Email,
		"Image":        u.Image,
		"DisplayName":  u.Context.String(".settings.display-name", u.Name),
		"ProfileImage": u.Context.String(".settings.profile-image", u.Image),
	}
	for _, key := range includeKeys {
		if k, v := u.Context.GetKV(key); k != "" {
			// always filter out secure and settings user contexts
			if strings.ToLower(k) != "secure" && strings.ToLower(k) != "settings" {
				ctx[k] = v
			}
		}
	}
	ctx.CamelizeKeys()
	return
}

func (u *User) GetSettings(limitKeys ...string) (settings context.Context) {
	settings = context.Context{
		"Email":        u.Email,
		"DisplayName":  u.Context.String(".settings.display-name", u.Name),
		"ProfileImage": u.Context.String(".settings.profile-image", u.Image),
	}
	if ctx := u.Context.Context("settings"); len(ctx) > 0 {
		if len(limitKeys) > 0 {
			for _, key := range limitKeys {
				k, v := ctx.GetKV(key)
				settings[k] = v
			}
		} else {
			settings.ApplySpecific(ctx)
		}
	}
	settings.CamelizeKeys()
	return
}

func (u *User) GetSetting(key string) (value interface{}) {
	if ctx := u.Context.Context("settings"); len(ctx) > 0 {
		value = ctx.Get(key)
	}
	return
}

func (u *User) GetGroups() (groups feature.Groups) {
	groups = append(groups, u.Groups...)
	return
}

func (u *User) GetActions() (actions feature.Actions) {
	actions = append(actions, u.Actions...)
	return
}

func (u *User) IsVisitor() (visitor bool) {
	visitor = u.EID == userbase.VisitorEID
	return
}

func (u *User) Can(actions ...feature.Action) (allowed bool) {
	if u.AdminLocked {
		return
	}
	allowed = u.Actions.HasOneOf(actions)
	return
}

func (u *User) CanAll(actions ...feature.Action) (allowed bool) {
	if u.AdminLocked {
		return
	}
	allowed = u.Actions.HasAllOf(actions)
	return
}

func (u *User) Bytes() (data []byte) {
	var err error
	if data, err = json.MarshalIndent(u, "", "\t"); err != nil {
		panic(err)
	}
	return
}
