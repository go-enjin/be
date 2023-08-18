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
	"github.com/go-enjin/be/pkg/feature"
)

type UsersManager interface {
	UserProvider
	UserManager
}

type UserProvider interface {
	// GetUser returns the user by enjin ID
	GetUser(eid string) (user *User, err error)

	// ListUsers returns a paginated list of user EIDs
	ListUsers(start, page, numPerPage int) (list []string)
}

type UserManager interface {
	// NewUser constructs a new User instance and adds it to the userbase
	NewUser(au *AuthUser) (user *User, err error)

	// SetUser writes the given User to the system
	SetUser(user *User) (err error)

	// RemoveUser deletes a user from the system
	RemoveUser(eid string) (err error)
}

type UserActionsProvider interface {
	UserActions() (list feature.Actions)
}