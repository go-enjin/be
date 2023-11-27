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

package signals

import (
	"github.com/go-enjin/be/pkg/feature/signaling"
)

const (
	PreCreateUser    signaling.Signal = "pre-create-user"
	PostCreateUser   signaling.Signal = "post-create-user"
	PreSignUpUser    signaling.Signal = "pre-sign-up-user"
	PostSignUpUser   signaling.Signal = "post-sign-up-user"
	PreRetrieveUser  signaling.Signal = "pre-retrieve-user"
	PostRetrieveUser signaling.Signal = "post-retrieve-user"
	PreDeleteUser    signaling.Signal = "pre-delete-user"
	PostDeleteUser   signaling.Signal = "post-delete-user"

	PreUpdateUserName     signaling.Signal = "pre-update-user-name"
	PostUpdateUserName    signaling.Signal = "post-update-user-name"
	PreUpdateUserImage    signaling.Signal = "pre-update-user-image"
	PostUpdateUserImage   signaling.Signal = "post-update-user-image"
	PreUpdateUserContext  signaling.Signal = "pre-update-user-context"
	PostUpdateUserContext signaling.Signal = "post-update-user-context"

	PreSetUserName      signaling.Signal = "pre-set-user-name"
	PostSetUserName     signaling.Signal = "post-set-user-name"
	PreSetUserImage     signaling.Signal = "pre-set-user-image"
	PostSetUserImage    signaling.Signal = "post-set-user-image"
	PreSetUserContext   signaling.Signal = "pre-set-user-context"
	PostSetUserContext  signaling.Signal = "post-set-user-context"
	PreSetUserSetting   signaling.Signal = "pre-set-user-setting"
	PostSetUserSetting  signaling.Signal = "post-set-user-setting"
	PreSetUserSettings  signaling.Signal = "pre-set-user-settings"
	PostSetUserSettings signaling.Signal = "post-set-user-settings"

	PreUpdateUserActive       signaling.Signal = "pre-update-user-active"
	PostUpdateUserActive      signaling.Signal = "post-update-user-active"
	PreUpdateUserAdminLocked  signaling.Signal = "pre-update-user-admin-locked"
	PostUpdateUserAdminLocked signaling.Signal = "post-update-user-admin-locked"

	PreUpdateUserGroups       signaling.Signal = "pre-update-user-groups"
	PostUpdateUserGroups      signaling.Signal = "post-update-user-groups"
	PreUpdateUserPermissions  signaling.Signal = "pre-update-user-permissions"
	PostUpdateUserPermissions signaling.Signal = "post-update-user-permissions"

	PreSetUserActive       signaling.Signal = "pre-set-user-active"
	PostSetUserActive      signaling.Signal = "post-set-user-active"
	PreSetUserAdminLocked  signaling.Signal = "pre-set-user-admin-locked"
	PostSetUserAdminLocked signaling.Signal = "post-set-user-admin-locked"

	PreSetUserGroups       signaling.Signal = "pre-set-user-groups"
	PostSetUserGroups      signaling.Signal = "post-set-user-groups"
	PreSetUserPermissions  signaling.Signal = "pre-set-user-permissions"
	PostSetUserPermissions signaling.Signal = "post-set-user-permissions"

	PreSetGroupPermissions  signaling.Signal = "pre-set-group-permissions"
	PostSetGroupPermissions signaling.Signal = "post-set-group-permissions"

	PreCreateGroup    signaling.Signal = "pre-create-group"
	PostCreateGroup   signaling.Signal = "post-create-group"
	PreRetrieveGroup  signaling.Signal = "pre-retrieve-group"
	PostRetrieveGroup signaling.Signal = "post-retrieve-group"
	PreUpdateGroup    signaling.Signal = "pre-update-group"
	PostUpdateGroup   signaling.Signal = "post-update-group"
	PreDeleteGroup    signaling.Signal = "pre-delete-group"
	PostDeleteGroup   signaling.Signal = "post-delete-group"
)
