//go:build fs_userbase || all

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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/types/page"
	"github.com/go-enjin/be/types/page/matter"
	"github.com/go-enjin/be/types/users"
)

func (f *CFeature) makeAuthFilePath(eid string) (path string) {
	path = f.authPath + "/" + eid + ".auth.json"
	return
}

func (f *CFeature) getAuthUserExistsUnsafe(eid string) (path string, ok bool) {
	path = f.makeAuthFilePath(eid)
	for _, mp := range f.FindPathMountPoint(path) {
		if ok = mp.ROFS.Exists(path); ok {
			return
		}
	}
	return
}

func (f *CFeature) getUserExistsUnsafe(eid string) (path string, ok bool) {
	var err error
	userPath := f.userPath + "/" + eid
	path, err = f.FindPageMatterPath(userPath)
	ok = err == nil
	return
}

func (f *CFeature) getAuthUserUnsafe(eid string) (user *users.AuthUser, err error) {
	if path, ok := f.getAuthUserExistsUnsafe(eid); ok {

		var data []byte
		for _, mp := range f.FindPathMountPoint(path) {
			if data, err = mp.ROFS.ReadFile(path); err != nil && err != os.ErrNotExist {
				err = fmt.Errorf("error reading auth user file: %v - %v", path, err)
				return
			} else if err == nil {
				break
			}
		}

		var u users.AuthUser
		if err = json.Unmarshal(data, &u); err != nil {
			err = fmt.Errorf("error parsing auth user data: %v - %v", path, err)
			return
		}

		u.Origin = f.Tag().String()
		user = &u
		return
	}

	err = os.ErrNotExist
	return
}

func (f *CFeature) getUserUnsafe(eid string) (user *users.User, err error) {
	var au *users.AuthUser
	if au, err = f.getAuthUserUnsafe(eid); err != nil {
		if err != os.ErrNotExist {
			err = fmt.Errorf("error looking up auth user: %v - %v", eid, err)
		}
		return
	}

	if path, ok := f.getUserExistsUnsafe(eid); ok {
		var pm *matter.PageMatter
		if pm, err = f.ReadPageMatter(path); err != nil {
			if err != os.ErrNotExist {
				err = fmt.Errorf("error reading user page matter: %v - %v", path, err)
			}
			return
		}

		user, err = users.NewUserFromPageMatter(au, pm, f.Enjin.MustGetTheme(), f.Enjin.Context())
		return
	}

	err = os.ErrNotExist
	return
}

func (f *CFeature) makeAuthUserUnsafe(id, name, email, picture, audience string, attributes map[string]interface{}) (user *users.AuthUser, err error) {
	tag := f.Tag().Camel()
	makeCtxKey := func(camel string) (key string) {
		key = tag + camel
		return
	}
	eid, _ := sha.DataHash10([]byte(id))
	if u, e := f.getAuthUserUnsafe(eid); e == nil && u != nil {
		user = u
		user.RID = id
		user.EID = eid
		user.Name = name
		user.Email = email
		user.Image = picture
		user.Origin = f.Tag().String()
		user.Context.SetSpecific(makeCtxKey("Audience"), audience)
		user.Context.SetSpecific(makeCtxKey("Attributes"), beContext.NewFromMap(attributes))
		return
	}
	user = users.NewAuthUser(id, name, email, picture, beContext.Context{
		makeCtxKey("Audience"):   audience,
		makeCtxKey("Attributes"): beContext.NewFromMap(attributes),
	})
	user.Origin = f.Tag().String()
	return
}

