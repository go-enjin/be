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

package site_users

import (
	"fmt"

	"github.com/mrz1836/go-sanitize"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/filesystem"
	"github.com/go-enjin/be/pkg/feature/signaling"
	site_environ "github.com/go-enjin/be/pkg/feature/site-environ"
	uses_actions "github.com/go-enjin/be/pkg/feature/uses-actions"
	uses_enjin_salt "github.com/go-enjin/be/pkg/feature/uses-enjin-salt"
	uses_kvc "github.com/go-enjin/be/pkg/feature/uses-kvc"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/maps"
	bePath "github.com/go-enjin/be/pkg/path"
)

var (
	DefaultUserPath  = "/user"
	DefaultGroupPath = "/group"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "fs-site-users"

type Feature interface {
	feature.Feature
	feature.SiteUsersProvider
	signaling.Signaling
}

type MakeFeature interface {
	filesystem.MakeFeature[MakeFeature]
	uses_kvc.MakeFeature[MakeFeature]

	// SetUserPath specifies the underlying filesystem mount point to use for locating user files
	SetUserPath(path string) MakeFeature

	// SetGroupPath specifies the underlying filesystem mount point to use for locating group files
	SetGroupPath(path string) MakeFeature

	// InitGroup will create the specific group on startup, does nothing if the group already exists
	InitGroup(group feature.Group, actions ...feature.Action) MakeFeature

	// SetEnjinSalt specifies the default random string to use for making new Enjin identifiers (EID) and is overridden
	// by the corresponding command-line flag value (--fs-site-users-enjin-salt)
	SetEnjinSalt(value string) MakeFeature

	Make() Feature
}

type CFeature struct {
	filesystem.CFeature[MakeFeature]
	signaling.CSignaling
	uses_kvc.CUsesKVC[MakeFeature]
	uses_actions.CUsesActions

	env *site_environ.CSiteEnviron[MakeFeature]

	secretKey []byte
	enjinSalt *uses_enjin_salt.CUsesEnjinSalt

	initGroups map[feature.Group]feature.Actions
	initUsers  map[string]feature.Groups

	userPath  string
	groupPath string

	PermissionViewOwn     feature.Action
	PermissionViewOther   feature.Action
	PermissionUpdateOwn   feature.Action
	PermissionUpdateOther feature.Action
	PermissionDeleteOwn   feature.Action
	PermissionDeleteOther feature.Action

	PermissionSignUpUser  feature.Action
	PermissionCreateUser  feature.Action
	PermissionAdminPerms  feature.Action
	PermissionAdminGroups feature.Action

	userLocker feature.SyncRWLocker
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.Construct(f)
	return f
}

func (f *CFeature) UsageNotes() (notes []string) {
	notes = f.env.SiteEnvironUsageNotes()
	return
}

func (f *CFeature) Construct(this interface{}) {
	f.CFeature.Construct(this)
	f.CUsesActions.ConstructUsesActions(f)
	f.enjinSalt = uses_enjin_salt.New(this)
	f.PermissionViewOwn = f.Action("view-own", "user")
	f.PermissionViewOther = f.Action("view-other", "user")
	f.PermissionUpdateOwn = f.Action("update-own", "user")
	f.PermissionUpdateOther = f.Action("update-other", "user")
	f.PermissionDeleteOwn = f.Action("delete-own", "user")
	f.PermissionDeleteOther = f.Action("delete-other", "user")
	f.PermissionSignUpUser = f.Action("sign-up", "user")
	f.PermissionCreateUser = f.Action("create", "user")
	f.PermissionAdminPerms = f.Action("admin-perms", "user")
	f.PermissionAdminGroups = f.Action("admin-groups", "user")
	return
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.CFeature.Localized = false // themes are not localized filesystems!
	f.CSignaling.InitSignaling()
	f.CUsesKVC.InitUsesKVC(this)
	f.initGroups = make(map[feature.Group]feature.Actions)
	f.initUsers = make(map[string]feature.Groups)
	f.env = site_environ.New[MakeFeature](this,
		"init-group", "space separated permissions",
		"init-user-email", "email addresses of users to have groups assigned",
		"init-user-group", "space separated group names",
	)
	f.userPath = DefaultUserPath
	f.groupPath = DefaultGroupPath
	return
}

func (f *CFeature) InitGroup(group feature.Group, actions ...feature.Action) MakeFeature {
	f.initGroups[group] = f.initGroups[group].Append(actions...)
	return f
}

func (f *CFeature) InitUser(email string, group ...feature.Group) MakeFeature {
	if email = sanitize.Email(email, false); email != "" {
		f.initUsers[email] = f.initUsers[email].Append(group...)
	}
	return f
}

func (f *CFeature) SetUserPath(path string) MakeFeature {
	f.userPath = bePath.CleanWithSlash(path)
	return f
}

func (f *CFeature) SetGroupPath(path string) MakeFeature {
	f.groupPath = bePath.CleanWithSlash(path)
	return f
}

func (f *CFeature) SetEnjinSalt(value string) MakeFeature {
	f.enjinSalt.SetEnjinSalt(value)
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	} else if err = f.CUsesKVC.BuildUsesKVC(); err != nil {
		return
	} else if err = f.enjinSalt.BuildEnjinSalt(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	} else if err = f.CUsesKVC.StartupUsesKVC(f.Enjin.Features()); err != nil {
		return
	} else if err = f.enjinSalt.StartupEnjinSalt(ctx); err != nil {
		return
	} else if err = f.env.StartupSiteEnviron(); err != nil {
		return
	} else if err = f.loadEnvironment(); err != nil {
		return
	}

	readStore := f.KVC().MustBucket(ReadLocksBucket)
	writeStore := f.KVC().MustBucket(WriteLocksBucket)
	f.userLocker = f.Enjin.NewSyncRWLocker(f.Tag(), "user-lock", readStore, writeStore)

	if !f.MountPoints.HasRootOrAllOf(f.userPath, f.groupPath) {
		err = fmt.Errorf("at least one root (/) mount point is required")
		return
	} else if found := f.MountPoints.FindPathPoints(f.userPath); found.Len() == 0 {
		err = fmt.Errorf("user mount point not found: %v", f.userPath)
		return
	} else if found = f.MountPoints.FindPathPoints(f.groupPath); found.Len() == 0 {
		err = fmt.Errorf("group mount point not found: %v", f.groupPath)
		return
	}

	for _, group := range maps.SortedKeys(f.initGroups) {
		var actions feature.Actions
		if f.GroupPresent(group) {
			if actions, err = f.getGroup(group); err != nil {
				err = fmt.Errorf("error getting group permissions %q: %v", group, err)
				return
			}
		}
		actions = actions.Append(f.initGroups[group]...)
		if err = f.setGroup(group, actions...); err != nil {
			err = fmt.Errorf("error initializing group %q: %v", group, err)
			return
		}
	}

	for _, email := range maps.SortedKeys(f.initUsers) {
		rid := f.MakeRealID(email)
		eid := f.MakeEnjinID(rid)
		if au, ee := f.getUser(eid); ee == nil {
			au.Groups = au.Groups.Append(f.initUsers[email]...)
			_ = f.setUser(au)
		}
	}

	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) UserActions() (list feature.Actions) {

	list = feature.Actions{
		f.PermissionViewOwn,
		f.PermissionViewOther,
		f.PermissionUpdateOwn,
		f.PermissionUpdateOther,
		f.PermissionDeleteOwn,
		f.PermissionDeleteOther,
		f.PermissionAdminPerms,
		f.PermissionAdminGroups,
	}

	return
}

func (f *CFeature) MakeRealID(email string) (rid string) {
	rid = f.KebabTag + "_" + email
	return
}

func (f *CFeature) MakeEnjinID(rid string) (eid string) {
	eid, _ = sha.DataHash10(f.enjinSalt.GetEnjinSalt() + rid)
	return
}
