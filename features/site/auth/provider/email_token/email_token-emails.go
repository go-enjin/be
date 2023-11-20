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

package email_token

import (
	"net/http"

	"github.com/Shopify/gomail"

	beContext "github.com/go-enjin/be/pkg/context"
)

func (f *CFeature) sendUserEmail(r *http.Request, to, subject, template string, body beContext.Context) (err error) {
	var msg *gomail.Message
	if msg, err = f.emailProvider.NewEmail(template, body); err != nil {
		return
	}
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	err = f.emailSender.SendEmail(r, f.emailAccount, msg)
	return
}