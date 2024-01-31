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

	"github.com/go-corelibs/x-text/language"

	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/net/headers"
)

type Builder interface {
	// SiteTag is the programmatic tag for identifying this particular enjin
	SiteTag(key string) Builder
	// SiteName is the name of the site, typically used in the branding header and browser tab titles
	SiteName(name string) Builder
	// SiteTagLine is an optional bit of text to display under the site name in the branding header
	SiteTagLine(title string) Builder

	// SiteCopyrightName specifies the first component of a standard copyright declaration for the site
	SiteCopyrightName(name string) Builder
	// SiteCopyrightYear specifies the year component of a standard copyright declaration and if left empty, the current
	// year is used in its place
	SiteCopyrightYear(year string) Builder
	// SiteCopyrightNotice specifies the last component of a standard copyright declaration and if left empty, the
	// phrase "All rights reserved" is used in its place
	SiteCopyrightNotice(notice string) Builder

	// SetEnjinTextFn is the means for translating the site name, tag line and copyright components at runtime using a
	// language printer for every request to the site
	SetEnjinTextFn(fn EnjinTextFn) Builder

	SiteLanguageMode(mode lang.Mode) Builder
	SiteDefaultLanguage(tag language.Tag) Builder
	SiteSupportedLanguages(tags ...language.Tag) Builder
	SiteLanguageDisplayNames(names map[language.Tag]string) Builder

	// Set a custom context key with value
	Set(key string, value interface{}) Builder

	// SetAlwaysHtmlRedirect configures the enjin to always use the HTML meta tag http-equiv=refresh method of
	// redirecting visitors to another page, the default is to use HTTP status 303 (See other) header redirections
	// unless explicitly calling the enjin ServeHtmlRedirect method
	SetAlwaysHtmlRedirect(always bool) Builder

	// SetHtmlRedirectDelay specifies the number of seconds to delay visitors when performing ServeHtmlRedirect calls,
	// the default is 0 seconds (immediately on load)
	SetHtmlRedirectDelay(seconds int) Builder

	// AddPreset includes the given presets just before the normal enjin Build() phase, presets are added in the same
	// order given and any features added will be inserted before all other enjin features present
	AddPreset(presets ...Preset) Builder

	// AddHtmlHeadTag adds a custom (singleton) HTML tag to the <head> section
	// of the page output, example meta tag:
	//   AddHtmlHeadTag("meta",map[string]string{"name":"og:thing","content":"stuff"})
	AddHtmlHeadTag(name string, attr map[string]string) Builder

	// AddDomains restricts inbound requests to only the domain names given
	AddDomains(domains ...string) Builder

	// AddFlags adds custom command line flags
	AddFlags(flags ...cli.Flag) Builder

	// AddCommands adds custom command line commands
	AddCommands(commands ...*cli.Command) Builder

	// AddConsole adds custom command line go-curses consoles (ctk.Window)
	AddConsole(c Console) Builder

	// AddFeature appends the given feature within the built Enjin (appends it to the end of the list of features
	// included)
	AddFeature(f ...Feature) Builder

	// PrependFeature prepends the given feature within the built Enjin (inserts it at the start of the list of
	// features included)
	PrependFeature(f Feature) Builder

	// AddRouteProcessor adds the given route processor to the Enjin route
	// processing middleware
	AddRouteProcessor(route string, processor ReqProcessFn) Builder

	// AddOutputTranslator adds the given output filter (for the given mime type)
	AddOutputTranslator(mime string, filter TranslateOutputFn) Builder

	// AddModifyHeadersFn adds the given headers.ModifyHeadersFn function to the
	// default headers enjin middleware layer
	AddModifyHeadersFn(fn headers.ModifyHeadersFn) Builder

	// AddModifyContentSecurityPolicyFn includes the given CSP modifier function with each request
	AddCspModifierFn(tag string, fn CspModifierFn) Builder

	AddNotifyHook(name string, hook NotifyHook) Builder

	// AddPageFromString is a convenience wrapper around AddPage
	AddPageFromString(path, raw string) Builder

	// SetStatusPage overrides specific HTTP error pages, ie: 404
	SetStatusPage(status int, path string) Builder

	// AddTheme includes the given theme within the built Enjin
	AddTheme(t Theme) Builder

	// SetTheme configures the default theme
	SetTheme(name string) Builder

	// HotReload enables or disables hot-reloading theme templates and content files
	HotReload(enabled bool) Builder

	// SetPublicAccess specifies the list of actions unauthorized visitors have
	//
	// Note: not specifying any public access results in no public access given
	SetPublicAccess(actions ...Action) Builder

	// AddUserAction adds the given actions to the overall known list of
	// registered user actions across all features
	AddUserAction(actions ...Action) Builder

	// Build constructs an Enjin Runner from the Builder configuration
	Build() Runner
}
