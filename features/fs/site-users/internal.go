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
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mrz1836/go-sanitize"

	"github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/userbase"
	beUser "github.com/go-enjin/be/types/users"
)

func (f *CFeature) loadEnvironment() (err error) {

	if named, ok := f.env.GetSiteEnviron("init-group"); ok {
		for name, value := range named {
			actions := feature.ParseActions(value)
			group := feature.Group(name)
			f.InitGroup(group, actions...)
		}
	}

	if initUserEmails, ok := f.env.GetSiteEnviron("init-user-email"); ok {
		if initUserGroups, ok := f.env.GetSiteEnviron("init-user-group"); ok {
			for key, spacedEmails := range initUserEmails {
				if spacedGroups, ok := initUserGroups[key]; ok {
					var groups feature.Groups
					for _, name := range strings.Split(spacedGroups, " ") {
						if name != "" {
							groups = groups.Append(feature.Group(strings.ToLower(name)))
						}
					}
					for _, value := range strings.Split(spacedEmails, " ") {
						if email := sanitize.Email(value, false); email != "" {
							f.InitUser(email, groups...)
						}
					}
				}
			}
		}
	}

	return
}

func (f *CFeature) getGroup(group feature.Group) (permissions feature.Actions, err error) {
	if !f.GroupPresent(group) {
		err = errors.ErrGroupNotFound
		return
	}
	var data []byte
	if data, err = f.MountPoints.ReadFile(f.makeGroupPath(group)); err != nil {
		return
	}
	permissions = feature.Actions{}
	err = json.Unmarshal(data, &permissions)
	return
}

func (f *CFeature) setGroup(group feature.Group, permissions ...feature.Action) (err error) {
	actions := feature.Actions(permissions)
	err = f.MountPoints.WriteFile(f.makeGroupPath(group), actions.Bytes())
	return
}

func (f *CFeature) getUser(eid string) (user *beUser.User, err error) {
	if !f.UserPresent(eid) {
		err = errors.ErrUserNotFound
		return
	}
	var data []byte
	if data, err = f.MountPoints.ReadFile(f.makeUserPath(eid)); err != nil {
		return
	}
	user = &beUser.User{}
	err = json.Unmarshal(data, user)
	return
}

func (f *CFeature) setUser(au *beUser.User) (err error) {
	if !f.UserPresent(au.EID) {
		err = errors.ErrUserNotFound
		return
	}
	err = f.MountPoints.WriteFile(f.makeUserPath(au.EID), au.Bytes())
	return
}

func (f *CFeature) getCheckUserPerm(uid, eid string, self, other feature.Action) (need feature.Action) {
	if uid == eid {
		need = self
	} else {
		need = other
	}
	return
}

func (f *CFeature) checkUserCan(r *http.Request, uid, eid string, self, other feature.Action, more ...feature.Action) (allowed bool) {
	actions := feature.Actions{
		f.getCheckUserPerm(uid, eid, f.PermissionUpdateOwn, f.PermissionUpdateOther),
	}
	allowed = userbase.CurrentUserCanAll(r, actions.Append(more...)...)
	return
}