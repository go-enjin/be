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

package user_manager

import (
	"errors"
	"net/http"

	"github.com/mrz1836/go-sanitize"

	beContext "github.com/go-enjin/be/pkg/context"
	bePkgEditor "github.com/go-enjin/be/pkg/editor"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) opCreateUser(form beContext.Context, r *http.Request) {
	eid := userbase.GetCurrentEID(r)
	printer := lang.GetPrinterFromRequest(r)

	if !userbase.CurrentUserCan(r, f.Action("create", "user")) {
		log.WarnRF(r, "user %q attempted to create a new user without permission!", eid)
		f.Site().PushErrorNotice(eid, true, berrs.PermissionDeniedError(printer))
		return
	}

	var userEmail string
	if userEmail = form.String(bePkgEditor.CreateUserActionKey+"~email", ""); userEmail == "" {
		f.Site().PushErrorNotice(eid, true, printer.Sprintf(`An email address is required to create a new user.`))
		return
	}
	userEmail = sanitize.Email(userEmail, false)

	userName := form.String(bePkgEditor.CreateUserActionKey+"~name", "")

	saf := f.Site().SiteAuth()
	su := f.Site().SiteUsers()

	userRID := su.MakeRealID(userEmail)
	userEID := su.MakeEnjinID(userRID)

	if err := su.CreateUser(r, saf.Tag().Kebab(), userRID, userEID, userEmail); err != nil {
		log.ErrorRF(r, "error creating new user: %v - %v - %v", userEID, userEmail, err)
		if errors.Is(err, berrs.ErrExistsAlready) {
			f.Site().PushErrorNotice(eid, true, printer.Sprintf("A user with the email %[1]s exists already", userEmail))
		} else {
			f.Site().PushErrorNotice(eid, true, berrs.UnexpectedError(printer))
		}
		return
	}
	f.Site().PushInfoNotice(eid, true, printer.Sprintf(`New user created with email: %[1]s`, userEmail))

	if userName != "" {
		if err := su.SetUserName(r, userEID, userName); err != nil {
			log.ErrorRF(r, "error settings newly created user's name: %q - %q - %v", userEID, userName, err)
		}
	}
	return
}