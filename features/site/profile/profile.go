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

package profile

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/message"

	beContext "github.com/go-enjin/be/pkg/context"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/menu"
	"github.com/go-enjin/be/pkg/request"
	"github.com/go-enjin/be/pkg/slices"
	"github.com/go-enjin/be/pkg/userbase"
	"github.com/go-enjin/be/types/site"
)

const (
	SetupNonceName = "site-profile--nonce"
	SetupNonceKey  = "site-profile--form"
)

var (
	DefaultProfileImagePath  = "/media/profiles"
	DefaultProfileImageNames = []string{
		"enjineer",
		"spongezero",
		"spongezero-astronaut",
		"spongezero-cosmic-without",
		"spongezero-cosmos-within",
		"spongezero-on-fire",
	}
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "site-profile"

type Feature interface {
	feature.SiteFeature
	feature.SiteUserSetupStage
}

type MakeFeature interface {
	feature.SiteMakeFeature[MakeFeature]

	// EnableSelfProfilePage enables (or disables) the site menu item and request handler for the (read-only) profile
	// pages
	EnableSelfProfilePage(enabled bool) MakeFeature

	// EnableOtherProfilePages enables (or disables) the (read-only) other user profile pages
	EnableOtherProfilePages(enabled bool) MakeFeature

	// EnableProfileImages enables the profile image aspects of this feature
	EnableProfileImages(enabled bool) MakeFeature
	// SetProfileImagePath enables profile images and specifies the public filesystem path prefix to use
	SetProfileImagePath(path string) MakeFeature
	// DefaultProfileImageNames enables profile images and adds the default image URLs to the media profiles list
	DefaultProfileImageNames() MakeFeature
	// AddProfileImageNames enables profile images and adds the given image URLs to the media profiles list
	AddProfileImageNames(names ...string) MakeFeature
	// SetProfileImageNames enables profile images and replaces the media profiles list with the given image names
	SetProfileImageNames(names ...string) MakeFeature

	Make() Feature
}

type CFeature struct {
	site.CSiteFeature[MakeFeature]

	selfProfilePageEnabled   bool
	otherProfilePagesEnabled bool

	profileImagesEnabled bool
	profileImageNames    []string
	profileImagePath     string

	viewOwnProfile   feature.Action
	viewOtherProfile feature.Action
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeatureKey("profile")
	f.SetSiteFeatureIcon("fa-solid fa-id-card")
	f.SetSiteFeatureLabel(func(printer *message.Printer) (label string) {
		label = printer.Sprintf("Profile")
		return
	})
	f.Construct(f)
	return f
}

func (f *CFeature) Construct(this interface{}) {
	f.CSiteFeature.Construct(this)
	f.CUsesActions.ConstructUsesActions(this)
	f.profileImagePath = DefaultProfileImagePath
	f.viewOwnProfile = feature.NewAction(f.Tag().String(), "view-own", "profile")
	f.viewOtherProfile = feature.NewAction(f.Tag().String(), "view-other", "profile")
	return
}

func (f *CFeature) Init(this interface{}) {
	f.CSiteFeature.Init(this)
	return
}

func (f *CFeature) EnableSelfProfilePage(enabled bool) MakeFeature {
	f.selfProfilePageEnabled = enabled
	return f
}

func (f *CFeature) EnableOtherProfilePages(enabled bool) MakeFeature {
	f.otherProfilePagesEnabled = enabled
	return f
}

func (f *CFeature) EnableProfileImages(enabled bool) MakeFeature {
	f.profileImagesEnabled = enabled
	return f
}

func (f *CFeature) SetProfileImagePath(path string) MakeFeature {
	f.EnableProfileImages(true)
	f.profileImagePath = path
	return f
}

func (f *CFeature) DefaultProfileImageNames() MakeFeature {
	f.EnableProfileImages(true)
	f.profileImageNames = slices.Merge(f.profileImageNames, DefaultProfileImageNames)
	return f
}

func (f *CFeature) AddProfileImageNames(images ...string) MakeFeature {
	f.EnableProfileImages(true)
	f.profileImageNames = slices.Merge(f.profileImageNames, images)
	return f
}

func (f *CFeature) SetProfileImageNames(images ...string) MakeFeature {
	f.EnableProfileImages(true)
	f.profileImageNames = images
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CSiteFeature.Build(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CSiteFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) UserActions() (list feature.Actions) {
	list = append(
		f.CSiteFeature.UserActions(),
		f.viewOwnProfile,
		f.viewOtherProfile,
	)

	return
}

func (f *CFeature) SiteFeatureMenu(r *http.Request) (m menu.Menu) {
	if f.selfProfilePageEnabled {
		info := f.SiteFeatureInfo(r)
		m = menu.Menu{
			{
				Text: info.Label,
				Href: f.SiteFeaturePath(),
				Icon: info.Icon,
			},
		}
	}
	return
}

func (f *CFeature) SiteSettingsFields(r *http.Request) (fields beContext.Fields) {
	printer := lang.GetPrinterFromRequest(r)

	fields = beContext.Fields{
		"display-name": {
			Key:          "display-name",
			Tab:          "user",
			Label:        printer.Sprintf("Display Name"),
			Format:       "string",
			Category:     "profile",
			Input:        "text",
			Weight:       7,
			Placeholder:  printer.Sprintf("Full Name"),
			NoResetValue: true,
		},
	}

	if f.profileImagesEnabled {
		var availableImages []string
		for _, name := range f.profileImageNames {
			imagePath := f.profileImagePath + "/" + name
			for _, extn := range []string{"png", "jpg", "gif", "webp"} {
				foundPath := imagePath + "." + extn
				if f.Enjin.PublicFileSystems().Lookup().FileExists(foundPath) {
					availableImages = append(availableImages, foundPath)
				}
			}
		}

		fields["profile-image"] = &beContext.Field{
			Key:          "profile-image",
			Tab:          "user",
			Label:        printer.Sprintf("Profile Image"),
			Format:       "url",
			Category:     "profile",
			Input:        "select",
			Weight:       8,
			DefaultValue: "",
			NoResetValue: true,
			ValueOptions: append([]string{""}, availableImages...),
		}
	}

	return
}

func (f *CFeature) RouteSiteFeature(r chi.Router) {
	if f.selfProfilePageEnabled {
		r.Get("/", f.ServeProfilePage)
	}
	if f.otherProfilePagesEnabled {
		r.Get("/{eid:[a-f0-9]{10}}", f.ServeProfilePage)
	}
}

func (f *CFeature) ServeProfilePage(w http.ResponseWriter, r *http.Request) {

	var eid string
	if other := chi.URLParam(r, "eid"); other != "" {
		if userbase.CurrentUserCan(r, f.viewOtherProfile) {
			eid = other
		}
	} else if userbase.CurrentUserCan(r, f.viewOwnProfile) {
		eid = userbase.GetCurrentEID(r)
	}

	if eid == "" {
		// permission denied
		f.Enjin.ServeNotFound(w, r)
		return
	}

	var err error
	var au feature.AuthUser
	if au, err = f.Site().SiteUsers().RetrieveUser(r, eid); err != nil {
		f.Enjin.ServeNotFound(w, r)
		return
	}

	ctx := beContext.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
		"User":        au.GetSettings(),
	}

	t := f.SiteFeatureTheme()
	if err = f.Site().PrepareAndServePage("site", "profile", f.SiteFeaturePath(), t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing %v feature page: %v", f.Tag(), err)
		f.Enjin.ServeInternalServerError(w, r)
		return
	}
}

func (f *CFeature) SiteUserSetupStageReady(eid string, r *http.Request) (ready bool) {
	au := userbase.GetCurrentAuthUser(r)
	if displayName, ok := au.GetSetting("display-name").(string); ok {
		ready = displayName != ""
	}
	return
}

func (f *CFeature) SiteUserSetupStageHandler(saf feature.SiteAuthFeature, w http.ResponseWriter, r *http.Request) {
	var err error
	var au feature.AuthUser
	printer := lang.GetPrinterFromRequest(r)

	su := f.Site().SiteUsers()
	eid := userbase.GetCurrentEID(r)

	su.LockUser(r, eid)
	defer su.UnlockUser(r, eid)

	if au, err = su.RetrieveUser(r, eid); err != nil {
		f.Enjin.ServeNotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {

		if nonce := request.SafeQueryFormValue(r, SetupNonceName); nonce != "" {
			if f.Enjin.VerifyNonce(SetupNonceKey, nonce) {

				if displayName := request.SafeQueryFormValue(r, "display-name"); displayName == "" {
					r = feature.AddErrorNotice(r, true, printer.Sprintf("A display name is required"))

				} else if err = f.Site().SiteUsers().SetUserSetting(r, eid, "display-name", displayName); err != nil {

					log.ErrorRF(r, "error setting user display-name setting: %v - %v", eid, err)
					r = feature.AddErrorNotice(r, true, berrs.UnexpectedError(printer))
				} else {
					// good-to-go!
					f.Enjin.ServeRedirect(r.URL.Path, w, r)
					return
				}

			}
		}

	}

	ctx := beContext.Context{
		"FeatureInfo": f.SiteFeatureInfo(r),
		"FormAction":  r.URL.Path,
		"Nonces": feature.Nonces{
			{Name: SetupNonceName, Key: SetupNonceKey},
		},
		"Name": au.GetName(),
	}

	t := f.SiteFeatureTheme()
	if err = f.Site().PrepareAndServePage("site", "profile--setup", f.SiteFeaturePath(), t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing %v feature page: %v", f.Tag(), err)
		f.Enjin.ServeInternalServerError(w, r)
		return
	}
	return
}