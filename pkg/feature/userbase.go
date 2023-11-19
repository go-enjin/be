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

package feature

import (
	"net/http"
)

type UsersManager interface {
	Feature

	UserProvider
	UserManager
}

type UserProvider interface {
	Feature

	// GetUser returns the user by enjin ID
	GetUser(eid string) (user User, err error)

	// ListUsers returns a paginated list of user EIDs
	ListUsers(start, page, numPerPage int) (list []string)
}

type UserManager interface {
	Feature

	// NewUser constructs a new User instance and adds it to the userbase
	NewUser(au AuthUser) (user User, err error)

	// SetUser writes the given User to the system
	SetUser(user User) (err error)

	// RemoveUser deletes a user from the system
	RemoveUser(eid string) (err error)
}

type UserActionsProvider interface {
	Feature

	UserActions() (list Actions)
	Action(verb string, details ...string) (action Action)
}

type AuthProvider interface {
	Feature

	AuthenticateRequest(w http.ResponseWriter, r *http.Request) (handled bool, modified *http.Request)
}

type AuthUserApi interface {
	Feature

	RequireApiUser(next http.Handler) http.Handler
	RequireUserCan(action Action) func(next http.Handler) http.Handler
}

type GroupsProvider interface {
	Feature

	// IsUserInGroup returns true if the user is in the given group
	IsUserInGroup(eid string, group Group) (present bool)

	// GetUserGroups returns the user's list of groups
	GetUserGroups(eid string) (groups Groups)
}

type GroupsManager interface {
	Feature

	// GetGroupActions returns the list of actions associated with group
	GetGroupActions(group Group) (actions Actions)

	// UpdateGroup appends the given actions to the group, creating the group
	// if non exists
	UpdateGroup(group Group, actions ...Action) (err error)

	// RemoveGroup deletes the given group from the system. Does not delete any
	// fallback groups
	RemoveGroup(group Group) (err error)

	// AddUserToGroup adds a user to the list of groups given
	AddUserToGroup(eid string, groups ...Group) (err error)

	// RemoveUserFromGroup removes a user from each of the given groups
	RemoveUserFromGroup(eid string, groups ...Group) (err error)
}

type AuthUserProvider interface {
	Feature

	// AuthUserPresent returns true if a user with the EID given is present
	AuthUserPresent(eid string) (present bool)

	// GetAuthUser returns the user by user.EID
	GetAuthUser(eid string) (user AuthUser, err error)
}

type AuthUserManager interface {
	Feature

	// NewAuthUser constructs a new AuthUser instance and saves it to the
	// userbase
	NewAuthUser(rid, name, email, picture, audience string, attributes map[string]interface{}) (user AuthUser, err error)

	// SetAuthUser writes the given User to the system
	SetAuthUser(user AuthUser) (err error)

	// RemoveAuthUser deletes a user from the system
	RemoveAuthUser(eid string) (err error)
}

type SecretsProvider interface {
	Feature

	// GetUserSecret returns the user's password hash, returns "" if the user
	// secret is not found
	GetUserSecret(id string) (hash string)
}

type Manager interface {
	AuthUserProvider
	AuthUserManager
	UserProvider
	UserManager
	GroupsProvider
	GroupsManager
}