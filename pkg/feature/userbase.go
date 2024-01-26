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

type UserActionsProvider interface {
	Feature

	UserActions() (list Actions)
	Action(verb string, details ...string) (action Action)
}

type AuthProvider interface {
	Feature

	AuthenticateRequest(w http.ResponseWriter, r *http.Request) (handled bool, modified *http.Request)
}

type GroupsProvider interface {
	Feature

	// IsUserInGroup returns true if the user is in the given group
	IsUserInGroup(eid string, group Group) (present bool)

	// GetUserGroups returns the user's list of groups
	GetUserGroups(eid string) (groups Groups)
}

type UserProvider interface {
	Feature

	// UserPresent returns true if a user with the EID given is present
	UserPresent(eid string) (present bool)

	// GetUser returns the user by user.EID
	GetUser(eid string) (user User, err error)
}

type UserManager interface {
	Feature

	// NewUser constructs a new User instance and saves it to the userbase
	NewUser(rid, name, email, picture, audience string, attributes map[string]interface{}) (user User, err error)

	// SetUser writes the given User to the system
	SetUser(user User) (err error)

	// RemoveUser deletes a user from the system
	RemoveUser(eid string) (err error)
}

type SecretsProvider interface {
	Feature

	// GetUserSecret returns the user's password hash, returns "" if the user
	// secret is not found
	GetUserSecret(id string) (hash string)
}
