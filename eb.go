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

package be

import (
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/language"

	beCli "github.com/go-enjin/be/pkg/cli"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/net/headers"
)

var _ feature.Builder = (*EnjinBuilder)(nil)

type htmlHeadTag struct {
	name string
	attr map[string]string
}

type EnjinBuilder struct {
	flags        []cli.Flag
	commands     cli.Commands
	pages        map[string]feature.Page
	htmlHeadTags []htmlHeadTag
	context      context.Context
	theme        string
	theming      map[string]feature.Theme
	themeOrder   []string
	features     *feature.FeaturesCache
	headers      []headers.ModifyHeadersFn
	domains      []string
	consoles     map[feature.Tag]feature.Console
	processors   map[string]feature.ReqProcessFn
	translators  map[string]feature.TranslateOutputFn
	transformers map[string]feature.TransformOutputFn
	slugsums     bool
	statusPages  map[int]string
	hotReload    bool

	tag             string
	name            string
	copyrightName   string
	copyrightNotice string
	tagLine         string

	langMode    lang.Mode
	localeTags  []language.Tag
	localeNames map[language.Tag]string
	defaultLang language.Tag

	publicUser  feature.Actions
	userActions feature.Actions

	presets []feature.Preset

	fFormatProviders                []feature.PageFormatProvider
	fRequestFilters                 []feature.RequestFilter
	fPageContextUpdaters            []feature.PageContextUpdater
	fPageContextModifiers           []feature.PageContextModifier
	fPageRestrictionHandlers        []feature.PageRestrictionHandler
	fMenuProviders                  []feature.MenuProvider
	fDataRestrictionHandlers        []feature.DataRestrictionHandler
	fOutputTranslators              []feature.OutputTranslator
	fOutputTransformers             []feature.OutputTransformer
	fPageTypeProcessors             []feature.PageTypeProcessor
	fServePathFeatures              []feature.ServePathFeature
	fDatabases                      []feature.Database
	fEmailSenders                   []feature.EmailSender
	fRequestModifiers               []feature.RequestModifier
	fRequestRewriters               []feature.RequestRewriter
	fPermissionsPolicyModifiers     []feature.PermissionsPolicyModifier
	fContentSecurityPolicyModifiers []feature.ContentSecurityPolicyModifier
	fUseMiddlewares                 []feature.UseMiddleware
	fHeadersModifiers               []feature.HeadersModifier
	fProcessors                     []feature.Processor
	fApplyMiddlewares               []feature.ApplyMiddleware
	fPageProviders                  []feature.PageProvider
	fFileProviders                  []feature.FileProvider
	fQueryIndexFeatures             []feature.QueryIndexFeature
	fPageContextProviders           []feature.PageContextProvider
	fAuthProviders                  []feature.AuthProvider
	fUserActionsProviders           []feature.UserActionsProvider
	fEnjinContextProvider           []feature.EnjinContextProvider
	fPageShortcodeProcessors        []feature.PageShortcodeProcessor
	fFuncMapProviders               []feature.FuncMapProvider
	fTemplatePartialsProvider       []feature.TemplatePartialsProvider
	fThemeRenderers                 []feature.ThemeRenderer
	fServiceLoggers                 []feature.ServiceLogger
	fLocalesProviders               []feature.LocalesProvider
	fPrepareServePagesFeatures      []feature.PrepareServePagesFeature
	fFinalizeServePagesFeatures     []feature.FinalizeServePagesFeature
	fPageContextFieldsProviders     []feature.PageContextFieldsProvider
	fPageContextParsersProviders    []feature.PageContextParsersProvider

	fPanicHandler      feature.PanicHandler
	fServiceListener   feature.ServiceListener
	fRoutePagesHandler feature.RoutePagesHandler
	fServePagesHandler feature.ServePagesHandler
	fServiceLogHandler feature.ServiceLogHandler

	enjins []*EnjinBuilder

	publicFileSystems fs.Registry

	buildPages map[string]string
}

func New() (be *EnjinBuilder) {
	be = new(EnjinBuilder)
	be.theme = ""
	be.flags = make([]cli.Flag, 0)
	be.commands = make(cli.Commands, 0)
	be.pages = make(map[string]feature.Page)
	be.context = context.New()
	be.theming = make(map[string]feature.Theme)
	be.features = feature.NewFeaturesCache()
	be.headers = make([]headers.ModifyHeadersFn, 0)
	be.domains = make([]string, 0)
	be.consoles = make(map[feature.Tag]feature.Console)
	be.processors = make(map[string]feature.ReqProcessFn)
	be.translators = make(map[string]feature.TranslateOutputFn)
	be.transformers = make(map[string]feature.TransformOutputFn)
	be.slugsums = true
	be.statusPages = make(map[int]string)
	be.hotReload = false
	be.langMode = lang.NewQueryMode().Make()
	be.defaultLang = language.Und
	be.publicUser = make(feature.Actions, 0)
	be.buildPages = make(map[string]string)
	return be
}

func (eb *EnjinBuilder) IncludeEnjin(enjins ...*EnjinBuilder) feature.Builder {
	var checked []*EnjinBuilder
	for _, enjin := range enjins {
		for _, check := range checked {
			if check.tag == enjin.tag {
				log.FatalDF(1, "enjin tags must be unique")
			}
		}
		checked = append(checked, enjin)
	}
	eb.enjins = append(eb.enjins, checked...)
	return eb
}

func (eb *EnjinBuilder) HotReload(enabled bool) feature.Builder {
	eb.hotReload = enabled
	return eb
}

func (eb *EnjinBuilder) IgnoreSlugsums() *EnjinBuilder {
	eb.slugsums = false
	return eb
}

