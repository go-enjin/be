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
	"encoding/json"
	"net/http"

	beErrors "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/signals"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) makeGroupPath(group feature.Group) (path string) {
	path = f.groupPath + "/" + group.String() + ".json"
	return
}

func (f *CFeature) GroupPresent(group feature.Group) (present bool) {
	present = f.MountPoints.Exists(f.makeGroupPath(group))
	return
}

func (f *CFeature) CreateGroup(r *http.Request, group feature.Group, permissions ...feature.Action) (err error) {

	if stop := f.Emit(signals.PreCreateGroup, f.Tag().String(), r, group, &permissions); stop {
		err = beErrors.ErrSignalStopped
		return
	} else if !userbase.CurrentUserCan(r, f.PermissionAdminGroups) {
		err = beErrors.ErrPermissionDenied
		return
	} else if f.GroupPresent(group) {
		err = beErrors.ErrExistsAlready
		return
	}

	actions := feature.Actions(permissions)
	err = f.MountPoints.WriteFile(f.makeGroupPath(group), actions.Bytes())

	f.Emit(signals.PostCreateGroup, f.Tag().String(), r, group, actions, err)
	return
}

func (f *CFeature) RetrieveGroup(r *http.Request, group feature.Group) (permissions feature.Actions, err error) {

	if stop := f.Emit(signals.PreRetrieveGroup, f.Tag().String(), r, group, &permissions); stop {
		err = beErrors.ErrSignalStopped
		return
	} else if !userbase.CurrentUserCan(r, f.PermissionAdminGroups) {
		err = beErrors.ErrPermissionDenied
		return
	} else if !f.GroupPresent(group) {
		err = beErrors.ErrGroupNotFound
		return
	}

	var data []byte
	if data, err = f.MountPoints.ReadFile(f.makeGroupPath(group)); err != nil {
		return
	}
	permissions = feature.Actions{}
	err = json.Unmarshal(data, &permissions)

	f.Emit(signals.PostRetrieveGroup, f.Tag().String(), r, group, permissions, err)
	return
}

func (f *CFeature) UpdateGroup(r *http.Request, group feature.Group, permissions ...feature.Action) (err error) {
	actions := feature.Actions(permissions)

	if stop := f.Emit(signals.PreUpdateGroup, f.Tag().String(), r, group, &actions); stop {
		err = beErrors.ErrSignalStopped
		return
	} else if !userbase.CurrentUserCan(r, f.PermissionAdminGroups) {
		err = beErrors.ErrPermissionDenied
		return
	} else if !f.GroupPresent(group) {
		err = beErrors.ErrGroupNotFound
		return
	}

	if !userbase.CurrentUserCan(r, f.PermissionAdminGroups) {
		err = beErrors.ErrPermissionDenied
		return
	} else if !f.GroupPresent(group) {
		err = beErrors.ErrGroupNotFound
		return
	}
	err = f.MountPoints.WriteFile(f.makeGroupPath(group), actions.Bytes())

	f.Emit(signals.PostUpdateGroup, f.Tag().String(), r, group, actions, err)
	return
}

func (f *CFeature) DeleteGroup(r *http.Request, group feature.Group) (err error) {

	if stop := f.Emit(signals.PreDeleteGroup, f.Tag().String(), r, group); stop {
		err = beErrors.ErrSignalStopped
		return
	} else if !userbase.CurrentUserCan(r, f.PermissionAdminGroups) {
		err = beErrors.ErrPermissionDenied
		return
	} else if !f.GroupPresent(group) {
		err = beErrors.ErrGroupNotFound
		return
	}

	err = f.MountPoints.RemoveFile(f.makeGroupPath(group))

	f.Emit(signals.PostDeleteGroup, f.Tag().String(), r, group, err)
	return
}
