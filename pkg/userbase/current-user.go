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

func GetCurrentUser(r *http.Request) (u feature.User) {
	if v := r.Context().Value(gCurrentUserKey); v != nil {
		u, _ = v.(feature.User)
	}
	return
}

func SetCurrentUser(r *http.Request, u feature.User) (modified *http.Request) {
	modified = r.Clone(context.WithValue(r.Context(), gCurrentUserKey, u))
	modified = AppendCurrentPermissions(modified, u.GetActions()...)
	return
}
