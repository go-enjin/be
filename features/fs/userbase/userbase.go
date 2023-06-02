//go:build fs_userbase || all

// Copyright (c) 2022  The Go-Enjin Authors
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
	"encoding/json"
	"fmt"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/filesystem"
	"github.com/go-enjin/be/pkg/feature/signaling"
	"github.com/go-enjin/be/pkg/page/matter"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/userbase"
)

const Tag feature.Tag = "fs-userbases"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	filesystem.Feature[MakeFeature]

	signaling.SignalsSupport

	userbase.AuthUserProvider
	userbase.AuthUserManager
	userbase.UserProvider
	userbase.UserManager
	userbase.GroupsProvider
	userbase.GroupsManager
}

type MakeFeature interface {
	filesystem.MakeFeature[MakeFeature]

	// IncludeGroup adds an in-memory group to the userbase
	IncludeGroup(group userbase.Group, actions ...userbase.Action) MakeFeature

	// AddDefaultGroups adds the given groups to the list of default groups for
	// all new users
	AddDefaultGroups(groups ...userbase.Group) MakeFeature

	// SetDefaultGroups sets the given groups as the list of default groups for
	// all new users
	SetDefaultGroups(groups ...userbase.Group) MakeFeature

	// SetAuthPath overrides the default /auth path prefix for AuthUser storage
	SetAuthPath(path string) MakeFeature

	// SetUserPath overrides the default /user path prefix for User storage
	SetUserPath(path string) MakeFeature

	// SetGroupPath overrides the default /group path prefix for User Groups
	// storage
	SetGroupPath(path string) MakeFeature

	// SetGroupsPath overrides the default /groups path prefix for Group storage
	SetGroupsPath(path string) MakeFeature

	// SetUserBasePaths overrides the default userbase paths for all storage
	SetUserBasePaths(auth, user, group, groups string) MakeFeature

	Make() Feature
}

type CFeature struct {
	filesystem.CFeature[MakeFeature]
	signaling.CSignaling

	authPath string

	userPath string

	groupPath string

	groupsPath string

	fallbackGroups map[userbase.Group]userbase.Actions

	defaultGroups   userbase.Groups
	defaultLanguage language.Tag
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
	f.CSignaling.InitSignaling()
	f.authPath = "/auth"
	f.userPath = "/user"
	f.groupPath = "/group"
	f.groupsPath = "/groups"
	f.fallbackGroups = make(map[userbase.Group]userbase.Actions)
	f.defaultGroups = DefaultNewUserGroups
	f.defaultLanguage = language.Und
}

func (f *CFeature) IncludeGroup(group userbase.Group, actions ...userbase.Action) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.fallbackGroups[group] = append(f.fallbackGroups[group], actions...)
	return f
}

func (f *CFeature) AddDefaultGroups(groups ...userbase.Group) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.defaultGroups = append(f.defaultGroups, groups...)
	return f
}

func (f *CFeature) SetDefaultGroups(groups ...userbase.Group) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.defaultGroups = groups
	return f
}

func (f *CFeature) SetDefaultLanguage(tag language.Tag) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.defaultLanguage = tag
	return f
}

func (f *CFeature) SetAuthPath(path string) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.authPath = bePath.CleanWithSlash(path)
	return f
}

func (f *CFeature) SetUserPath(path string) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.userPath = bePath.CleanWithSlash(path)
	return f
}

func (f *CFeature) SetGroupPath(path string) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.groupPath = bePath.CleanWithSlash(path)
	return f
}

func (f *CFeature) SetGroupsPath(path string) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.groupsPath = bePath.CleanWithSlash(path)
	return f
}