func (f *CFeature) setAuthUserUnsafe(user *users.AuthUser) (err error) {

	authUserFilename := f.makeAuthFilePath(user.EID)

	write := func(u *users.AuthUser) (err error) {
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

func (f *CFeature) makeUserUnsafe(au *users.AuthUser) (user *users.User, created bool, err error) {

	if u, e := f.getUserUnsafe(au.EID); e == nil && u != nil {
		user = u
		user.RID = au.RID
		user.EID = au.EID
		user.Origin = f.Tag().String()
		user.AuthUser = *au
		return
	}

	created = true

	realpath := f.userPath + "/" + au.EID + "." + DefaultNewUserPageFormat
	pm := matter.NewPageMatter(f.Tag().String(), realpath, DefaultNewUserPageBody, matter.JsonMatter, beContext.Context{
		"Name": au.Name,
	})

	var uu *users.User
	if uu, err = users.NewUserFromPageMatter(au, pm, f.Enjin.MustGetTheme(), f.Enjin.Context()); err != nil {
		err = fmt.Errorf("error constructing user from PageMatter: %v - %v", au.EID, err)
		return
	}
	user = uu
	user.SetLanguage(f.defaultLanguage)
	return
}

func (f *CFeature) setUserUnsafe(user *users.User) (err error) {
	userUrlPath := f.userPath + "/" + user.EID
	userFilename := f.userPath + "/" + user.EID + "." + user.CPage.Format()
	if user.CPage.PageMatter().Path != userFilename {
		log.DebugF("correcting inconsistency: %v != [correct] %v", user.CPage.PageMatter().Path, userFilename)
		// user.Page.Path = userFilename
		user.CPage.PageMatter().Path = userFilename
	}
	user.CPage.SetSlugUrl(userUrlPath)

	var pm *matter.PageMatter
	if pm, err = page.NewMatterFromPage(&user.CPage); err != nil {
		err = fmt.Errorf("error constructing new front-matter from user: %v - %v", user.EID, err)
		return
	}
	pm.Path = userFilename
	pm.Updated = time.Now()

	err = f.WritePageMatter(pm)
	return
}

func (f *CFeature) makeGroupActionsFilePath(group feature.Group) (path string) {
	path = f.groupPath + "/" + group.String() + ".group"
	return
}

func (f *CFeature) parseGroupActionsFile(group feature.Group) (pm *matter.PageMatter, bodyList feature.Actions, exists bool) {
	groupFilename := f.makeGroupActionsFilePath(group)

	if path, ee := f.FindPageMatterPath(groupFilename); ee == nil {
		exists = true
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		if pm, ee = f.ReadPageMatter(path); ee != nil {
			// this should never happen
			log.WarnF("(will overwrite) error reading group page matter: %v - %v - %v", group, path, ee)
			pm = nil
		}
	}
	if pm == nil {
		pm = matter.NewPageMatter(f.Tag().String(), groupFilename+".text", "", matter.JsonMatter, beContext.Context{})
	}

	bodyList = feature.NewActionsFromStringNL(pm.Body)
	return
}

func (f *CFeature) parseUserGroupsFile(eid string) (pm *matter.PageMatter, bodyList feature.Groups) {
	groupsFilename := f.groupsPath + "/" + eid + ".groups"

	if path, ee := f.FindPageMatterPath(groupsFilename); ee == nil {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		if pm, ee = f.ReadPageMatter(path); ee != nil {
			// this should never happen
			log.ErrorF("(will overwrite) error reading groups page matter: %v - %v - %v", eid, path, ee)
			pm = nil
		}
	}
	if pm == nil {
		pm = matter.NewPageMatter(f.Tag().String(), groupsFilename+".text", "", matter.JsonMatter, beContext.Context{})
	}

	bodyList = feature.NewGroupsFromStringNL(pm.Body)
	return
}

func (f *CFeature) addUserToGroupsUnsafe(eid string, groups ...feature.Group) (err error) {

	pm, bodyList := f.parseUserGroupsFile(eid)

	for _, group := range groups {
		if !bodyList.Has(group) {
			bodyList = append(bodyList, group)
		}
	}

	newGroupLookup := make(map[feature.Group]int)
	for idx, group := range bodyList {
		newGroupLookup[group] = idx + 1
	}

	pm.Matter["Lookup"] = newGroupLookup
	pm.Body = bodyList.AsNewlines()

	if err = f.WritePageMatter(pm); err != nil {
		err = fmt.Errorf("error writing page matter after pruning groups: %v - %v - %v", eid, pm.Path, err)
	}

	return
}

func (f *CFeature) removeUserFromGroupsUnsafe(eid string, groups ...feature.Group) (err error) {

	pm, bodyList := f.parseUserGroupsFile(eid)

	bodyList = bodyList.Remove(groups...)

	newGroupLookup := make(map[feature.Group]int)
	for idx, item := range bodyList {
		newGroupLookup[item] = idx + 1
	}

	pm.Matter["Lookup"] = newGroupLookup
	pm.Body = bodyList.AsNewlines()

	if err = f.WritePageMatter(pm); err != nil {
		err = fmt.Errorf("error writing page matter after pruning groups: %v - %v - %v", eid, pm.Path, err)
	}

	return
}

func (f *CFeature) isUserInGroupUnsafe(eid string, group feature.Group) (present bool) {
	_, bodyList := f.parseUserGroupsFile(eid)
	present = bodyList.Has(group)
	return
}

func (f *CFeature) getUserGroupsUnsafe(eid string) (groups feature.Groups) {
	_, groups = f.parseUserGroupsFile(eid)
	return
}