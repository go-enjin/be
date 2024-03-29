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

	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/slices"
	"github.com/go-corelibs/values"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/headers"
)

func (eb *EnjinBuilder) Set(key string, value interface{}) feature.Builder {
	eb.context[key] = value
	return eb
}

func (eb *EnjinBuilder) SetAlwaysHtmlRedirect(always bool) feature.Builder {
	eb.alwaysHtmlRedirect = always
	return eb
}

func (eb *EnjinBuilder) SetHtmlRedirectDelay(seconds int) feature.Builder {
	eb.htmlRedirectDelay = seconds
	return eb
}

func (eb *EnjinBuilder) AddHtmlHeadTag(name string, attr map[string]string) feature.Builder {
	if len(attr) <= 1 {
		log.FatalDF(1, "AddHtmlHeadTag requires at least two attribute keys")
	}
	eb.htmlHeadTags = append(eb.htmlHeadTags, htmlHeadTag{
		name: name,
		attr: attr,
	})
	return eb
}

func (eb *EnjinBuilder) AddDomains(domains ...string) feature.Builder {
	for _, domain := range domains {
		if !slices.Present(domain, eb.domains...) {
			eb.domains = append(eb.domains, domain)
		}
	}
	return eb
}

func (eb *EnjinBuilder) AddPreset(presets ...feature.Preset) feature.Builder {
	// reverse order the presets in order to end up with the dev's specified order of features because the standard
	// preset inclusion policy is to prepend instead of append, this way the first preset added results in those
	// features to be at the start of the features list
	for idx := len(presets) - 1; idx >= 0; idx-- {
		preset := presets[idx]
		eb.presets = append(eb.presets, preset)
		log.DebugDF(1, "including %v preset features...", preset.Label())
		if err := preset.Preset(eb); err != nil {
			log.FatalDF(1, "preset [%v] - %v", preset.Label(), err)
		}
	}
	return eb
}

func checkRegisterFeature[T interface{}](f feature.Feature, list []T) []T {
	var check *T
	if ff, ok := feature.AsTyped[T](f); ok {
		log.DebugDF(2, "registering %v as a %v", f.Tag(), values.TypeOf(check)[1:])
		list = append(list, ff)
	}
	return list
}

func checkRegisterSingleFeature[T interface{}](f feature.Feature, existing T) (typed T) {
	if ff, ok := feature.AsTyped[T](f); ok {
		label := values.TypeOf((*T)(nil))[1:]
		if !values.IsEmpty(existing) {
			log.FatalDF(2, "only one %s feature allowed", label)
		}
		log.DebugDF(2, "registering %v as a %v", f.Tag(), label)
		typed = ff
	} else {
		typed = existing
	}
	return
}

func (eb *EnjinBuilder) includeFeature(f feature.Feature) {
	eb.fFormatProviders = checkRegisterFeature(f, eb.fFormatProviders)
	eb.fRequestFilters = checkRegisterFeature(f, eb.fRequestFilters)
	eb.fPageContextModifiers = checkRegisterFeature(f, eb.fPageContextModifiers)
	eb.fPageContextUpdaters = checkRegisterFeature(f, eb.fPageContextUpdaters)
	eb.fPageRestrictionHandlers = checkRegisterFeature(f, eb.fPageRestrictionHandlers)
	eb.fMenuProviders = checkRegisterFeature(f, eb.fMenuProviders)
	eb.fDataRestrictionHandlers = checkRegisterFeature(f, eb.fDataRestrictionHandlers)
	eb.fOutputTranslators = checkRegisterFeature(f, eb.fOutputTranslators)
	eb.fOutputTransformers = checkRegisterFeature(f, eb.fOutputTransformers)
	eb.fPageTypeProcessors = checkRegisterFeature(f, eb.fPageTypeProcessors)
	eb.fServePathFeatures = checkRegisterFeature(f, eb.fServePathFeatures)
	eb.fDatabases = checkRegisterFeature(f, eb.fDatabases)
	eb.fEmailSenders = checkRegisterFeature(f, eb.fEmailSenders)
	eb.fRequestModifiers = checkRegisterFeature(f, eb.fRequestModifiers)
	eb.fRequestRewriters = checkRegisterFeature(f, eb.fRequestRewriters)
	eb.fPermissionsPolicyModifiers = checkRegisterFeature(f, eb.fPermissionsPolicyModifiers)
	eb.fContentSecurityPolicyModifiers = checkRegisterFeature(f, eb.fContentSecurityPolicyModifiers)
	eb.fUseMiddlewares = checkRegisterFeature(f, eb.fUseMiddlewares)
	eb.fHeadersModifiers = checkRegisterFeature(f, eb.fHeadersModifiers)
	eb.fProcessors = checkRegisterFeature(f, eb.fProcessors)
	eb.fApplyMiddlewares = checkRegisterFeature(f, eb.fApplyMiddlewares)
	eb.fPageProviders = checkRegisterFeature(f, eb.fPageProviders)
	eb.fFileProviders = checkRegisterFeature(f, eb.fFileProviders)
	eb.fQueryIndexFeatures = checkRegisterFeature(f, eb.fQueryIndexFeatures)
	eb.fPageContextProviders = checkRegisterFeature(f, eb.fPageContextProviders)
	eb.fAuthProviders = checkRegisterFeature(f, eb.fAuthProviders)
	eb.fUserActionsProviders = checkRegisterFeature(f, eb.fUserActionsProviders)
	eb.fEnjinContextProvider = checkRegisterFeature(f, eb.fEnjinContextProvider)
	eb.fPageShortcodeProcessors = checkRegisterFeature(f, eb.fPageShortcodeProcessors)
	eb.fFuncMapProviders = checkRegisterFeature(f, eb.fFuncMapProviders)
	eb.fTemplatePartialsProvider = checkRegisterFeature(f, eb.fTemplatePartialsProvider)
	eb.fThemeRenderers = checkRegisterFeature(f, eb.fThemeRenderers)
	eb.fServiceLoggers = checkRegisterFeature(f, eb.fServiceLoggers)
	eb.fLocalesProviders = checkRegisterFeature(f, eb.fLocalesProviders)
	eb.fPrepareServePagesFeatures = checkRegisterFeature(f, eb.fPrepareServePagesFeatures)
	eb.fFinalizeServePagesFeatures = checkRegisterFeature(f, eb.fFinalizeServePagesFeatures)
	eb.fPageContextFieldsProviders = checkRegisterFeature(f, eb.fPageContextFieldsProviders)
	eb.fPageContextParsersProviders = checkRegisterFeature(f, eb.fPageContextParsersProviders)

	eb.fNonceFactory = checkRegisterSingleFeature(f, eb.fNonceFactory)
	eb.fTokenFactory = checkRegisterSingleFeature(f, eb.fTokenFactory)
	eb.fSyncLockerFactory = checkRegisterSingleFeature(f, eb.fSyncLockerFactory)

	eb.fPanicHandler = checkRegisterSingleFeature(f, eb.fPanicHandler)
	eb.fLocaleHandler = checkRegisterSingleFeature(f, eb.fLocaleHandler)
	eb.fServiceListener = checkRegisterSingleFeature(f, eb.fServiceListener)
	eb.fRoutePagesHandler = checkRegisterSingleFeature(f, eb.fRoutePagesHandler)
	eb.fServePagesHandler = checkRegisterSingleFeature(f, eb.fServePagesHandler)
	eb.fServiceLogHandler = checkRegisterSingleFeature(f, eb.fServiceLogHandler)
}

