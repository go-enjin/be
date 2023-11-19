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
	"net/http"

	"github.com/mrz1836/go-sanitize"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) opDeleteUser(form context.Context, r *http.Request) {
	eid := userbase.GetCurrentEID(r)
	printer := lang.GetPrinterFromRequest(r)

	if !userbase.CurrentUserCan(r, f.Action("delete", "user")) {
		log.WarnRF(r, "user %q attempted to create a new user without permission!", eid)
		f.Site().PushErrorNotice(eid, true, errors.PermissionDeniedError(printer))
		return
	}

	var userEID, userEmail, confirmed string
	if userEID = form.String("target", ""); userEID == "" {
		return
	} else if userEmail = form.String(editor.DeleteUserActionKey+"~email", ""); userEmail == "" {
		f.Site().PushErrorNotice(eid, true, printer.Sprintf("The user's email address is required to confirm the deletion of that user."))
		return
	} else if confirmed = form.String(editor.DeleteUserActionKey+"-confirmed", "false"); confirmed != "true" {
		return
	}

	userEmail = sanitize.Email(userEmail, false)

	su := f.Site().SiteUsers()
	su.LockUser(r, userEID)
	defer su.UnlockUser(r, userEID)

	if !su.UserPresent(userEID) {
		log.WarnRF(r, "user %q attempting to delete a user that does not exist!", eid)
		return
	}

	if err := su.DeleteUser(r, userEID); err != nil {
		f.Site().PushErrorNotice(eid, true, errors.UnexpectedError(printer))
		return
	}

	f.Site().PushInfoNotice(eid, true, printer.Sprintf(`Deleted user: %[1]s`, userEmail))
	return
}