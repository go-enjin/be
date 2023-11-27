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

	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/request"
)

func (f *CFeature) HandleUserManager(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		f.Enjin.ServeRedirect(f.SiteFeaturePath(), w, r)
		return
	}

	form := request.SafeParseForm(r)

	switch request.SafeQueryFormValue(r, "submit") {
	case editor.CreateUserActionKey:
		f.opCreateUser(form, r)

	case editor.DeleteUserActionKey:
		f.opDeleteUser(form, r)

	case editor.ReactivateUserActionKey:
		f.opReactivateUser(form, r)

	case editor.DeactivateUserActionKey:
		f.opDeactivateUser(form, r)

	case editor.AdminLockUserActionKey:
		f.opAdminLockUser(form, r)

	case editor.AdminUnlockUserActionKey:
		f.opAdminUnlockUser(form, r)

	case editor.ResetUserOtpActionKey:
		f.opResetUserOtp(form, r)
	}

	f.Enjin.ServeRedirect(f.SiteFeaturePath(), w, r)
	return
}
