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

	"github.com/go-enjin/be/pkg/feature"
)

func GetCurrentAuthUser(r *http.Request) (u feature.AuthUser) {
	if v := r.Context().Value(gCurrentAuthUserKey); v != nil {
		u, _ = v.(feature.AuthUser)
	}
	return
}

func SetCurrentAuthUser(u feature.AuthUser, r *http.Request) (m *http.Request) {
	if m = r.Clone(context.WithValue(r.Context(), gCurrentAuthUserKey, u)); u == nil {
		m = setCurrentEID(m, nil)
	} else {
		m = setCurrentEID(m, u.GetEID())
	}
	return
}