func (f *CFeature) SetUserBasePaths(auth, user, group, groups string) MakeFeature {
	f.Lock()
	defer f.Unlock()
	f.authPath = bePath.CleanWithSlash(auth)
	f.userPath = bePath.CleanWithSlash(user)
	f.groupPath = bePath.CleanWithSlash(group)
	f.groupsPath = bePath.CleanWithSlash(groups)
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	f.Lock()
	defer f.Unlock()

	if f.defaultLanguage == language.Und {
		f.defaultLanguage = f.Enjin.SiteDefaultLanguage()
	}

	if found := f.FindPathMountPoint(f.authPath + "/"); len(found) == 0 {
		err = fmt.Errorf("%v feature error finding mount point for: auth - %v filesystems", f.Tag(), f.authPath)
		return
	}

	if found := f.FindPathMountPoint(f.userPath + "/"); len(found) == 0 {
		err = fmt.Errorf("%v feature error finding mount point for: user - %v filesystems", f.Tag(), f.userPath)
		return
	}

	if found := f.FindPathMountPoint(f.groupPath + "/"); len(found) == 0 {
		err = fmt.Errorf("%v feature error finding mount point for: group - %v filesystems", f.Tag(), f.groupPath)
		return
	}

	if found := f.FindPathMountPoint(f.groupsPath + "/"); len(found) == 0 {
		err = fmt.Errorf("%v feature error finding mount point for: groups - %v filesystems", f.Tag(), f.groupsPath)
		return
	}

	// filename := f.makeGroupActionsFilePath("players") + ".text"
	// for _, mp := range f.FindPathMountPoint(filename) {
	// 	if mp.ROFS.Exists(filename) {
	// 		log.WarnF("*** FOUND: %v - %#+v", filename, mp)
	// 	} else {
	// 		log.WarnF("NOT FOUND: %v - %#+v", filename, mp)
	// 	}
	// }
	//
	// groups := f.GetGroupActions("players")
	// log.ErrorF("actions for players group: %#+v", groups)
	// err = fmt.Errorf("stop")
	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) UserActions() (list userbase.Actions) {

	tag := f.Tag().Kebab()
	list = userbase.Actions{
		userbase.NewAction(tag, "view-own", "auth-user"),
		userbase.NewAction(tag, "view-other", "auth-user"),
		userbase.NewAction(tag, "edit-own", "auth-user"),
		userbase.NewAction(tag, "edit-other", "auth-user"),
		userbase.NewAction(tag, "delete-own", "auth-user"),
		userbase.NewAction(tag, "delete-other", "auth-user"),
		userbase.NewAction(tag, "view-own", "user"),
		userbase.NewAction(tag, "view-other", "user"),
		userbase.NewAction(tag, "edit-own", "user"),
		userbase.NewAction(tag, "edit-other", "user"),
		userbase.NewAction(tag, "delete-own", "user"),
		userbase.NewAction(tag, "delete-other", "user"),
	}

	return
}

func (f *CFeature) NewAuthUser(id, name, email, picture, audience string, attributes map[string]interface{}) (user *userbase.AuthUser, err error) {
	f.Lock()
	defer f.Unlock()
	if user, err = f.makeAuthUserUnsafe(id, name, email, picture, audience, attributes); err == nil {
		err = f.setAuthUserUnsafe(user)
	}
	return
}

func (f *CFeature) SetAuthUser(user *userbase.AuthUser) (err error) {
	f.Lock()
	defer f.Unlock()

	authUserFilename := f.makeAuthFilePath(user.EID)

	write := func(u *userbase.AuthUser) (err error) {
		if v, e := json.Marshal(u); e == nil {
			for _, mp := range f.FindPathMountPoint(authUserFilename) {
				if mp.RWFS != nil {
					err = mp.RWFS.WriteFile(authUserFilename, v, 0660)
					return
				}
			}
		}
		err = fmt.Errorf("read/write filesystem not found for: %v", authUserFilename)
		return
	}

	if u, e := f.getAuthUserUnsafe(user.EID); e == nil {
		var changed bool
		if u.Name != user.Name {
			changed = true
			u.Name = user.Name
		}
		if u.Email != user.Email {
			changed = true
			u.Email = user.Email
		}
		if u.Image != user.Image {
			changed = true
			u.Image = user.Image
		}
		if changed {
			err = write(u)
			return
		}
		// nop
		return
	}

	err = write(user)
	return
}

func (f *CFeature) AuthUserPresent(eid string) (present bool) {
	f.RLock()
	defer f.RUnlock()
	_, present = f.getAuthUserExistsUnsafe(eid)
	return
}

func (f *CFeature) GetAuthUser(eid string) (user *userbase.AuthUser, err error) {
	f.RLock()
	defer f.RUnlock()
	user, err = f.getAuthUserUnsafe(eid)
	return
}

func (f *CFeature) RemoveAuthUser(eid string) (err error) {
	f.Lock()
	defer f.Unlock()
	authUserFilename := f.makeAuthFilePath(eid)
	for _, mp := range f.FindPathMountPoint(authUserFilename) {
		if mp.RWFS != nil {
			err = mp.RWFS.Remove(authUserFilename)
			return
		}
	}
	err = fmt.Errorf("read/write filesystem not found for: %v", authUserFilename)
	return
}

func (f *CFeature) NewUser(au *userbase.AuthUser) (user *userbase.User, err error) {
	f.Lock()
	defer f.Unlock()

	if u, e := f.getUserUnsafe(au.EID); e == nil {
		// user exists already, set and load
		user = u
		return
	}

	var created bool
	if user, created, err = f.makeUserUnsafe(au); err != nil {
		err = fmt.Errorf("error making new user structure: %v - %v", au.EID, err)
		return
	}

	if err = f.setUserUnsafe(user); err != nil {
		err = fmt.Errorf("error saving new user: %v - %v", user.EID, err)
		return
	}

	if err = f.addUserToGroupsUnsafe(user.EID, f.defaultGroups...); err != nil {
		err = fmt.Errorf("error adding user to default groups: %v - %V", user.EID, err)
		return
	}

	if created {
		f.Emit(userbase.UserSignupSignal, f.Tag().String(), user)
	} else {
		f.Emit(userbase.UserLoginSignal, f.Tag().String(), user)
	}

	return
}

