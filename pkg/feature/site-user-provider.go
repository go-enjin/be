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
	"net/http"

	beContext "github.com/go-enjin/be/pkg/context"
)

type SiteUsersProvider interface {
	Feature

	MakeRealID(email string) (rid string)
	MakeEnjinID(rid string) (eid string)

	GroupPresent(group Group) (present bool)
	CreateGroup(r *http.Request, group Group, permissions ...Action) (err error)
	RetrieveGroup(r *http.Request, group Group) (permissions Actions, err error)
	UpdateGroup(r *http.Request, group Group, permissions ...Action) (err error)
	DeleteGroup(r *http.Request, group Group) (err error)

	UserPresent(eid string) (present bool)

	IsUserLocked(r *http.Request, eid string) (locked bool)
	LockUser(r *http.Request, eid string)
	UnlockUser(r *http.Request, eid string)
	RLockUser(r *http.Request, eid string)
	RUnlockUser(r *http.Request, eid string)

	ListUsers(r *http.Request, pg, numPerPage int, sortDesc bool) (list []AuthUser, total int)
	SignUpUser(r *http.Request, claims *CSiteAuthClaims) (err error)
	CreateUser(r *http.Request, origin, rid, eid, email string) (err error)
	RetrieveUser(r *http.Request, eid string) (user AuthUser, err error)
	DeleteUser(r *http.Request, eid string) (err error)

	UpdateUserName(r *http.Request, eid string, name string) (err error)
	UpdateUserImage(r *http.Request, eid string, image string) (err error)
	UpdateUserContext(r *http.Request, eid string, ctx beContext.Context) (err error)
	UpdateUserGroups(r *http.Request, eid string, groups ...Group) (err error)
	UpdateUserPermissions(r *http.Request, eid string, permissions ...Action) (err error)

	SetUserName(r *http.Request, eid string, name string) (err error)
	SetUserImage(r *http.Request, eid string, image string) (err error)
	SetUserContext(r *http.Request, eid string, ctx beContext.Context) (err error)
	SetUserSetting(r *http.Request, eid string, key string, value interface{}) (err error)
	SetUserSettings(r *http.Request, eid string, ctx beContext.Context) (err error)
	SetUserGroups(r *http.Request, eid string, groups ...Group) (err error)
	SetUserPermissions(r *http.Request, eid string, permissions ...Action) (err error)

	UpdateUserActive(r *http.Request, eid string, active bool) (err error)
	UpdateUserAdminLocked(r *http.Request, eid string, active bool) (err error)
	SetUserActive(r *http.Request, eid string, active bool) (err error)
	SetUserAdminLocked(r *http.Request, eid string, active bool) (err error)

	GetUserStatus(r *http.Request, eid string) (active, locked, visitor bool, err error)
	GetUserActive(r *http.Request, eid string) (active bool, err error)
	GetUserAdminLocked(r *http.Request, eid string) (locked bool, err error)
}