//go:build user_base_htenv || user_bases || all

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

package htenv

import (
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/go-corelibs/slices"
	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/types/users"
)

func (f *CFeature) loadEnvironment() (err error) {

	if named, ok := f.env.GetSiteEnviron("user"); ok {
		for name, hash := range named {
			f.hashes[name] = hash
			f.users[name] = users.NewUser(f.Tag().String()+"--"+name, name, "", "", beContext.Context{})
		}
	}

	if named, ok := f.env.GetSiteEnviron("group"); ok {
		for groupName, userList := range named {
			group := feature.Group(groupName)
			for _, username := range strings.Split(userList, " ") {
				if username = strcase.ToKebab(username); username != "" {
					if !slices.Within(username, f.groups[group]) {
						f.groups[group] = append(f.groups[group], username)
					}
				}
			}
		}
	}

	usernames := maps.SortedKeys(f.users)
	groupNames := maps.SortedKeys(f.groups)
	log.DebugDF(1, "found %d env users: %v", len(usernames), usernames)
	log.DebugDF(1, "found %d env groups: %v", len(groupNames), groupNames)
	return
}
