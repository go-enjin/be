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
	"net/http"
	"net/url"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/signals"
	"github.com/go-enjin/be/pkg/userbase"
	beUser "github.com/go-enjin/be/types/users"
)

func (f *CFeature) UpdateUserName(r *http.Request, eid string, name string) (err error) {
	uid := userbase.GetCurrentEID(r)

	if stop := f.Emit(signals.PreUpdateUserName, f.Tag().String(), r, eid, &name); stop {
		err = errors.ErrSignalStopped
		return
	} else if !f.checkUserCan(r, uid, eid, f.PermissionUpdateOwn, f.PermissionUpdateOther) {
		err = errors.ErrPermissionDenied
		return
	}

	err = f.SetUserName(r, eid, name)

	f.Emit(signals.PostUpdateUserName, f.Tag().String(), r, eid, name, err)
	return
}

func (f *CFeature) SetUserName(r *http.Request, eid string, name string) (err error) {
	var au *beUser.User

	if stop := f.Emit(signals.PreSetUserName, f.Tag().String(), r, eid, &name); stop {
		err = errors.ErrSignalStopped
		return
	} else if !f.UserPresent(eid) {
		err = errors.ErrUserNotFound
		return
	} else if au, err = f.getUser(eid); err != nil {
		return
	}

	au.Name = forms.StrictSanitize(name)
	err = f.setUser(au)

	f.Emit(signals.PostSetUserName, f.Tag().String(), r, eid, name, err)
	return
}

func (f *CFeature) UpdateUserImage(r *http.Request, eid string, image string) (err error) {
	uid := userbase.GetCurrentEID(r)

	if stop := f.Emit(signals.PreUpdateUserImage, f.Tag().String(), r, eid, &image); stop {
		err = errors.ErrSignalStopped
		return
	} else if !f.checkUserCan(r, uid, eid, f.PermissionUpdateOwn, f.PermissionUpdateOther) {
		err = errors.ErrPermissionDenied
		return
	}

	err = f.SetUserImage(r, eid, image)

	f.Emit(signals.PostUpdateUserImage, f.Tag().String(), r, eid, image, err)
	return
}

func (f *CFeature) SetUserImage(r *http.Request, eid string, image string) (err error) {
	var au *beUser.User

	if stop := f.Emit(signals.PreSetUserImage, f.Tag().String(), r, eid, &image); stop {
		err = errors.ErrSignalStopped
		return
	} else if _, err = url.Parse(image); err != nil {
		err = fmt.Errorf("error parsing image URL: %v", err)
		return
	} else if !f.UserPresent(eid) {
		err = errors.ErrUserNotFound
		return
	} else if au, err = f.getUser(eid); err != nil {
		return
	}

	au.Image = image
	err = f.setUser(au)

	f.Emit(signals.PostSetUserImage, f.Tag().String(), r, eid, image)
	return
}

func (f *CFeature) UpdateUserContext(r *http.Request, eid string, ctx beContext.Context) (err error) {
	uid := userbase.GetCurrentEID(r)

	if stop := f.Emit(signals.PreUpdateUserContext, f.Tag().String(), r, eid, ctx); stop {
		err = errors.ErrSignalStopped
		return
	} else if !f.checkUserCan(r, uid, eid, f.PermissionUpdateOwn, f.PermissionUpdateOther) {
		err = errors.ErrPermissionDenied
		return
	}

	err = f.SetUserContext(r, eid, ctx)

	f.Emit(signals.PostUpdateUserContext, f.Tag().String(), r, eid, ctx, err)
	return
}

func (f *CFeature) SetUserContext(r *http.Request, eid string, ctx beContext.Context) (err error) {
	var au *beUser.User

	if stop := f.Emit(signals.PreSetUserContext, f.Tag().String(), r, eid, ctx); stop {
		err = errors.ErrSignalStopped
		return
	} else if au, err = f.getUser(eid); err != nil {
		return
	}

	ctx.KebabKeys()
	au.Context.ApplySpecific(ctx)
	err = f.setUser(au)

	f.Emit(signals.PostSetUserContext, f.Tag().String(), r, eid, ctx)
	return
}

