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

package site_users

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/maruel/natural"

	clPath "github.com/go-corelibs/path"
	clStrings "github.com/go-corelibs/strings"
	beContext "github.com/go-enjin/be/pkg/context"
	beErrors "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/signals"
	"github.com/go-enjin/be/pkg/userbase"
	beUser "github.com/go-enjin/be/types/users"
)

func (f *CFeature) makeUserPath(eid string) (path string) {
	path = f.userPath + "/" + eid + ".json"
	return
}

func (f *CFeature) ListUsers(r *http.Request, pg, numPerPage int, sortDesc bool) (list []feature.User, total int) {
	for _, file := range f.MountPoints.ListFiles("/user") {
		eid := clPath.Base(file)
		if au, err := f.getUser(eid); err == nil {
			list = append(list, au)
		} else {
			log.ErrorRF(r, "error getting auth user: %v - %w", eid, err)
		}
	}
	sort.Slice(list, func(i, j int) (less bool) {
		a, b := list[i], list[j]
		if sortDesc {
			less = natural.Less(b.GetEmail(), a.GetEmail())
			return
		}
		less = natural.Less(a.GetEmail(), b.GetEmail())
		return
	})
	total = len(list)
	if numPerPage <= 0 || numPerPage > total {
		return
	}
	start := pg * numPerPage
	end := start + numPerPage
	if end >= total {
		end = total - 1
	}
	list = list[start:end]
	return
}

func (f *CFeature) UserPresent(eid string) (present bool) {
	present = f.MountPoints.Exists(f.makeUserPath(eid))
	return
}

func (f *CFeature) makeUser(origin, rid, eid, email string) (u *beUser.User, err error) {
	u = &beUser.User{
		RID:     rid,
		EID:     eid,
		Name:    clStrings.NameFromEmail(email),
		Email:   email,
		Origin:  origin,
		Active:  true,
		Context: beContext.Context{},
		Groups: feature.Groups{
			userbase.PublicGroup,
			userbase.UsersGroup,
		},
	}
	err = f.MountPoints.WriteFile(f.makeUserPath(u.GetEID()), u.Bytes())
	return
}

func (f *CFeature) SignUpUser(r *http.Request, claims *feature.CSiteAuthClaims) (err error) {
	if len(claims.Audience) == 0 {
		err = fmt.Errorf("claims.Audience[0] not found")
		return
	}

	if stop := f.Emit(signals.PreSignUpUser, f.Tag().String(), r, claims.Audience[0], claims.RID, claims.EID, claims.Email); stop {
		err = beErrors.ErrSignalStopped
		return
	} else if !userbase.CurrentUserCan(r, f.PermissionSignUpUser) {
		err = beErrors.ErrPermissionDenied
		return
	} else if f.UserPresent(claims.EID) {
		err = beErrors.ErrExistsAlready
		return
	}

	var u feature.User
	if u, err = f.makeUser(claims.Audience[0], claims.RID, claims.EID, claims.Email); err != nil {
		return
	}

	f.Emit(signals.PostSignUpUser, f.Tag().String(), r, u)
	return
}

func (f *CFeature) CreateUser(r *http.Request, origin, rid, eid, email string) (err error) {

	if stop := f.Emit(signals.PreCreateUser, f.Tag().String(), r, origin, rid, eid, email); stop {
		err = beErrors.ErrSignalStopped
		return
	} else if !userbase.CurrentUserCan(r, f.PermissionCreateUser) {
		err = beErrors.ErrPermissionDenied
		return
	} else if f.UserPresent(eid) {
		err = beErrors.ErrExistsAlready
		return
	}

	var u feature.User
	if u, err = f.makeUser(origin, rid, eid, email); err != nil {
		return
	}

	f.Emit(signals.PostCreateUser, f.Tag().String(), r, u)
	return
}

func (f *CFeature) RetrieveUser(r *http.Request, eid string) (user feature.User, err error) {
	//uid := userbase.GetCurrentEID(r)
	if stop := f.Emit(signals.PreRetrieveUser, f.Tag().String(), r, eid); stop {
		err = beErrors.ErrSignalStopped
		return
	}

	var au *beUser.User
	if au, err = f.getUser(eid); err == nil {

		au.Actions = au.Actions.Append(f.Enjin.PublicUserActions()...)

		for _, group := range au.Groups {
			if group != userbase.PublicGroup {
				if actions, ee := f.getGroup(group); ee == nil {
					au.Actions = au.Actions.Append(actions...)
				}
			}
		}

		user = au
	}

	f.Emit(signals.PostRetrieveUser, f.Tag().String(), r, user, err)
	return
}

func (f *CFeature) DeleteUser(r *http.Request, eid string) (err error) {
	uid := userbase.GetCurrentEID(r)

	var au *beUser.User
	if au, err = f.getUser(eid); err == nil {

		if stop := f.Emit(signals.PreDeleteUser, f.Tag().String(), r, eid, au); stop {
			err = beErrors.ErrSignalStopped
			return
		} else if !f.checkUserCan(r, uid, eid, f.PermissionDeleteOwn, f.PermissionDeleteOther) {
			err = beErrors.ErrPermissionDenied
			return
		}

		err = f.MountPoints.RemoveFile(f.makeUserPath(eid))

		f.Emit(signals.PostDeleteUser, f.Tag().String(), r, eid, au)
	}

	return
}
