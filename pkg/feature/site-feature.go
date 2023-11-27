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

	"github.com/go-chi/chi/v5"

	"github.com/go-enjin/golang-org-x-text/message"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature/signaling"
	"github.com/go-enjin/be/pkg/menu"
)

type SiteFeatures TypedFeatures[SiteFeature]

type SiteFeatureLabelFn func(printer *message.Printer) (label string)

type SiteMakeFeature[MakeTypedFeature interface{}] interface {
	SetSiteFeatureKey(key string) MakeTypedFeature
	SetSiteFeatureIcon(icon string) MakeTypedFeature
	SetSiteFeatureTheme(name string) MakeTypedFeature
	SetSiteFeatureLabel(fn SiteFeatureLabelFn) MakeTypedFeature
}

type SiteFeature interface {
	Feature
	UserActionsProvider
	signaling.Signaling

	Site() (s Site)

	Action(verb string, details ...string) (action Action)

	SiteFeatureInfo(r *http.Request) (info *CSiteFeatureInfo)

	SiteFeatureKey() (name string)
	SiteFeatureIcon() (icon string)
	SiteFeaturePath() (path string)
	SiteFeatureMenu(r *http.Request) (m menu.Menu)
	SiteFeatureTheme() (t Theme)
	SiteFeatureLabel(printer *message.Printer) (label string)

	SetupSiteFeature(s Site) (err error)
	RouteSiteFeature(r chi.Router)

	SiteSettingsFields(r *http.Request) (fields beContext.Fields)
	SiteSettingsPanel(settingsPath string) (serve, handle http.HandlerFunc)
}

type SiteRootFeature interface {
	SiteFeature

	SiteRootHandler() (this http.Handler)
}

type SiteUserRequestModifier interface {
	SiteFeature

	ModifyUserRequest(au AuthUser, r *http.Request) (modified *http.Request)
}
