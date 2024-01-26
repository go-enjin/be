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

package userbase

import (
	"net/http"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

var (
	VisitorEID  = "visitor"
	VisitorName = "Visitor"

	UsersGroup  feature.Group = "users"
	PublicGroup feature.Group = "public"

	UserActiveKey      = "user-active"
	UserAdminLockedKey = "user-admin-locked"
)

func IsValidEID(eid string) (valid bool) {
	valid = eid != "" && (eid == VisitorEID || len(eid) == 10)
	return
}

func IsUserActive(au feature.User) (active bool) {
	if adminLocked, ok := au.UnsafeContext().Get(UserAdminLockedKey).(bool); ok && adminLocked {
		return
	}
	if v, ok := au.UnsafeContext().Get(UserActiveKey).(bool); ok {
		active = v
	}
	return
}

func CurrentUserCan(r *http.Request, actions ...feature.Action) (allow bool) {
	if permissions := GetCurrentPermissions(r); permissions.Len() > 0 {
		for _, action := range actions {
			if allow = permissions.Has(action); allow {
				log.DebugRF(r, "current user can: %q", action)
				return
			}
		}
	}
	log.DebugRF(r, "current user has none of: %+v", actions)
	return
}

func CurrentUserCanAll(r *http.Request, actions ...feature.Action) (allow bool) {
	if user := GetCurrentPermissions(r); user != nil {
		for _, action := range actions {
			if allow = user.Has(action); !allow {
				log.DebugRF(r, "current cannot: %q", action)
				return
			}
		}
	}
	log.DebugRF(r, "current user has all of: %+v", actions)
	return
}

func RequireUserCan(enjin feature.Internals, actions ...feature.Action) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var can bool
			var permissions feature.Actions
			if permissions = GetCurrentPermissions(r); permissions.Len() == 0 {
				permissions = enjin.PublicUserActions()
			}
			for _, action := range actions {
				if can = permissions.Has(action); can {
					log.WarnRF(r, "current user can: %v", action)
					break
				}
			}
			if !can {
				log.WarnRF(r, "current user has none of: %+v", actions)
				enjin.ServeNotFound(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireUserCanAll(enjin feature.Internals, actions ...feature.Action) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var permissions feature.Actions
			if permissions = GetCurrentPermissions(r); permissions.Len() > 0 {
				permissions = enjin.PublicUserActions()
			}
			for _, action := range actions {
				if !permissions.Has(action) {
					log.WarnRF(r, "current user has no permission to: %+v", action)
					enjin.ServeNotFound(w, r)
					return
				}
			}
			log.WarnRF(r, "current user has all of: %+v", actions)
			next.ServeHTTP(w, r)
		})
	}
}
