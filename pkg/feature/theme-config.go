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
	"html/template"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
	"github.com/go-enjin/be/pkg/net/headers/policy/permissions"
)

type ThemeAuthor struct {
	Name     string
	Homepage string
}

type ThemeSupports struct {
	Menus      MenuSupports              `json:"menus,omitempty"`
	Layouts    []string                  `json:"layouts,omitempty"`
	Locales    []language.Tag            `json:"locales,omitempty"`
	Archetypes map[string]context.Fields `json:"archetypes,omitempty"`
}

type ThemeConfig struct {
	Name        string
	Parent      string
	License     string
	LicenseLink string
	Description string
	Homepage    string
	Authors     []ThemeAuthor
	Extends     string

	RootStyles  []template.CSS
	BlockStyles map[string][]template.CSS
	BlockThemes map[string]map[string]interface{}

	FontawesomeLinks   map[string]string
	FontawesomeClasses []string

	CacheControl string

	PermissionsPolicy     []permissions.Directive
	ContentSecurityPolicy csp.ContentSecurityPolicyConfig

	Supports ThemeSupports

	Context context.Context
}

func (tc *ThemeConfig) Copy() (config *ThemeConfig) {
	config = &ThemeConfig{
		Name:                  tc.Name,
		Parent:                tc.Parent,
		License:               tc.License,
		LicenseLink:           tc.LicenseLink,
		Description:           tc.Description,
		Homepage:              tc.Homepage,
		Authors:               tc.Authors,
		Extends:               tc.Extends,
		RootStyles:            tc.RootStyles[:],
		BlockStyles:           make(map[string][]template.CSS),
		BlockThemes:           make(map[string]map[string]interface{}),
		FontawesomeLinks:      maps.CopyBaseMap(tc.FontawesomeLinks),
		FontawesomeClasses:    tc.FontawesomeClasses[:],
		PermissionsPolicy:     tc.PermissionsPolicy[:],
		ContentSecurityPolicy: tc.ContentSecurityPolicy,
		Supports: ThemeSupports{
			Menus:      tc.Supports.Menus[:],
			Layouts:    tc.Supports.Layouts[:],
			Locales:    tc.Supports.Locales[:],
			Archetypes: make(map[string]context.Fields),
		},
		Context: tc.Context.Copy(),
	}
	for k, v := range tc.BlockStyles {
		config.BlockStyles[k] = v[:]
	}
	for k, v := range tc.BlockThemes {
		tc.BlockThemes[k] = maps.DeepCopy(v)
	}
	for k, v := range tc.Supports.Archetypes {
		tc.Supports.Archetypes[k] = v.Copy()
	}
	return
}