func (eb *EnjinBuilder) PrependFeature(f feature.Feature) feature.Builder {
	if f == nil {
		return eb
	}
	log.DebugF("prepending feature: %v", f.Tag())
	if err := eb.features.Prepend(f); err != nil {
		log.FatalDF(1, "error prepending feature: %T - %v", f, err)
	}
	eb.includeFeature(f)
	return eb
}

func (eb *EnjinBuilder) AddFeature(features ...feature.Feature) feature.Builder {
	if len(features) == 0 {
		return eb
	}
	for _, f := range features {
		if f == nil {
			continue
		}
		log.DebugF("adding feature: %v", f.Tag())
		if err := eb.features.Add(f); err != nil {
			log.FatalDF(1, "error adding feature: %T - %v", f, err)
		}
		eb.includeFeature(f)
	}
	return eb
}

func (eb *EnjinBuilder) AddFlags(flags ...cli.Flag) feature.Builder {
	for _, flag := range flags {
		fNames := flag.Names()
		var found bool
		for _, present := range eb.flags {
			pNames := present.Names()
			if fNames[0] == pNames[0] {
				found = true
			}
		}
		if found {
			log.DebugDF(1, "skipping existing flag: %v", fNames[0])
		} else {
			eb.flags = append(eb.flags, flag)
		}
	}
	return eb
}

func (eb *EnjinBuilder) AddCommands(commands ...*cli.Command) feature.Builder {
	eb.commands = append(
		eb.commands,
		commands...,
	)
	return eb
}

func (eb *EnjinBuilder) AddModifyHeadersFn(fn headers.ModifyHeadersFn) feature.Builder {
	if fn != nil {
		eb.headers = append(eb.headers, fn)
	}
	return eb
}

func (eb *EnjinBuilder) AddCspModifierFn(tag string, fn feature.CspModifierFn) feature.Builder {
	eb.cspModifierFnOrder = append(eb.cspModifierFnOrder, tag)
	eb.cspModifierFns[tag] = fn
	return eb
}

func (eb *EnjinBuilder) AddNotifyHook(name string, hook feature.NotifyHook) feature.Builder {
	_notifyHooks[name] = hook
	return eb
}

func (eb *EnjinBuilder) AddRouteProcessor(route string, processor feature.ReqProcessFn) feature.Builder {
	if _, ok := eb.processors[route]; ok {
		log.FatalF("%v route processor already exists", route)
	}
	eb.processors[route] = processor
	return eb
}

func (eb *EnjinBuilder) resolveFeatureDeps() (err error) {
	found := make(feature.Tags, 0)
	for _, f := range eb.features.List() {
		if deps := f.Depends(); len(deps) > 0 {
			for _, d := range deps {
				if !found.Has(d) {
					err = fmt.Errorf("%v is missing dependency: %v", f.Tag(), d)
					return
				}
			}
		}
		found = append(found, f.Tag())
	}
	for _, c := range eb.consoles {
		if deps := c.Depends(); len(deps) > 0 {
			for _, d := range deps {
				if !found.Has(d) {
					err = fmt.Errorf("%v is missing dependency: %v", c.Tag(), d)
					return
				}
			}
		}
		found = append(found, c.Tag())
	}
	return
}
