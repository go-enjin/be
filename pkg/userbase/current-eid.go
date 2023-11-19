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
)

func IsVisitor(r *http.Request) (visiting bool) {
	visiting = GetCurrentEID(r) == VisitorEID
	return
}

func GetCurrentEID(r *http.Request) (eid string) {
	if currentEid, ok := r.Context().Value(gCurrentEidKey).(string); ok {
		eid = currentEid
	} else {
		eid = VisitorEID
	}
	return
}

func setCurrentEID(r *http.Request, eid interface{}) (modified *http.Request) {
	modified = r.Clone(context.WithValue(r.Context(), gCurrentEidKey, eid))
	return
}