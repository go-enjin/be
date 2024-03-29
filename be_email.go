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

package be

import (
	"fmt"
	"net/http"

	"github.com/Shopify/gomail"

	"github.com/go-enjin/be/pkg/feature"
)

func (e *Enjin) FindEmailAccount(account string) (emailSender feature.EmailSender) {
	for _, es := range e.eb.fEmailSenders {
		if es.HasEmailAccount(account) {
			emailSender = es
			return
		}
	}
	return
}

func (e *Enjin) SendEmail(r *http.Request, account string, message *gomail.Message) (err error) {
	if es := e.FindEmailAccount(account); es != nil {
		err = es.SendEmail(r, account, message)
		return
	}
	err = fmt.Errorf("account not found")
	return
}
