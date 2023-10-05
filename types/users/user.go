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
	"fmt"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/types/page"
	"github.com/go-enjin/be/types/page/matter"
)

var (
	_ feature.User = (*User)(nil)
)

type User struct {
	page.CPage
	AuthUser

	Origin string `json:"origin"`

	Groups  feature.Groups  `json:"-"`
	Actions feature.Actions `json:"-"`
}

func NewUserFromPageMatter(user *AuthUser, pm *matter.PageMatter, formats feature.PageFormatProvider, enjin context.Context) (u *User, err error) {
	var pg *page.CPage
	if pg, err = page.NewFromPageMatter(pm, formats, enjin); err != nil {
		err = fmt.Errorf("error creating page from given page matter: %v", err)
		return
	}
	rid, eid := user.RID, user.EID
	pg.Context().SetSpecific("RID", rid)
	pg.Context().SetSpecific("EID", eid)
	pg.PageMatter().Matter.SetSpecific("RID", rid)
	pg.PageMatter().Matter.SetSpecific("EID", eid)
	u = &User{
		Origin:   pm.Origin,
		CPage:    *pg,
		AuthUser: *user,
	}
	return
}

func (u *User) GetOrigin() (origin string) {
	origin = u.Origin
	return
}

func (u *User) GetGroups() (groups feature.Groups) {
	groups = u.Groups
	return
}

func (u *User) GetActions() (actions feature.Actions) {
	actions = u.Actions
	return
}

func (u *User) AsPage() feature.Page {
	return &u.CPage
}

func (u *User) Can(action feature.Action) (allowed bool) {
	allowed = u.Actions.Has(action)
	return
}

func (u *User) FilteredContext(includeKeys ...string) (ctx context.Context) {
	ctx = context.Context{}
	ctx["EID"] = u.EID
	ctx["Name"] = u.Name
	ctx["Email"] = u.Email
	ctx["Image"] = u.Image
	for _, key := range includeKeys {
		ctx[key] = u.CPage.Context().Get(key)
	}
	return
}