func (f *CFeature) SetUserSetting(r *http.Request, eid string, key string, value interface{}) (err error) {
	var au *beUser.User

	if stop := f.Emit(signals.PreSetUserSetting, f.Tag().String(), r, eid, key, value); stop {
		err = errors.ErrSignalStopped
		return
	} else if au, err = f.getUser(eid); err != nil {
		return
	}

	settings := au.Context.Context("settings")
	settings.SetSpecific(key, value)
	au.Context.SetSpecific("settings", settings)

	err = f.setUser(au)

	f.Emit(signals.PostSetUserSetting, f.Tag().String(), r, eid, key, value)
	return
}

func (f *CFeature) SetUserSettings(r *http.Request, eid string, ctx beContext.Context) (err error) {
	var au *beUser.User

	if stop := f.Emit(signals.PreSetUserSettings, f.Tag().String(), r, eid, ctx); stop {
		err = errors.ErrSignalStopped
		return
	} else if au, err = f.getUser(eid); err != nil {
		return
	}

	settings := au.Context.Context("settings")
	settings.ApplySpecific(ctx)
	au.Context.SetSpecific("settings", settings)

	err = f.setUser(au)

	f.Emit(signals.PostSetUserSettings, f.Tag().String(), r, eid, ctx)
	return
}

func (f *CFeature) UpdateUserGroups(r *http.Request, eid string, groups ...feature.Group) (err error) {
	uid := userbase.GetCurrentEID(r)
	list := feature.Groups(groups)

	if stop := f.Emit(signals.PreUpdateUserGroups, f.Tag().String(), r, eid, &list); stop {
		err = errors.ErrSignalStopped
		return
	} else if !f.checkUserCan(r, uid, eid, f.PermissionUpdateOwn, f.PermissionUpdateOther, f.PermissionAdminGroups) {
		err = errors.ErrPermissionDenied
		return
	}

	err = f.SetUserGroups(r, eid, list...)

	f.Emit(signals.PostUpdateUserGroups, f.Tag().String(), r, eid, list, err)
	return
}

func (f *CFeature) SetUserGroups(r *http.Request, eid string, groups ...feature.Group) (err error) {
	list := feature.Groups(groups)
	var au *beUser.User

	if stop := f.Emit(signals.PreSetUserGroups, f.Tag().String(), r, eid, &list); stop {
		err = errors.ErrSignalStopped
		return
	} else if !f.UserPresent(eid) {
		err = errors.ErrUserNotFound
		return
	} else if au, err = f.getUser(eid); err != nil {
		return
	}

	au.Groups = feature.Groups{
		userbase.PublicGroup,
		userbase.UsersGroup,
	}.Append(list...)
	err = f.setUser(au)

	f.Emit(signals.PostSetUserGroups, f.Tag().String(), r, eid, list, err)
	return
}

func (f *CFeature) UpdateUserPermissions(r *http.Request, eid string, permissions ...feature.Action) (err error) {
	uid := userbase.GetCurrentEID(r)
	actions := feature.Actions(permissions)

	if stop := f.Emit(signals.PreUpdateUserPermissions, f.Tag().String(), r, eid, &actions); stop {
		err = errors.ErrSignalStopped
		return
	} else if !f.checkUserCan(r, uid, eid, f.PermissionUpdateOwn, f.PermissionUpdateOther, f.PermissionAdminPerms) {
		err = errors.ErrPermissionDenied
		return
	}

	err = f.SetUserPermissions(r, eid, actions...)

	f.Emit(signals.PostUpdateUserPermissions, f.Tag().String(), r, eid, actions, err)
	return
}

func (f *CFeature) SetUserPermissions(r *http.Request, eid string, permissions ...feature.Action) (err error) {
	actions := feature.Actions(permissions)
	var au *beUser.User

	if stop := f.Emit(signals.PreSetUserPermissions, f.Tag().String(), r, eid, &actions); stop {
		err = errors.ErrSignalStopped
		return
	} else if !f.UserPresent(eid) {
		err = errors.ErrUserNotFound
		return
	} else if au, err = f.getUser(eid); err != nil {
		return
	}

	au.Actions = au.Actions.Append(actions...)
	err = f.setUser(au)

	f.Emit(signals.PostSetUserPermissions, f.Tag().String(), r, eid, actions, err)
	return
}