func (f *CFeature) SetUser(user *userbase.User) (err error) {
	f.Lock()
	defer f.Unlock()
	err = f.setUserUnsafe(user)
	return
}

func (f *CFeature) GetUser(eid string) (user *userbase.User, err error) {
	f.RLock()
	defer f.RUnlock()
	if user, err = f.getUserUnsafe(eid); err != nil {
		return
	}
	user.Groups = f.getUserGroupsUnsafe(user.EID)
	for _, group := range user.Groups {
		actions := f.GetGroupActions(group)
		user.Actions = user.Actions.Append(actions...)
	}
	return
}

func (f *CFeature) RemoveUser(id string) (err error) {
	f.Lock()
	defer f.Unlock()

	var path string
	userFilename := f.userPath + "/" + id
	if path, err = f.FindPageMatterPath(userFilename); err != nil {
		err = nil
		return
	}

	for _, mp := range f.FindPathMountPoint(path) {
		if mp.RWFS != nil {
			if err = mp.RWFS.Remove(path); err != nil {
				return
			}
		}
	}

	err = fmt.Errorf("read/write filesystem not found for: %v", userFilename)
	return
}

func (f *CFeature) GetGroupActions(group userbase.Group) (actions userbase.Actions) {
	// get all registered user actions
	allActions := f.Enjin.FindAllUserActions()
	// read the group actions file
	if _, parsed, exists := f.parseGroupActionsFile(group); !exists {
		// filter for only known actions from the parsed list
		actions = allActions.FilterKnown(parsed)
		// include any fallback group actions as this group has not actual file
		if fallbacks, ok := f.fallbackGroups[group]; ok {
			// filter for only known actions from the fallback list
			if known := allActions.FilterKnown(fallbacks); known.Len() > 0 {
				actions = actions.Append(known...)
			}
		}
		// include all public actions
		actions = actions.Append(f.Enjin.GetPublicAccess()...)
	} else {
		// filter for only known actions from the parsed list
		actions = allActions.FilterKnown(parsed)
	}
	return
}

func (f *CFeature) UpdateGroup(group userbase.Group, actions ...userbase.Action) (err error) {
	f.Lock()
	defer f.Unlock()

	var exists bool
	var pm *matter.PageMatter
	var existing userbase.Actions

	if pm, existing, exists = f.parseGroupActionsFile(group); !exists {
		// new group, check if there's any fallback actions ready
		if fallbacks, ok := f.fallbackGroups[group]; ok {
			existing = existing.Append(fallbacks...)
		}
	}

	// update with the new actions
	for _, action := range actions {
		if !existing.Has(action) {
			existing = existing.Append(action)
		}
	}

	newActionLookup := make(map[userbase.Action]int)
	for idx, action := range existing {
		newActionLookup[action] = idx + 1
	}

	pm.Matter["ActionLookup"] = newActionLookup
	pm.Body = existing.AsNewlines()

	err = f.WritePageMatter(pm)
	return
}

func (f *CFeature) RemoveGroup(group userbase.Group) (err error) {
	f.Lock()
	defer f.Unlock()
	target := f.makeGroupActionsFilePath(group)
	if path, ee := f.FindPageMatterPath(target); ee == nil {
		for _, mp := range f.FindPathMountPoint(path) {
			if mp.RWFS != nil {
				err = mp.RWFS.Remove(path)
				return
			}
		}
	}
	err = fmt.Errorf("read/write filesystem not found for: %v", target)
	return
}

func (f *CFeature) AddUserToGroup(eid string, groups ...userbase.Group) (err error) {
	f.Lock()
	defer f.Unlock()
	err = f.addUserToGroupsUnsafe(eid, groups...)
	return
}

func (f *CFeature) RemoveUserFromGroup(eid string, groups ...userbase.Group) (err error) {
	f.Lock()
	defer f.Unlock()
	err = f.removeUserFromGroupsUnsafe(eid, groups...)
	return
}

func (f *CFeature) IsUserInGroup(eid string, group userbase.Group) (present bool) {
	f.RLock()
	defer f.RUnlock()
	present = f.isUserInGroupUnsafe(eid, group)
	return
}

func (f *CFeature) GetUserGroups(eid string) (groups userbase.Groups) {
	f.RLock()
	defer f.RUnlock()
	groups = f.getUserGroupsUnsafe(eid)
	return
}

// func (f *CFeature) ProcessRequestPageType(r *http.Request, p *page.Page) (pg *page.Page, redirect string, processed bool, err error) {
// 	switch strings.ToLower(p.Type) {
// 	case "user":
// 		// these are individual user pages
// 	case "group":
// 		// these are the users associated with one or more groups
// 	case "groups":
// 		// these are group definitions
// 	}
// 	return
// }