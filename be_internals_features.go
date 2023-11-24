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

package be

import (
	"github.com/go-enjin/be/pkg/feature"
)

func (e *Enjin) GetFormatProviders() []feature.PageFormatProvider {
	return e.eb.fFormatProviders
}

func (e *Enjin) GetRequestFilters() []feature.RequestFilter {
	return e.eb.fRequestFilters
}

func (e *Enjin) GetPageContextModifiers() []feature.PageContextModifier {
	return e.eb.fPageContextModifiers
}

func (e *Enjin) GetPageContextUpdaters() []feature.PageContextUpdater {
	return e.eb.fPageContextUpdaters
}

func (e *Enjin) GetPageRestrictionHandlers() []feature.PageRestrictionHandler {
	return e.eb.fPageRestrictionHandlers
}

func (e *Enjin) GetMenuProviders() []feature.MenuProvider {
	return e.eb.fMenuProviders
}

func (e *Enjin) GetDataRestrictionHandlers() []feature.DataRestrictionHandler {
	return e.eb.fDataRestrictionHandlers
}

func (e *Enjin) GetOutputTranslators() []feature.OutputTranslator {
	return e.eb.fOutputTranslators
}

func (e *Enjin) GetOutputTransformers() []feature.OutputTransformer {
	return e.eb.fOutputTransformers
}

func (e *Enjin) GetPageTypeProcessors() []feature.PageTypeProcessor {
	return e.eb.fPageTypeProcessors
}

func (e *Enjin) GetServePathFeatures() []feature.ServePathFeature {
	return e.eb.fServePathFeatures
}

func (e *Enjin) GetDatabases() []feature.Database {
	return e.eb.fDatabases
}

func (e *Enjin) GetEmailSenders() []feature.EmailSender {
	return e.eb.fEmailSenders
}

func (e *Enjin) GetRequestModifiers() []feature.RequestModifier {
	return e.eb.fRequestModifiers
}

func (e *Enjin) GetRequestRewriters() []feature.RequestRewriter {
	return e.eb.fRequestRewriters
}

func (e *Enjin) GetPermissionsPolicyModifiers() []feature.PermissionsPolicyModifier {
	return e.eb.fPermissionsPolicyModifiers
}

func (e *Enjin) GetContentSecurityPolicyModifiers() []feature.ContentSecurityPolicyModifier {
	return e.eb.fContentSecurityPolicyModifiers
}

func (e *Enjin) GetUseMiddlewares() []feature.UseMiddleware {
	return e.eb.fUseMiddlewares
}

func (e *Enjin) GetHeadersModifiers() []feature.HeadersModifier {
	return e.eb.fHeadersModifiers
}

func (e *Enjin) GetProcessors() []feature.Processor {
	return e.eb.fProcessors
}

func (e *Enjin) GetApplyMiddlewares() []feature.ApplyMiddleware {
	return e.eb.fApplyMiddlewares
}

func (e *Enjin) GetPageProviders() []feature.PageProvider {
	return e.eb.fPageProviders
}

func (e *Enjin) GetFileProviders() []feature.FileProvider {
	return e.eb.fFileProviders
}

func (e *Enjin) GetQueryIndexFeatures() []feature.QueryIndexFeature {
	return e.eb.fQueryIndexFeatures
}

func (e *Enjin) GetPageContextProviders() []feature.PageContextProvider {
	return e.eb.fPageContextProviders
}

func (e *Enjin) GetAuthProviders() []feature.AuthProvider {
	return e.eb.fAuthProviders
}

func (e *Enjin) GetUserActionsProviders() []feature.UserActionsProvider {
	return e.eb.fUserActionsProviders
}

func (e *Enjin) GetEnjinContextProvider() []feature.EnjinContextProvider {
	return e.eb.fEnjinContextProvider
}

func (e *Enjin) GetPageShortcodeProcessors() []feature.PageShortcodeProcessor {
	return e.eb.fPageShortcodeProcessors
}

func (e *Enjin) GetFuncMapProviders() []feature.FuncMapProvider {
	return e.eb.fFuncMapProviders
}

func (e *Enjin) GetTemplatePartialsProvider() []feature.TemplatePartialsProvider {
	return e.eb.fTemplatePartialsProvider
}

func (e *Enjin) GetThemeRenderers() []feature.ThemeRenderer {
	return e.eb.fThemeRenderers
}

func (e *Enjin) GetServiceLoggers() []feature.ServiceLogger {
	return e.eb.fServiceLoggers
}

func (e *Enjin) GetLocalesProviders() []feature.LocalesProvider {
	return e.eb.fLocalesProviders
}

func (e *Enjin) GetPrepareServePagesFeatures() []feature.PrepareServePagesFeature {
	return e.eb.fPrepareServePagesFeatures
}

func (e *Enjin) GetFinalizeServePagesFeatures() []feature.FinalizeServeRequestFeature {
	return e.eb.fFinalizeServePagesFeatures
}

func (e *Enjin) GetPageContextFieldsProviders() []feature.PageContextFieldsProvider {
	return e.eb.fPageContextFieldsProviders
}

func (e *Enjin) GetPageContextParsersProviders() []feature.PageContextParsersProvider {
	return e.eb.fPageContextParsersProviders
}

func (e *Enjin) GetPanicHandler() feature.PanicHandler {
	return e.eb.fPanicHandler
}

func (e *Enjin) GetLocaleHandler() feature.LocaleHandler {
	return e.eb.fLocaleHandler
}

func (e *Enjin) GetServiceListener() feature.ServiceListener {
	return e.eb.fServiceListener
}

func (e *Enjin) GetRoutePagesHandler() feature.RoutePagesHandler {
	return e.eb.fRoutePagesHandler
}

func (e *Enjin) GetServePagesHandler() feature.ServePagesHandler {
	return e.eb.fServePagesHandler
}

func (e *Enjin) GetServiceLogHandler() feature.ServiceLogHandler {
	return e.eb.fServiceLogHandler
}
