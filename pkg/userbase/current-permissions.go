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

func GetCurrentPermissions(r *http.Request) (actions feature.Actions) {
	if u := GetCurrentAuthUser(r); u != nil {
		actions = u.GetActions()
	}

	if additional, ok := r.Context().Value(gCurrentPermissionsKey).(feature.Actions); ok {
		for _, action := range additional {
			if actions.Has(action) {
				continue
			}
			actions = append(actions, action)
		}
	}
	return
}

func SetCurrentPermissions(r *http.Request, actions ...feature.Action) (modified *http.Request) {
	modified = r.Clone(context.WithValue(r.Context(), gCurrentPermissionsKey, feature.Actions(actions)))
	return
}

func AppendCurrentPermissions(r *http.Request, actions ...feature.Action) (modified *http.Request) {
	var updated feature.Actions
	if existing, ok := r.Context().Value(gCurrentPermissionsKey).(feature.Actions); ok {
		updated = existing
	}
	for _, action := range actions {
		if updated.Has(action) {
			continue
		}
		updated = append(updated, action)
	}
	modified = SetCurrentPermissions(r, updated...)
	return
}
