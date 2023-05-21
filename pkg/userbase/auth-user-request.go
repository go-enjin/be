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
)

const CurrentAuthUserKey beContext.RequestKey = "current-auth-user"

func GetCurrentAuthUser(r *http.Request) (u *AuthUser) {
	if v := r.Context().Value(CurrentAuthUserKey); v != nil {
		u, _ = v.(*AuthUser)
	}
	return
}

func SetCurrentAuthUser(u *AuthUser, r *http.Request) (modified *http.Request) {
	modified = r.Clone(context.WithValue(r.Context(), CurrentAuthUserKey, u))
	return
}