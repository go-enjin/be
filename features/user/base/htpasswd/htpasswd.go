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

package htpasswd

import (
	"fmt"

	fHtpasswd "github.com/foomo/htpasswd"
	"github.com/tg123/go-htpasswd"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/userbase"
)

const Tag feature.Tag = "user-base-htpasswd"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	userbase.UsersProvider
	userbase.GroupsProvider
	userbase.SecretsProvider
}

type MakeFeature interface {
	Make() Feature

	AddHTPasswdFile(filepath string) MakeFeature
	AddHTGroupsFile(filepath string) MakeFeature
}

type CFeature struct {
	feature.CFeature

	htpasswd map[string]*htpasswd.File
	htgroups map[string]*htpasswd.HTGroup

	parsedPwd map[string]fHtpasswd.HashedPasswords
	parsedGrp map[string]*htgroups
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.htpasswd = make(map[string]*htpasswd.File)
	f.htgroups = make(map[string]*htpasswd.HTGroup)

	f.parsedPwd = make(map[string]fHtpasswd.HashedPasswords)
	f.parsedGrp = make(map[string]*htgroups)
}

func (f *CFeature) AddHTPasswdFile(filepath string) MakeFeature {
	var err error
	if f.htpasswd[filepath], err = htpasswd.New(filepath, htpasswd.DefaultSystems, nil); err != nil {
		log.FatalDF(1, "error loading htpasswd file: %v", err)
	}
	if f.parsedPwd[filepath], err = fHtpasswd.ParseHtpasswdFile(filepath); err != nil {
		log.FatalDF(1, "error parsing htpasswd file: %v", err)
	}
	return f
}

func (f *CFeature) AddHTGroupsFile(filepath string) MakeFeature {
	var err error
	if f.htgroups[filepath], err = htpasswd.NewGroups(filepath, nil); err != nil {
		log.FatalDF(1, "error loading htgroups file: %v", err)
	}
	if f.parsedGrp[filepath], err = newHtgroups(filepath); err != nil {
		log.FatalDF(1, "error parsing htpasswd file: %v", err)
	}
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	tag := f.Tag().String()
	b.AddFlags(
		&cli.StringSliceFlag{
			Name:     globals.MakeFlagName(tag, "htpasswd"),
			Usage:    "include one or more htpasswd files",
			EnvVars:  globals.MakeFlagEnvKeys(tag, "HTPASSWD"),
			Category: tag,
		},
		&cli.StringSliceFlag{
			Name:     globals.MakeFlagName(tag, "htgroups"),
			Usage:    "include one or more htgroups files",
			EnvVars:  globals.MakeFlagEnvKeys(tag, "HTGROUPS"),
			Category: tag,
		},
	)
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	tag := f.Tag().String()

	if flagName := globals.MakeFlagName(tag, "htpasswd"); ctx.IsSet(flagName) {
		if list := ctx.StringSlice(flagName); len(list) > 0 {
			for _, filepath := range list {
				f.AddHTPasswdFile(filepath)
			}
		}
	}

	if flagName := globals.MakeFlagName(tag, "htgroups"); ctx.IsSet(flagName) {
		if list := ctx.StringSlice(flagName); len(list) > 0 {
			for _, filepath := range list {
				f.AddHTGroupsFile(filepath)
			}
		}
	}

	if len(f.htpasswd) == 0 {
		err = fmt.Errorf("at least one htpasswd file is required")
	}
	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) GetUserSecret(id string) (hash string) {
	f.RLock()
	defer f.RUnlock()
	if u, ok := f.parsedPwd[id]; ok {
		hash = string(u.Bytes())
	}
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
	f.RLock()
	defer f.RUnlock()
	if _, found := f.parsedPwd[id]; found {
		user = userbase.NewUser(id, id, "", "")
	} else {
		err = fmt.Errorf("user not found")
	}
	return
}

func (f *CFeature) IsUserInGroup(id string, group string) (present bool) {
	f.RLock()
	defer f.RUnlock()
	for _, htg := range f.htgroups {
		if htg.IsUserInGroup(id, group) {
			present = true
			return
		}
	}
	return
}

func (f *CFeature) GetUserGroups(id string) (groups []string) {
	f.RLock()
	defer f.RUnlock()
	for _, htg := range f.htgroups {
		if userGroups := htg.GetUserGroups(id); len(userGroups) > 0 {
			groups = append(groups, userGroups...)
			return
		}
	}
	return
}