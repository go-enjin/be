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

package feature

import (
	"net/http"

	"github.com/Shopify/gomail"

	beContext "github.com/go-enjin/be/pkg/context"
)

type EmailSender interface {
	Feature

	HasEmailAccount(account string) (present bool)
	SendEmail(r *http.Request, account string, message *gomail.Message) (err error)
}

type EmailProvider interface {
	Feature

	NewEmail(path string, bodyCtx beContext.Context) (message *gomail.Message, err error)
	MakeEmailBody(path string, ctx beContext.Context) (matter beContext.Context, body string, err error)

	ListTemplates() (names []string)
	HasTemplate(name string) (present bool)
}
