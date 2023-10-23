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
	"github.com/go-chi/chi/v5"

	"github.com/go-enjin/be/pkg/feature/signaling"
	"github.com/go-enjin/be/pkg/menu"
)

type SiteFeature interface {
	Feature
	signaling.Signaling

	Site() (s Site)

	SiteFeatureName() (name string)
	SiteFeaturePath() (path string)
	SiteFeatureMenu() (m menu.Menu)
	SiteFeatureTheme() (t Theme)

	SetupSiteFeature(s Site)
	RouteSiteFeature(r chi.Router)
}

type SiteMakeFeature[MakeTypedFeature interface{}] interface {
	SetSiteFeatureName(name string) MakeTypedFeature
	SetSiteFeatureTheme(name string) MakeTypedFeature
}
