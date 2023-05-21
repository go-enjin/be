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

type AuthUserProvider interface {
	// AuthUserPresent returns true if a user with the EID given is present
	AuthUserPresent(eid string) (present bool)

	// GetAuthUser returns the user by user.EID
	GetAuthUser(eid string) (user *AuthUser, err error)
}

type AuthUserManager interface {
	// NewAuthUser constructs a new AuthUser instance and saves it to the
	// userbase
	NewAuthUser(eid, name, email, picture, audience string, attributes map[string]interface{}) (user *AuthUser, err error)

	// SetAuthUser writes the given User to the system
	SetAuthUser(user *AuthUser) (err error)

	// RemoveAuthUser deletes a user from the system
	RemoveAuthUser(eid string) (err error)
}