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

package users

import (
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/hash/sha"
)

var (
	_ feature.AuthUser = (*AuthUser)(nil)
)

type AuthUser struct {
	RID     string          `json:"real-id"`
	EID     string          `json:"enjin-id"`
	Name    string          `json:"name"`
	Email   string          `json:"email"`
	Image   string          `json:"image"`
	Origin  string          `json:"origin"`
	Context context.Context `json:"context"`
}

func NewAuthUser(id, name, email, image string, ctx context.Context) (user *AuthUser) {
	eid, _ := sha.DataHash10([]byte(id))
	user = &AuthUser{
		RID:     id,
		EID:     eid,
		Name:    name,
		Email:   email,
		Image:   image,
		Context: ctx,
	}
	return
}

func (a *AuthUser) GetRID() (rid string) {
	rid = a.RID
	return
}

func (a *AuthUser) GetEID() (eid string) {
	eid = a.EID
	return
}

func (a *AuthUser) GetName() (name string) {
	name = a.Name
	return
}

func (a *AuthUser) GetEmail() (email string) {
	email = a.Email
	return
}

func (a *AuthUser) GetImage() (image string) {
	image = a.Image
	return
}

func (a *AuthUser) GetOrigin() (origin string) {
	origin = a.Origin
	return
}

func (a *AuthUser) GetContext() (context context.Context) {
	context = a.Context
	return
}