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

package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Shopify/gomail"

	"github.com/go-enjin/github-com-go-pkgz-auth/provider"
	"github.com/go-enjin/github-com-go-pkgz-auth/token"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) authAddLocalLoginProvider() {
	// TODO: implement standard local login
}

func (f *CFeature) authAddVerifyEmailProviderFunc(providerName string) {
	f.authService.AddVerifProvider(
		providerName,
		"{{ .Site }}\t{{ .User }}\t{{ .Token }}",
		provider.SenderFunc(func(address string, text string) (err error) {
			var fields []string
			if fields = strings.Split(text, "\t"); len(fields) != 3 {
				err = fmt.Errorf("error parsing auth email text")
				return
			}
			bodyCtx := beContext.New()
			bodyCtx.Set("UserAuthSiteID", fields[0])
			bodyCtx.Set("UserAuthUsername", fields[1])
			bodyCtx.Set("UserAuthToken", fields[2])
			// this does not do what you think it does lol
			// bodyCtx.Set("UserAuthURL", f.authOpts.URL+"/auth/"+providerName+"/login?site="+strcase.ToKebab(f.Enjin.SiteTag())+"token="+fields[2])
			var msg *gomail.Message
			if msg, err = f.emailProvider.NewEmail(f.verifyEmailTemplate, bodyCtx); err != nil {
				log.ErrorF("error creating new %v email: %v", f.verifyEmailTemplate, err)
				return
			}
			msg.SetHeader("To", address)
			// msg.SetHeader("Subject", "Verify new account")
			if err = f.Enjin.SendEmail(f.verifyEmailAccount, msg); err != nil {
				log.DebugF("error sending user auth email: %v", err)
			}
			return
		}),
	)
	return
}

func (f *CFeature) authAudSecretsFunc(aud string) (secret string, err error) {
	var ok bool
	if secret, ok = f.audSecrets[aud]; ok {
		log.DebugF(`using "%v" audience secret`, aud)
		return
	}
	if secret, ok = f.audSecrets["_default_"]; ok {
		if aud != "ignore" {
			log.DebugF(`using "_default_" audience secret for: "%v"`, aud)
		}
		return
	}
	err = fmt.Errorf(`audience not found: "%v"`, aud)
	return
}

func (f *CFeature) authClaimsUpdFunc(claims token.Claims) token.Claims {
	if claims.User != nil {

		if claims.User.Attributes == nil {
			claims.User.Attributes = make(map[string]interface{})
		}

		eid, _ := sha.DataHash10([]byte(claims.User.ID))

		if !f.publicSignups {
			// user must be present in the userbase for all logins
			if f.ubm.AuthUserPresent(eid) {
				if u, e := f.createOrUpdateAuthUser(claims.User); e != nil {
					log.ErrorF("error updating user: %v - %v", eid, e)
				} else {
					log.WarnF("***** auth claims updated authorized user: %#+v", u)
				}
			} else {
				log.ErrorF("unauthorized signup attempt: %#+v", claims.User)
				claims.User.Attributes["public-signup-denied"] = "true"
			}
		} else if u, e := f.createOrUpdateAuthUser(claims.User); e != nil {
			log.ErrorF("error updating user: %v - %v", claims.User.ID, e)
		} else {
			log.WarnF("***** auth claims updated public user: %#+v", u)
		}

		// if claims.User.Name == "dev_admin" {
		// 	// set attributes for dev_admin
		// 	claims.User.SetAdmin(true)
		// 	claims.User.SetStrAttr("custom-key", "some value")
		// }
	}
	return claims
}

func (f *CFeature) authValidatorFunc(token string, claims token.Claims) (valid bool) {

	if claims.User != nil {

		if denied, ok := claims.User.Attributes["public-signup-denied"]; ok && denied == "true" {
			valid = false
			log.WarnF("%v feature public user signup denied: %#+v", f.Tag(), claims.User)
			return
		}

		eid, _ := sha.DataHash10([]byte(claims.User.ID))
		if u, e := f.ubm.GetAuthUser(eid); e == nil {
			valid = u != nil
			return
		} else {
			log.ErrorF("error getting user for validation func: %v - %v", eid, e)
		}

	}

	return
}

func (f *CFeature) AuthApiServeHTTP(next http.Handler, w http.ResponseWriter, r *http.Request) {
	if cu := userbase.GetCurrentAuthUser(r); cu == nil {

		if tokenUser, err := token.GetUserInfo(r); err == nil {

			eid, _ := sha.DataHash10([]byte(tokenUser.ID))

			if ubAu, ee := f.ubm.GetAuthUser(eid); ee == nil {

				r = userbase.SetCurrentAuthUser(ubAu, r)
				r.URL.User = url.User(ubAu.EID)
				log.WarnF("setting current auth user: %v", ubAu.EID)

				if u, eee := f.ubm.GetUser(eid); eee == nil {

					r = userbase.SetCurrentUser(u, r)
					log.WarnF("setting current user: %v - groups=%#+v, actions=%#+v", u.EID, u.Groups, u.Actions)

				} else {
					log.ErrorRF(r, "error getting User - %v - %v", eid, eee)
				}

			} else {
				log.ErrorRF(r, "error getting AuthUser - %v - %v", eid, ee)
			}

		} else {
			log.DebugRF(r, "token user not present")
		}

	} else {
		log.DebugRF(r, "current user present: %v", cu.EID)
	}
	next.ServeHTTP(w, r)
}