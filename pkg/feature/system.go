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
	"net/http"

	"github.com/Shopify/gomail"
	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/go-enjin/golang-org-x-text/message"
	"github.com/go-enjin/golang-org-x-text/message/catalog"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature/signaling"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/net/headers"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
	"github.com/go-enjin/be/pkg/net/headers/policy/permissions"
)

type EnjinBase interface {
	SiteTag() (key string)
	SiteName() (name string)
	SiteTagLine() (tagLine string)

	SiteCopyrightName() (name string)
	SiteCopyrightYear() (year string)
	SiteCopyrightNotice() (notice string)

	SiteLocales() (locales lang.Tags)
	SiteLanguageMode() (mode lang.Mode)
	SiteLanguageCatalog() (c catalog.Catalog)
	SiteDefaultLanguage() (tag language.Tag)
	SiteSupportsLanguage(tag language.Tag) (supported bool)
	SiteLanguageDisplayName(tag language.Tag) (name string, ok bool)

	FindTranslations(url string) (pages Pages)
	FindTranslationUrls(url string) (pages map[language.Tag]string)
	FindPage(tag language.Tag, url string) (p Page)
	FindPages(prefix string) (pages []Page)
}

type Service interface {
	Prefix() (prefix string)
	Context() (ctx context.Context)

	GetTheme() (t Theme, err error)
	MustGetTheme() (t Theme)
	ThemeNames() (names []string)
	GetThemeName() (name string)
	GetThemeNamed(name string) (t Theme, err error)
	MustGetThemeNamed(name string) (t Theme)

	ServerName() (name string)
	ServiceInfo() (scheme, listen string, port int)

	PermissionsPolicy() (handler *permissions.PolicyHandler)
	ContentSecurityPolicy() (handler *csp.PolicyHandler)

	FinalizeServeRequest(w http.ResponseWriter, r *http.Request)

	ServeRedirect(destination string, w http.ResponseWriter, r *http.Request)
	ServeRedirectHomePath(w http.ResponseWriter, r *http.Request)

	Serve204(w http.ResponseWriter, r *http.Request)
	Serve400(w http.ResponseWriter, r *http.Request)
	Serve401(w http.ResponseWriter, r *http.Request)
	ServeBasic401(w http.ResponseWriter, r *http.Request)
	Serve403(w http.ResponseWriter, r *http.Request)
	Serve404(w http.ResponseWriter, r *http.Request)
	Serve405(w http.ResponseWriter, r *http.Request)
	Serve500(w http.ResponseWriter, r *http.Request)

	ServeNotFound(w http.ResponseWriter, r *http.Request)
	ServeForbidden(w http.ResponseWriter, r *http.Request)
	ServeInternalServerError(w http.ResponseWriter, r *http.Request)

	ServeStatusPage(status int, w http.ResponseWriter, r *http.Request)
	ServePage(p Page, w http.ResponseWriter, r *http.Request) (err error)
	ServePath(urlPath string, w http.ResponseWriter, r *http.Request) (err error)
	ServeJSON(v interface{}, w http.ResponseWriter, r *http.Request) (err error)
	ServeStatusJSON(status int, v interface{}, w http.ResponseWriter, r *http.Request) (err error)
	ServeData(data []byte, mime string, w http.ResponseWriter, r *http.Request)

	MatchQL(query string) (pages []Page)
	MatchStubsQL(query string) (stubs []*PageStub)
	SelectQL(query string) (selected map[string]interface{})

	CheckMatchQL(query string) (pages []Page, err error)
	CheckMatchStubsQL(query string) (stubs []*PageStub, err error)
	CheckSelectQL(query string) (selected map[string]interface{}, err error)

	FindPageStub(shasum string) (stub *PageStub)
	FindPage(tag language.Tag, url string) (p Page)
	FindFile(path string) (data []byte, mime string, err error)

	FindEmailAccount(account string) (emailSender EmailSender)
	SendEmail(r *http.Request, account string, message *gomail.Message) (err error)

	GetPublicAccess() (actions Actions)
	FindAllUserActions() (list Actions)

	NonceFactory
	TokenFactory
	SyncLockerFactory

	Notify(tag string)
	NotifyF(tag, format string, argv ...interface{})

	TranslateShortcodes(content string, ctx context.Context) (modified string)

	GetThemeRenderer(ctx context.Context) (renderer ThemeRenderer)

	signaling.EmitterSupport
}

