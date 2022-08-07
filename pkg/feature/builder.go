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
	"embed"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/net/headers"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/theme"
)

type Builder interface {
	// Set a custom context key with value
	Set(key string, value interface{}) Builder

	// AddDomains restricts inbound requests to only the domain names given
	AddDomains(domains ...string) Builder

	// AddFlags adds custom command line flags
	AddFlags(flags ...cli.Flag) Builder

	// AddCommands adds custom command line commands
	AddCommands(commands ...*cli.Command) Builder

	// AddFeature includes the given feature within the built Enjin
	AddFeature(f Feature) Builder

	// AddRouteProcessor adds the given route processor to the Enjin route
	// processing middleware
	AddRouteProcessor(route string, processor ReqProcessFn) Builder

	// AddOutputFilter adds the given output filter (for the given mime type)
	AddOutputTranslator(mime string, filter TranslateOutputFn) Builder

	// AddModifyHeadersFn adds the given headers.ModifyHeadersFn function to the
	// default headers enjin middleware layer
	AddModifyHeadersFn(fn headers.ModifyHeadersFn) Builder

	AddNotifyHook(name string, hook NotifyHook) Builder

	// AddPage includes the given page within the pages enjin middleware
	AddPage(p *page.Page) Builder

	// AddPageFromString is a convenience wrapper around AddPage
	AddPageFromString(path, raw string) Builder

	// AddTheme includes the given theme within the built Enjin
	AddTheme(t *theme.Theme) Builder

	// SetTheme configures the default theme
	SetTheme(name string) Builder

	// AddThemes is a convenience wrapper include all themes in the given path
	AddThemes(path string) Builder

	// EmbedTheme is a wrapper around AddTheme for an embed.FS theme
	EmbedTheme(name, path string, tfs embed.FS) Builder

	// EmbedThemes is a wrapper to include all themes in the given embed.FS
	EmbedThemes(path string, fs embed.FS) Builder

	// Build constructs an Enjin Runner from the Builder configuration
	Build() Runner
}