func (eb *EnjinBuilder) prepareBuild() {
	var err error

	if err = eb.resolveFeatureDeps(); err != nil {
		log.FatalDF(2, "error resolving feature dependencies: %v", err)
	}

	if eb.tag == "" {
		log.FatalDF(2, "missing .SiteTag")
	}
	eb.Set("SiteTag", eb.tag)
	if eb.name == "" {
		log.FatalDF(2, "missing .SiteName")
	}
	eb.Set("SiteName", eb.name)

	if eb.copyrightName != "" {
		eb.Set("CopyrightName", eb.copyrightName)
	}
	if eb.copyrightNotice == "" && eb.copyrightName != "" {
		eb.copyrightNotice = fmt.Sprintf("© %v", time.Now().Year())
		eb.Set("CopyrightNotice", eb.copyrightNotice)
	} else {
		eb.Set("CopyrightNotice", eb.copyrightNotice)
	}
	if eb.tagLine != "" {
		eb.Set("SiteTagLine", eb.tagLine)
	}

	eb.Set("LanguageMode", eb.langMode)
	eb.Set("DefaultLanguage", eb.defaultLang.String())
	eb.Set("DefaultLanguageTag", eb.defaultLang)

	eb.publicFileSystems = fs.NewRegistry(eb.tag)

	for _, f := range eb.features.List() {
		if err = f.Self().Build(eb); err != nil {
			log.FatalDF(2, "feature [%v] - %v", f.Tag(), err)
			return
		}
	}

	if eb.theme != "" {
		if _, ok := eb.theming[eb.theme]; !ok {
			log.FatalDF(2, "specified theme not found: %v", eb.theme)
		}
	}

	if eb.theme == "" {
		for k, _ := range eb.theming {
			eb.theme = k
			break
		}
	}

	for _, t := range eb.theming {
		// add format providers to all themes
		for _, p := range eb.fFormatProviders {
			t.AddFormatProvider(p)
		}
	}

	eb.buildPagesFromStrings()

	for tag, console := range eb.consoles {
		if err = console.Build(eb); err != nil {
			log.FatalDF(2, "console [%v] - %v", tag, err)
			return
		}
	}

	if len(eb.htmlHeadTags) > 0 {
		var tags []template.HTML
		for _, tag := range eb.htmlHeadTags {
			var attrs string
			for _, key := range maps.ReverseSortedKeys(tag.attr) {
				if attrs != "" {
					attrs += " "
				}
				attrs += fmt.Sprintf(`%v="%v"`, key, template.HTMLEscapeString(tag.attr[key]))
			}
			if attrs == "" {
				// log.FatalF("SetHtmlHeadTag #%d requires attributes", idx+1)
				continue
			}
			tags = append(
				tags,
				template.HTML(
					fmt.Sprintf("<%v %v/>\n", tag.name, attrs),
				),
			)
		}
		eb.Set("HtmlHeadTags", tags)
	}
}

func (eb *EnjinBuilder) Build() feature.Runner {
	eb.prepareBuild()

	eb.flags = append(
		eb.flags,
		&cli.StringFlag{
			Name:     "prefix",
			Usage:    "for dev and stg sites to prefix labels",
			Value:    os.Getenv("USER"),
			Aliases:  []string{"P"},
			EnvVars:  eb.MakeEnvKeys("PREFIX"),
			Category: "general",
		},
		&cli.BoolFlag{
			Name:     "quiet",
			Usage:    "set log level to WARN",
			Aliases:  []string{"q"},
			EnvVars:  eb.MakeEnvKeys("QUIET"),
			Category: "general",
		},
		&cli.BoolFlag{
			Name:     "debug",
			Usage:    "enable verbose logging for debugging purposes",
			EnvVars:  eb.MakeEnvKeys("DEBUG"),
			Category: "general",
		},
		&cli.StringFlag{
			Name:     "log-level",
			Usage:    "set logging level: error, warn, info, debug or trace",
			EnvVars:  eb.MakeEnvKeys("LOG_LEVEL"),
			Category: "general",
		},
		&cli.StringSliceFlag{
			Name:     "domain",
			Usage:    "restrict inbound requests to only the domain names given",
			EnvVars:  eb.MakeEnvKeys("DOMAIN"),
			Category: "service",
		},
		&cli.BoolFlag{
			Name:     "strict",
			Usage:    "use strict Slugsums validation (extraneous files are errors)",
			Aliases:  []string{"s"},
			EnvVars:  eb.MakeEnvKeys("STRICT"),
			Category: "security",
		},
		&cli.StringFlag{
			Name:     "sums-integrity",
			Usage:    "specify the sha256sum of the Shasums file for --strict validations",
			Value:    "",
			EnvVars:  eb.MakeEnvKeys("SUMS_INTEGRITY_" + strcase.ToScreamingSnake(globals.BinHash)),
			Category: "security",
		},
	)

	for _, enjin := range eb.enjins {
		enjin.prepareBuild()
		tag := strcase.ToKebab(enjin.tag)
		key := strcase.ToScreamingSnake(tag)
		eb.flags = append(eb.flags, &cli.StringSliceFlag{
			Name:     tag + "-domain",
			Usage:    "specify one or more domains for the " + tag + " enjin",
			EnvVars:  eb.MakeEnvKeys(key + "_DOMAIN"),
			Category: "service",
		})
		for _, f := range enjin.flags {
			names := f.Names()
			if !beCli.FlagInFlags(names[0], eb.flags) {
				eb.flags = append(eb.flags, f)
			}
		}
		for _, c := range enjin.commands {
			if !beCli.CommandInCommands(c.Name, eb.commands) {
				eb.commands = append(eb.commands, c)
			}
		}
	}

	eb.AddUserAction(
		feature.NewAction("enjin", "view", "page"),
	)

	return newEnjin(eb)
}