type System interface {
	Service
	signaling.Signaling

	Router() (router *chi.Mux)
}

type RootInternals interface {
	Internals

	SetupRootEnjin(ctx *cli.Context) (err error)
}

type Internals interface {
	Service
	signaling.Signaling
	EnjinBase

	Self() (self interface{})

	Features() (cache *FeaturesCache)

	Pages() (pages map[string]Page)
	Theme() (theme string)
	Theming() (theming map[string]Theme)
	Headers() (headers []headers.ModifyHeadersFn)
	Domains() (domains []string)
	Consoles() (consoles map[Tag]Console)
	Processors() (processors map[string]ReqProcessFn)
	Translators() (translators map[string]TranslateOutputFn)
	Transformers() (transformers map[string]TransformOutputFn)
	Slugsums() (enabled bool)

	ReloadLocales()
	HotReloading() (enabled bool)

	DB(tag string) (db interface{}, err error)
	MustDB(tag string) (db interface{})
	SpecificDB(fTag Tag, tag string) (db interface{}, err error)
	MustSpecificDB(fTag Tag, tag string) (db interface{})

	MakeFuncMap(ctx context.Context) (fm FuncMap)

	PublicFileSystems() (registry fs.Registry)

	ListTemplatePartials(block, position string) (names []string)
	GetTemplatePartial(block, position, name string) (tmpl string, ok bool)

	MakeLanguagePrinter(requested string) (tag language.Tag, printer *message.Printer)

	PublicUserActions() (actions Actions)
	ValidateUserRequest(action Action, w http.ResponseWriter, r *http.Request) (valid bool)

	PageContextParsers() (parsers context.Parsers)
	MakePageContextField(key string, r *http.Request) (field *context.Field, ok bool)
	MakePageContextFields(r *http.Request) (fields context.Fields)

	GetFormatProviders() []PageFormatProvider
	GetRequestFilters() []RequestFilter
	GetPageContextModifiers() []PageContextModifier
	GetPageContextUpdaters() []PageContextUpdater
	GetPageRestrictionHandlers() []PageRestrictionHandler
	GetMenuProviders() []MenuProvider
	GetDataRestrictionHandlers() []DataRestrictionHandler
	GetOutputTranslators() []OutputTranslator
	GetOutputTransformers() []OutputTransformer
	GetPageTypeProcessors() []PageTypeProcessor
	GetServePathFeatures() []ServePathFeature
	GetDatabases() []Database
	GetEmailSenders() []EmailSender
	GetRequestModifiers() []RequestModifier
	GetRequestRewriters() []RequestRewriter
	GetPermissionsPolicyModifiers() []PermissionsPolicyModifier
	GetContentSecurityPolicyModifiers() []ContentSecurityPolicyModifier
	GetUseMiddlewares() []UseMiddleware
	GetHeadersModifiers() []HeadersModifier
	GetProcessors() []Processor
	GetApplyMiddlewares() []ApplyMiddleware
	GetPageProviders() []PageProvider
	GetFileProviders() []FileProvider
	GetQueryIndexFeatures() []QueryIndexFeature
	GetPageContextProviders() []PageContextProvider
	GetAuthProviders() []AuthProvider
	GetUserActionsProviders() []UserActionsProvider
	GetEnjinContextProvider() []EnjinContextProvider
	GetPageShortcodeProcessors() []PageShortcodeProcessor
	GetFuncMapProviders() []FuncMapProvider
	GetTemplatePartialsProvider() []TemplatePartialsProvider
	GetThemeRenderers() []ThemeRenderer
	GetServiceLoggers() []ServiceLogger
	GetLocalesProviders() []LocalesProvider
	GetPrepareServePagesFeatures() []PrepareServePagesFeature
	GetFinalizeServePagesFeatures() []FinalizeServeRequestFeature
	GetPageContextFieldsProviders() []PageContextFieldsProvider
	GetPageContextParsersProviders() []PageContextParsersProvider
	GetPanicHandler() PanicHandler
	GetLocaleHandler() LocaleHandler
	GetServiceListener() ServiceListener
	GetRoutePagesHandler() RoutePagesHandler
	GetServePagesHandler() ServePagesHandler
	GetServiceLogHandler() ServiceLogHandler
}

type CanSetupInternals interface {
	Setup(enjin Internals)
}

type HotReloadableFeature interface {
	Feature
	HotReload() (err error)
}