func (f *CFeature) UpdateUserActive(r *http.Request, eid string, active bool) (err error) {
	uid := userbase.GetCurrentEID(r)

	if stop := f.Emit(signals.PreUpdateUserActive, f.Tag().String(), r, eid, &active); stop {
		err = errors.ErrSignalStopped
		return
	} else if !f.checkUserCan(r, uid, eid, f.PermissionUpdateOwn, f.PermissionUpdateOther, f.PermissionAdminPerms) {
		err = errors.ErrPermissionDenied
		return
	}

	err = f.SetUserActive(r, eid, active)

	f.Emit(signals.PostUpdateUserActive, f.Tag().String(), r, eid, active, err)
	return
}

func (f *CFeature) SetUserActive(r *http.Request, eid string, active bool) (err error) {
	var au *beUser.User

	if stop := f.Emit(signals.PreSetUserActive, f.Tag().String(), r, eid, &active); stop {
		err = errors.ErrSignalStopped
		return
	} else if !f.UserPresent(eid) {
		err = errors.ErrUserNotFound
		return
	} else if au, err = f.getUser(eid); err != nil {
		return
	}

	au.Active = active
	err = f.setUser(au)

	f.Emit(signals.PostSetUserActive, f.Tag().String(), r, eid, active, err)
	return
}

func (f *CFeature) GetUserActive(r *http.Request, eid string) (active bool, err error) {
	var au *beUser.User

	if !f.UserPresent(eid) {
		err = errors.ErrUserNotFound
		return
	} else if au, err = f.getUser(eid); err != nil {
		return
	}

	active = au.Active
	return
}

func (f *CFeature) UpdateUserAdminLocked(r *http.Request, eid string, locked bool) (err error) {
	uid := userbase.GetCurrentEID(r)

	if stop := f.Emit(signals.PreUpdateUserAdminLocked, f.Tag().String(), r, eid, &locked); stop {
		err = errors.ErrSignalStopped
		return
	} else if !f.checkUserCan(r, uid, eid, f.PermissionUpdateOwn, f.PermissionUpdateOther, f.PermissionAdminPerms) {
		err = errors.ErrPermissionDenied
		return
	}

	err = f.SetUserAdminLocked(r, eid, locked)

	f.Emit(signals.PostUpdateUserAdminLocked, f.Tag().String(), r, eid, locked, err)
	return
}

func (f *CFeature) SetUserAdminLocked(r *http.Request, eid string, locked bool) (err error) {
	var au *beUser.User

	if stop := f.Emit(signals.PreSetUserAdminLocked, f.Tag().String(), r, eid, &locked); stop {
		err = errors.ErrSignalStopped
		return
	} else if !f.UserPresent(eid) {
		err = errors.ErrUserNotFound
		return
	} else if au, err = f.getUser(eid); err != nil {
		return
	}

	au.AdminLocked = locked
	err = f.setUser(au)

	f.Emit(signals.PostSetUserAdminLocked, f.Tag().String(), r, eid, locked, err)
	return
}

func (f *CFeature) GetUserAdminLocked(r *http.Request, eid string) (locked bool, err error) {
	var au *beUser.User

	if !f.UserPresent(eid) {
		err = errors.ErrUserNotFound
		return
	} else if au, err = f.getUser(eid); err != nil {
		return
	}

	locked = au.AdminLocked
	return
}

func (f *CFeature) GetUserStatus(r *http.Request, eid string) (active, locked, visitor bool, err error) {
	var au *beUser.User

	if !f.UserPresent(eid) {
		err = errors.ErrUserNotFound
		return
	} else if au, err = f.getUser(eid); err != nil {
		return
	}

	active = au.Active
	locked = au.AdminLocked
	visitor = au.IsVisitor()
	return
}
