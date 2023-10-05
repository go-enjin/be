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
	"context"
	"net/http"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

var (
	VisitorUserName = "visitor"
)

const CurrentUserKey beContext.RequestKey = "current-user"

func IsValidEID(eid string) (valid bool) {
	valid = eid != "" && (eid == VisitorUserName || len(eid) == 10)
	return
}

func GetCurrentUserEID(r *http.Request) (eid string) {
	if user := GetCurrentUser(r); user != nil {
		eid = user.GetEID()
	} else {
		eid = VisitorUserName
	}
	return
}

func GetCurrentUser(r *http.Request) (u feature.User) {
	if v := r.Context().Value(CurrentUserKey); v != nil {
		u, _ = v.(feature.User)
	}
	return
}

func SetCurrentUser(u feature.User, r *http.Request) (modified *http.Request) {
	modified = r.Clone(context.WithValue(r.Context(), CurrentUserKey, u))
	return
}

func CurrentUserCan(r *http.Request, actions ...feature.Action) (allow bool) {
	if user := GetCurrentUser(r); user != nil {
		for _, action := range actions {
			if allow = user.Can(action); allow {
				break
			}
		}
	}
	return
}

func CurrentUserCanAll(r *http.Request, actions ...feature.Action) (allow bool) {
	if user := GetCurrentUser(r); user != nil {
		for _, action := range actions {
			if allow = user.Can(action); !allow {
				break
			}
		}
	}
	return
}

func RequireCurrentUser(enjin feature.Internals) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if user := GetCurrentUser(r); user == nil {
				enjin.ServeNotFound(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireUserCan(enjin feature.Internals, actions ...feature.Action) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var can bool
			if user := GetCurrentUser(r); user != nil {
				for _, action := range actions {
					if can = user.Can(action); can {
						log.WarnRF(r, "user has: %v", action)
						break
					}
				}
			} else {
				for _, action := range actions {
					if can = enjin.PublicUserActions().Has(action); can {
						log.WarnRF(r, "public user has: %v", action)
						break
					}
				}
			}
			if !can {
				log.WarnRF(r, "user missing any of: %+v", actions)
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
			if user := GetCurrentUser(r); user != nil {
				for _, action := range actions {
					if !user.Can(action) {
						log.WarnRF(r, "user missing action: %+v", action)
						enjin.ServeNotFound(w, r)
						return
					}
				}
			}
			for _, action := range actions {
				if !enjin.PublicUserActions().Has(action) {
					log.WarnRF(r, "public user missing action: %+v", action)
					enjin.ServeNotFound(w, r)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

//func RequireCurrentUserMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		if user := GetCurrentUser(r); user == nil {
//			serve.
//		}
//	})
//}