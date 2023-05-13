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

import "net/http"

type AuthProvider interface {
	AuthenticateRequest(w http.ResponseWriter, r *http.Request) (handled bool, modified *http.Request)
}

type UsersProvider interface {
	AddUser(user *User) (err error)
	GetUser(id string) (user *User, err error)
}

type GroupsProvider interface {
	AddUserToGroup(id string, groups ...string) (err error)
	IsUserInGroup(id string, group string) (present bool)
	GetUserGroups(id string) (groups []string)
}

type SecretsProvider interface {
	GetUserSecret(id string) (hash string)
}

type UserProfilesProvider interface {
	AddUserProfile(u *User) (p *Profile, err error)
	GetUserProfile(u *User) (p *Profile, err error)
}