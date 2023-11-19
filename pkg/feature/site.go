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

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature/signaling"
)

type Site interface {
	Feature
	signaling.Signaling

	SitePath() (path string)
	SiteMenu(r *http.Request) (siteMenu beContext.Context)
	SiteTheme() (t Theme)
	SiteUsers() (sup SiteUsersProvider)
	SiteAuth() (sup SiteAuthFeature)

	SiteFeatures() (list SiteFeatures)

	PushImportantNotice(eid string, dismiss bool, message string, argv ...interface{})
	PushInfoNotice(eid string, dismiss bool, message string, argv ...interface{})
	PushWarnNotice(eid string, dismiss bool, message string, argv ...interface{})
	PushErrorNotice(eid string, dismiss bool, message string, argv ...interface{})
	PushNotices(eid string, notices ...*UserNotice)
	PullNotices(eid string) (notices UserNotices)

	GetContext(eid string) (ctx beContext.Context)
	SetContext(eid string, ctx beContext.Context)

	PreparePage(layout, pageType, pagePath string, t Theme, r *http.Request) (pg Page, ctx beContext.Context, err error)
	ServePreparedPage(pg Page, ctx beContext.Context, t Theme, w http.ResponseWriter, r *http.Request)
	PrepareAndServePage(layout, pageType, pagePath string, t Theme, w http.ResponseWriter, r *http.Request, ctx beContext.Context) (err error)

	RequireVerification(path string, w http.ResponseWriter, r *http.Request) (allowed bool)
	MustRequireVerification(path string, w http.ResponseWriter, r *http.Request) (allowed bool)
}