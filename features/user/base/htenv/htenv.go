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
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/slices"
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
	userbase.AuthUserProvider
	userbase.GroupsProvider
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature

	users  map[string]*userbase.AuthUser
	hashes map[string]string
	groups map[userbase.Group][]string
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.users = make(map[string]*userbase.AuthUser)
	f.hashes = make(map[string]string)
	f.groups = make(map[userbase.Group][]string)
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
	f.CFeature.Setup(enjin)
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	err = f.loadEnvironment()
	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) GetUserSecret(id string) (secret string) {
	f.RLock()
	defer f.RUnlock()
	secret, _ = f.hashes[id]
	return
}

func (f *CFeature) SetUser(user userbase.AuthUser) (err error) {
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

func (f *CFeature) AuthUserPresent(id string) (present bool) {
	f.RLock()
	defer f.RUnlock()
	_, present = f.users[id]
	return
}

func (f *CFeature) GetAuthUser(id string) (user *userbase.AuthUser, err error) {
	f.RLock()
	defer f.RUnlock()
	if u, ok := f.users[id]; ok {
		user = u
	} else {
		err = fmt.Errorf("user not found")
	}
	return
}

func (f *CFeature) IsUserInGroup(id string, group userbase.Group) (present bool) {
	f.RLock()
	defer f.RUnlock()
	if users, ok := f.groups[group]; ok {
		if slices.Within(id, users) {
			present = true
		}
	}
	return
}

func (f *CFeature) GetUserGroups(id string) (groups userbase.Groups) {
	f.RLock()
	defer f.RUnlock()
	for group, users := range f.groups {
		if slices.Within(id, users) {
			groups = append(groups, group)
		}
	}
	return
}