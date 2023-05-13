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
	"github.com/go-enjin/be/pkg/page"
	types "github.com/go-enjin/be/pkg/types/theme-types"
)

type Profile struct {
	User
	page.Page
}

func NewProfile(user *User, path, raw string, created, updated int64, formats types.FormatProvider, enjin beContext.Context) (p *Profile, err error) {
	var pg *page.Page
	if pg, err = page.New(path, raw, created, updated, formats, enjin); err != nil {
		return
	}
	p = &Profile{
		User: *user,
		Page: *pg,
	}
	p.Context.SetSpecific("UserID", p.ID)
	if v := p.Context.String("UserName", ""); v == "" {
		p.Context.SetSpecific("UserName", p.Name)
	}
	if v := p.Context.String("UserEmail", ""); v == "" {
		p.Context.SetSpecific("UserEmail", p.Email)
	}
	if v := p.Context.String("UserImage", ""); v == "" {
		p.Context.SetSpecific("UserImage", p.Image)
	}
	return
}