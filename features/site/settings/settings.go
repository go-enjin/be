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

package settings

import (
	"fmt"
	"html"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/maruel/natural"
	"github.com/microcosm-cc/bluemonday"
	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/menu"
	"github.com/go-enjin/be/pkg/request"
	"github.com/go-enjin/be/pkg/userbase"
	"github.com/go-enjin/be/types/site"
	"github.com/go-enjin/golang-org-x-text/message"
)

const (
	SettingPanelNonceName = "site-settings--nonce"
	SettingPanelNonceKey  = "site-settings--form"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "site-settings"

type Feature interface {
	feature.SiteFeature
}

type MakeFeature interface {
	feature.SiteMakeFeature[MakeFeature]

	Make() Feature
}

type CFeature struct {
	site.CSiteFeature[MakeFeature]
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeatureKey("settings")
	f.SetSiteFeatureIcon("fa-solid fa-gear")
	f.SetSiteFeatureLabel(func(printer *message.Printer) (label string) {
		label = printer.Sprintf("Settings")
		return
	})
	f.CSiteFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CSiteFeature.Init(this)
	return
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
	list = f.CSiteFeature.UserActions()
	return
}

func (f *CFeature) SiteFeatureMenu(r *http.Request) (m menu.Menu) {
	info := f.SiteFeatureInfo(r)
	m = menu.Menu{
		{
			Text: info.Label,
			Href: f.SiteFeaturePath(),
			Icon: info.Icon,
		},
	}
	return
}

func (f *CFeature) SiteSettings(r *http.Request) (settings map[string]beContext.Fields, order []string) {

	settings = make(map[string]beContext.Fields)
	for _, sf := range f.Site().SiteFeatures() {
		if fields := sf.SiteSettingsFields(r); fields.Len() > 0 {
			settings[sf.SiteFeatureKey()] = fields
		}
	}

	parsers := f.Enjin.PageContextParsers()
	for group, fields := range settings {
		for key, field := range fields {
			if field.Tab == "" {
				field.Tab = "page"
			}
			if field.Category == "" {
				field.Category = "general"
			}
			if fn, ok := parsers[field.Format]; ok {
				settings[group][key].Parse = fn
			} else {
				panic(fmt.Errorf("invalid field.Format: %q; possible values: %+v", field.Format, maps.SortedKeys(parsers)))
			}
		}
	}

	order = maps.Keys(settings)
	sort.Slice(order, func(i, j int) (less bool) {
		a, b := order[i], order[j]
		if less = a == "profile" && b != "profile"; less {
			return
		}
		less = natural.Less(a, b)
		return
	})
	return
}

func (f *CFeature) RouteSiteFeature(r chi.Router) {
	r.Post("/", f.ReceiveSettingsChanges)
	r.Get("/", f.ServeSettingsPage)
	for _, sf := range f.Site().SiteFeatures() {
		sfName := sf.SiteFeatureKey()
		settingsPath := f.SiteFeaturePath()
		if s, h := sf.SiteSettingsPanel(settingsPath + "/" + sfName); s != nil || h != nil {
			sfPath := "/" + sfName
			log.DebugF("%q feature routing settings page for: %q", f.Tag(), sfPath)
			r.Route(sfPath, func(r chi.Router) {
				r.Post("/*", h)
				r.Get("/*", s)
			})
		}
	}
}

func (f *CFeature) ServeSettingsPage(w http.ResponseWriter, r *http.Request) {

	// TODO: route settings panels changes if r.URL.Path != settings path

	au := userbase.GetCurrentAuthUser(r)
	matter := au.GetSettings()

	f.RenderSettingsWith(matter, w, r)
}

func (f *CFeature) RenderSettingsWith(matter beContext.Context, w http.ResponseWriter, r *http.Request) {
	t := f.SiteFeatureTheme()
	printer := lang.GetPrinterFromRequest(r)

	var order []string
	panels := make(map[string]string)
	settingsPath := f.SiteFeaturePath()
	for _, sf := range f.Site().SiteFeatures() {
		if s, _ := sf.SiteSettingsPanel(settingsPath); s != nil {
			name := sf.SiteFeatureLabel(printer)
			order = append(order, name)
			panels[name] = settingsPath + "/" + sf.SiteFeatureKey()
		}
	}

	settings, keys := f.SiteSettings(r)
	ctx := beContext.Context{
		"Title":                   printer.Sprintf(`Site Settings`),
		"UserSettings":            matter,
		"SiteSettings":            settings,
		"SiteSettingsKeys":        keys,
		"SiteSettingsPanelKeys":   order,
		"SiteSettingsPanelLookup": panels,
		"FormAction":              f.SiteFeaturePath(),
		"Nonces": feature.Nonces{
			{Name: SettingPanelNonceName, Key: SettingPanelNonceKey},
		},
	}

	if err := f.Site().PrepareAndServePage("site", "settings", f.SiteFeaturePath(), t, w, r, ctx); err != nil {
		log.ErrorRF(r, "error preparing %v feature page: %v", f.Tag(), err)
		f.Enjin.ServeInternalServerError(w, r)
		return
	}
}

func (f *CFeature) ReceiveSettingsChanges(w http.ResponseWriter, r *http.Request) {

	// TODO: route settings panels changes if r.URL.Path != settings path

	eid := userbase.GetCurrentEID(r)
	printer := lang.GetPrinterFromRequest(r)

	if nonce := request.SafeQueryFormValue(r, SettingPanelNonceName); nonce != "" {
		if !f.Enjin.VerifyNonce(SettingPanelNonceKey, nonce) {
			f.Site().PushErrorNotice(eid, true, berrs.FormExpiredError(printer))
			f.Enjin.ServeRedirect(f.SiteFeaturePath(), w, r)
			return
		}
	} else {
		f.Site().PushErrorNotice(eid, true, berrs.IncompleteFormError(printer))
		f.Enjin.ServeRedirect(f.SiteFeaturePath(), w, r)
		return
	}

	action, _ := feature.ParseEditorOpKey(r.PostFormValue("submit"))
	if action = strings.ToLower(strings.TrimSpace(action)); action == "cancel" {
		f.Site().PushInfoNotice(eid, true, printer.Sprintf("Changes discarded."))
		f.Enjin.ServeRedirect(f.SiteFeaturePath(), w, r)
		return
	}

	form := map[string]interface{}{}
	for _, k := range maps.SortedKeys(r.Form) {
		v := r.Form[k]
		for i := 0; i < len(v); i++ {
			v[i] = html.UnescapeString(v[i])
		}
		switch len(v) {
		case 0: // nop
		case 1:
			_ = maps.Set(k, v[0], form)
		case 2:
			if v[0] == v[1] {
				_ = maps.Set(k, v[0], form)
			} else {
				_ = maps.Set(k, v, form)
			}
		default:
			_ = maps.Set(k, v, form)
		}
	}

	var formMatter beContext.Context
	if bc, ok := form["matter"].(beContext.Context); ok {
		formMatter = bc
	} else if fm, ok := form["matter"].(map[string]interface{}); ok {
		formMatter = fm
	} else {
		// nop
	}

	strict := bluemonday.StrictPolicy()
	var parseCustom func(v interface{}) (parsed interface{})
	parseCustom = func(v interface{}) (parsed interface{}) {
		if value, ok := v.(string); ok {
			value = strict.Sanitize(value)
			parsed = html.UnescapeString(value)
		} else if list, ok := v.([]interface{}); ok {
			var items []interface{}
			for _, item := range list {
				items = append(items, parseCustom(item))
			}
			parsed = items
		} else if dictionary, ok := v.(map[string]interface{}); ok {
			cleaned := make(map[string]interface{})
			for dk, dv := range dictionary {
				cleaned[dk] = parseCustom(dv)
			}
			parsed = cleaned
		}
		return
	}

	errs := make(map[string]error)
	settings, settingGroups := f.SiteSettings(r)

	lookup := func(key string) (field *beContext.Field, ok bool) {
		for _, group := range settingGroups {
			if field, ok = settings[group].Lookup(key); ok {
				return
			}
		}
		return
	}

	matter := beContext.Context{}

	for k, v := range formMatter.AsDeepKeyed() {
		if field, ok := lookup(k); ok {
			if field.Input != "checkbox" {
				if vi, ee := field.Parse(field, v); ee != nil {
					errs[field.Key] = ee
				} else {
					//field.Value = vi
					_ = maps.Set(k, vi, matter)
				}
			}
		} else {
			log.WarnRF(r, "strict policy for custom field: %q", k)
			_ = maps.Set(k, parseCustom(v), matter)
		}
	}

	for _, group := range settingGroups {
		for _, field := range settings[group] {
			if field.Input == "checkbox" {
				fv := r.FormValue(".matter." + field.Key)
				_ = maps.Set("."+field.Key, fv != "", matter)
			}
		}
	}

	var err error
	var au feature.AuthUser
	if au, err = f.Site().SiteUsers().RetrieveUser(r, eid); err != nil {
		au = userbase.GetCurrentAuthUser(r)
	}

	uCtx := au.GetSettings()
	uCtx.KebabKeys()
	uCtx.ApplySpecific(matter)

	if err = f.Site().SiteUsers().SetUserContext(r, eid, beContext.Context{"settings": uCtx}); err != nil {
		log.ErrorRF(r, "error saving user settings changes: %v - %v", eid, err)
		f.Site().PushErrorNotice(eid, true, berrs.UnexpectedError(printer))
		f.RenderSettingsWith(matter, w, r)
		return
	}

	f.Enjin.ServeRedirect(f.SiteFeaturePath(), w, r)
}