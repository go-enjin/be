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
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/userbase"
)

const Tag feature.Tag = "user-base-htenv"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	userbase.SecretsProvider
	userbase.UsersProvider
	userbase.GroupsProvider
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature

	tag feature.Tag

	cliCtx *cli.Context
	enjin  feature.Internals

	users  map[string]*userbase.User
	hashes map[string]string
	groups map[string][]string
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = Tag
	return f
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.FeatureTag = Tag
	f.users = make(map[string]*userbase.User)
	f.hashes = make(map[string]string)
	f.groups = make(map[string][]string)
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	tag := f.Tag().String()
	b.AddFeatureNotes(
		f.Tag(),
		"this feature looks for dynamic environment variable keys",
		"YOUR_NAME translates to the user name \"your-name\"",
		"GROUP_NAME translates to the group name \"group-name\"",
		"add users: "+globals.MakeFlagEnvKey(tag, "USER_YOUR_NAME")+"='<bcrypt-password-hash>'",
		"add groups: "+globals.MakeFlagEnvKey(tag, "GROUP_GROUP_NAME")+"='<space separated usernames>'",
		"use 'enjenv be-crypt <password>' to generate user passwords",
	)
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.enjin = enjin
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	f.cliCtx = ctx
	err = f.loadEnvironment()
	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) GetUserSecret(id string) (secret string) {
	secret, _ = f.hashes[id]
	return
}

func (f *CFeature) AddUser(user *userbase.User) (err error) {
	// f.Lock()
	// defer f.Unlock()
	err = fmt.Errorf("cannot add user: %v is read-only", f.Tag())
	return
}

func (f *CFeature) AddUserToGroup(id string, groups ...string) (err error) {
	// f.Lock()
	// defer f.Unlock()
	err = fmt.Errorf("cannot add user to group(s): %v is read-only", f.Tag())
	return
}

func (f *CFeature) GetUser(id string) (user *userbase.User, err error) {
	if u, ok := f.users[id]; ok {
		user = u
	} else {
		err = fmt.Errorf("user not found")
	}
	return
}

func (f *CFeature) IsUserInGroup(id string, group string) (present bool) {
	if users, ok := f.groups[group]; ok {
		if beStrings.StringInSlices(id, users) {
			present = true
		}
	}
	return
}

func (f *CFeature) GetUserGroups(id string) (groups []string) {
	for group, users := range f.groups {
		if beStrings.StringInSlices(id, users) {
			groups = append(groups, group)
		}
	}
	return
}

var rxSplitEquals = regexp.MustCompile(`^\s*([^=]+?)\s*=\s*(.+?)\s*$`)

func (f *CFeature) loadEnvironment() (err error) {
	environ := os.Environ()

	var loadedUsers []string
	for _, env := range environ {
		if parts := rxSplitEquals.FindStringSubmatch(env); len(parts) == 3 {
			if name, hash, ok := f.parseEnvUser(parts[1], parts[2]); ok {
				f.hashes[name] = hash
				f.users[name] = userbase.NewUser(name, name, "", "")
				log.DebugF("including user: %v=%v", name, f.hashes[name])
				loadedUsers = append(loadedUsers, name)
			}
		}
	}

	var loadedGroups []string
	for _, env := range environ {
		if parts := rxSplitEquals.FindStringSubmatch(env); len(parts) == 3 {
			if name, users, ok := f.parseEnvGroup(parts[1], parts[2]); ok {
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

func (f *CFeature) parseEnvGroup(key, value string) (name string, users []string, ok bool) {
	prefix := globals.MakeFlagEnvKey(f.Tag().String(), "GROUP") + "_"
	prefix = strcase.ToKebab(prefix)
	prefixLen := len(prefix)
	name = strcase.ToKebab(key)
	if ln := len(name); prefixLen < ln {
		if name[0:prefixLen] == prefix {
			name = name[prefixLen:]
			for _, user := range strings.Split(value, " ") {
				user = strcase.ToKebab(user)
				if !beStrings.StringInStrings(user, users...) {
					users = append(users, user)
				}
			}
			ok = len(users) > 0
			// log.DebugF("group: %v, users: %v", name, users)
		}
	}
	return
}