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
	"os"
	"strings"

	"github.com/iancoleman/strcase"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/slices"
	"github.com/go-enjin/be/types/users"
)

func (f *CFeature) loadEnvironment() (err error) {
	environ := os.Environ()

	var loadedUsers []string
	for _, env := range environ {
		if parts := strings.SplitN(env, "=", 2); len(parts) == 2 {
			if name, hash, ok := f.parseEnvUser(parts[0], parts[1]); ok {
				f.hashes[name] = hash
				f.users[name] = users.NewAuthUser(name, name, "", "", beContext.Context{})
				log.DebugF("including user: %v=%v", name, f.hashes[name])
				loadedUsers = append(loadedUsers, name)
			}
		}
	}

	var loadedGroups feature.Groups
	for _, env := range environ {
		if parts := strings.SplitN(env, "=", 2); len(parts) == 2 {
			if name, users, ok := f.parseEnvGroup(parts[0], parts[1]); ok {
				f.groups[name] = append(f.groups[name], users...)
				loadedGroups = append(loadedGroups, name)
			}
		}
	}

	log.DebugF("found %d env users: %v", len(loadedUsers), loadedUsers)
	log.DebugF("found %d env groups: %v", len(loadedGroups), loadedGroups)
	return
}

func (f *CFeature) parseEnvUser(key, value string) (name, password string, ok bool) {
	prefix := globals.MakeFlagEnvKey(f.Tag().String(), "USER") + "_"
	prefix = strcase.ToKebab(prefix)
	prefixLen := len(prefix)
	name = strcase.ToKebab(key)
	if ln := len(name); prefixLen < ln {
		if name[0:prefixLen] == prefix {
			name = name[prefixLen:]
			password = value
			ok = true
			// log.DebugF("user: %v, pass: %v", name, password)
		}
	}
	return
}

func (f *CFeature) parseEnvGroup(key, value string) (group feature.Group, users []string, ok bool) {
	prefix := globals.MakeFlagEnvKey(f.Tag().String(), "GROUP") + "_"
	prefix = strcase.ToKebab(prefix)
	name := strcase.ToKebab(key)
	if strings.HasPrefix(name, prefix) {
		name = strings.TrimPrefix(name, prefix)
		group = feature.NewGroup(name)
		for _, user := range strings.Split(value, " ") {
			user = strcase.ToKebab(user)
			if !slices.Present(user, users...) {
				users = append(users, user)
			}
		}
		ok = len(users) > 0
		// log.DebugF("group: %v, users: %v", name, users)
	}
	return
}