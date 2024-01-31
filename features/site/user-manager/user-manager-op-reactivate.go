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

	"github.com/go-corelibs/x-text/message"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) opReactivateUser(form context.Context, r *http.Request) {
	eid := userbase.GetCurrentEID(r)
	printer := message.GetPrinter(r)

	if !userbase.CurrentUserCan(r, f.Action("update", "user")) {
		log.WarnRF(r, "user %q attempted to reactivate a user without permission!", eid)
		f.Site().PushErrorNotice(eid, true, errors.PermissionDeniedError(printer))
		return
	}

	var userEID, confirmed string
	if userEID = form.String("target", ""); userEID == "" {
		return
	} else if confirmed = form.String(editor.ReactivateUserActionKey+"-confirmed", "false"); confirmed != "true" {
		return
	}

	su := f.Site().SiteUsers()
	su.LockUser(r, userEID)
	defer su.UnlockUser(r, userEID)

	if !su.UserPresent(userEID) {
		log.WarnRF(r, "user %q attempting to reactivate a user that does not exist!", eid)
		return
	}

	if err := su.SetUserActive(r, userEID, true); err != nil {
		f.Site().PushErrorNotice(eid, true, errors.UnexpectedError(printer))
		return
	}

	f.Site().PushInfoNotice(eid, true, printer.Sprintf(`The user's primary sign-in has been reactivated.`))
	return
}
