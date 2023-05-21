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
	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/hash/sha"
)

type AuthUser struct {
	RID     string            `json:"real-id"`
	EID     string            `json:"enjin-id"`
	Name    string            `json:"name"`
	Email   string            `json:"email"`
	Image   string            `json:"image"`
	Origin  string            `json:"origin"`
	Context beContext.Context `json:"context"`
}

func NewAuthUser(id, name, email, image string, ctx beContext.Context) (user *AuthUser) {
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