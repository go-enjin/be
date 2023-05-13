// Copyright (c) 2022  The Go-Enjin Authors
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
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/net/headers"
	"github.com/go-enjin/be/pkg/theme"
)

type Builder interface {
	SiteTag(key string) Builder
	SiteName(name string) Builder
	SiteTagLine(title string) Builder
	SiteLanguageMode(mode lang.Mode) Builder
	SiteCopyrightName(name string) Builder
	SiteCopyrightNotice(notice string) Builder
	SiteDefaultLanguage(tag language.Tag) Builder
	SiteSupportedLanguages(tags ...language.Tag) Builder
	SiteLanguageDisplayNames(names map[language.Tag]string) Builder

	// Set a custom context key with value
	Set(key string, value interface{}) Builder

	// AddHtmlHeadTag adds a custom (singleton) HTML tag to the <head> section
	// of the page output, example meta tag:
	//   AddHtmlHeadTag("meta",map[string]string{"name":"og:thing","content":"stuff"})
	AddHtmlHeadTag(name string, attr map[string]string) Builder

	// AddDomains restricts inbound requests to only the domain names given
	AddDomains(domains ...string) Builder

	// AddFeatureNotes appends bullet-pointed notes to the CLI App's main
	// description field
	AddFeatureNotes(tag Tag, notes ...string) Builder

	// AddFlags adds custom command line flags
	AddFlags(flags ...cli.Flag) Builder

	// AddCommands adds custom command line commands
	AddCommands(commands ...*cli.Command) Builder

	// AddConsole adds custom command line go-curses consoles (ctk.Window)
	AddConsole(c Console) Builder

	// AddFeature includes the given feature within the built Enjin
	AddFeature(f Feature) Builder

	// AddRouteProcessor adds the given route processor to the Enjin route
	// processing middleware
	AddRouteProcessor(route string, processor ReqProcessFn) Builder

	// AddOutputTranslator adds the given output filter (for the given mime type)
	AddOutputTranslator(mime string, filter TranslateOutputFn) Builder

	// AddModifyHeadersFn adds the given headers.ModifyHeadersFn function to the
	// default headers enjin middleware layer
	AddModifyHeadersFn(fn headers.ModifyHeadersFn) Builder

	AddNotifyHook(name string, hook NotifyHook) Builder

	// AddPageFromString is a convenience wrapper around AddPage
	AddPageFromString(path, raw string) Builder

	// SetStatusPage overrides specific HTTP error pages, ie: 404
	SetStatusPage(status int, path string) Builder

	// AddTheme includes the given theme within the built Enjin
	AddTheme(t *theme.Theme) Builder

	// SetTheme configures the default theme
	SetTheme(name string) Builder

	// HotReload enables or disables hot-reloading theme templates and content files
	HotReload(enabled bool) Builder

	// Build constructs an Enjin Runner from the Builder configuration
	Build() Runner
}