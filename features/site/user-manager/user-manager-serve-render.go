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

package user_manager

import (
	"net/http"

	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) RenderUserManager(w http.ResponseWriter, r *http.Request) {
	printer := message.GetPrinter(r)
	eid := userbase.GetCurrentEID(r)

	ctx := context.Context{
		"Title":      f.SiteFeatureLabel(printer),
		"FormAction": f.SiteFeaturePath(),
		"Nonces": feature.Nonces{
			{Name: UserManagerNonceName, Key: UserManagerNonceKey},
		},
	}

	su := f.Site().SiteUsers()
	list, total := su.ListUsers(r, 0, -1, false)
	ctx.SetSpecific("TotalUsers", total)
	var userList []context.Context
	for _, u := range list {
		email := u.GetEmail()
		uCtx := u.SafeContext()
		var actions editor.Actions
		notSelf := u.GetEID() != eid
		if userbase.CurrentUserCan(r, f.Action("update", "user")) {
			if u.GetActive() {
				actions = append(actions, editor.MakeDeactivateUser(printer, email))
			} else {
				actions = append(actions, editor.MakeReactivateUser(printer, email))
			}
			if notSelf {
				if u.GetAdminLocked() {
					actions = append(actions, editor.MakeAdminUnlockUser(printer, email))
				} else {
					actions = append(actions, editor.MakeAdminLockUser(printer, email))
				}
				if userbase.CurrentUserCan(r, f.Site().SiteAuth().Action("reset-other", "multi-factors")) {
					if f.Site().SiteAuth().NumFactorsPresent() > 0 {
						actions = append(actions, editor.MakeResetUserOtp(printer, email))
					}
				}
			}
		}
		if notSelf {
			if userbase.CurrentUserCan(r, f.Action("delete", "user")) {
				actions = append(actions, editor.MakeDeleteUser(printer, email))
			}
		}
		uCtx.SetSpecific("Info", &editor.File{
			FSBT:    su.BaseTag().Kebab(),
			FSID:    su.Tag().Kebab(),
			Code:    "user",
			Path:    "",
			File:    u.GetEID() + ".json",
			Name:    u.GetEID(),
			Actions: actions,
		})
		userList = append(userList, uCtx)
	}

	ctx.SetSpecific("UserList", userList)

	if userbase.CurrentUserCan(r, f.Action("create", "user")) {
		ctx.SetSpecific("CreateUserAction", editor.MakeCreateUser(printer))
	}

	t := f.SiteFeatureTheme()
	if err := f.Site().PrepareAndServePage("site", "user-manager", f.SiteFeaturePath(), t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing and serving %v site feature page: %v", f.Tag(), err)
		f.Enjin.ServeInternalServerError(w, r)
		return
	}
}
