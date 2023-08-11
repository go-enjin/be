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

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/indexing"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/headers"
	"github.com/go-enjin/be/pkg/slices"
	types "github.com/go-enjin/be/pkg/types/theme-types"
	"github.com/go-enjin/be/pkg/userbase"
)

func (eb *EnjinBuilder) Set(key string, value interface{}) feature.Builder {
	eb.context[key] = value
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

func checkRegisterFeature[T interface{}](f feature.Feature, list []T) []T {
	var check *T
	if ff, ok := feature.AsTyped[T](f); ok {
		log.DebugDF(1, "registering %v as a %v", f.Tag(), fmt.Sprintf("%T", check)[1:])
		list = append(list, ff)
	}
	return list
}

func (eb *EnjinBuilder) AddFeature(f feature.Feature) feature.Builder {
	if f == nil {
		return eb
	}
	if err := eb.features.Add(f); err != nil {
		log.FatalDF(1, "error adding feature: %T - %v", f, err)
	}
	log.DebugF("adding feature: %v", f.Tag())

	eb.fFormatProviders = checkRegisterFeature[types.FormatProvider](f, eb.fFormatProviders)
	eb.fRequestFilters = checkRegisterFeature[feature.RequestFilter](f, eb.fRequestFilters)
	eb.fPageContextModifiers = checkRegisterFeature[feature.PageContextModifier](f, eb.fPageContextModifiers)
	eb.fPageRestrictionHandlers = checkRegisterFeature[feature.PageRestrictionHandler](f, eb.fPageRestrictionHandlers)
	eb.fMenuProviders = checkRegisterFeature[feature.MenuProvider](f, eb.fMenuProviders)
	eb.fDataRestrictionHandlers = checkRegisterFeature[feature.DataRestrictionHandler](f, eb.fDataRestrictionHandlers)
	eb.fOutputTranslators = checkRegisterFeature[feature.OutputTranslator](f, eb.fOutputTranslators)
	eb.fOutputTransformers = checkRegisterFeature[feature.OutputTransformer](f, eb.fOutputTransformers)
	eb.fPageTypeProcessors = checkRegisterFeature[feature.PageTypeProcessor](f, eb.fPageTypeProcessors)
	eb.fServePathFeatures = checkRegisterFeature[feature.ServePathFeature](f, eb.fServePathFeatures)
	eb.fDatabases = checkRegisterFeature[feature.Database](f, eb.fDatabases)
	eb.fEmailSenders = checkRegisterFeature[feature.EmailSender](f, eb.fEmailSenders)
	eb.fRequestModifiers = checkRegisterFeature[feature.RequestModifier](f, eb.fRequestModifiers)
	eb.fRequestRewriters = checkRegisterFeature[feature.RequestRewriter](f, eb.fRequestRewriters)
	eb.fPermissionsPolicyModifiers = checkRegisterFeature[feature.PermissionsPolicyModifier](f, eb.fPermissionsPolicyModifiers)
	eb.fContentSecurityPolicyModifiers = checkRegisterFeature[feature.ContentSecurityPolicyModifier](f, eb.fContentSecurityPolicyModifiers)
	eb.fUseMiddlewares = checkRegisterFeature[feature.UseMiddleware](f, eb.fUseMiddlewares)
	eb.fHeadersModifiers = checkRegisterFeature[feature.HeadersModifier](f, eb.fHeadersModifiers)
	eb.fProcessors = checkRegisterFeature[feature.Processor](f, eb.fProcessors)
	eb.fApplyMiddlewares = checkRegisterFeature[feature.ApplyMiddleware](f, eb.fApplyMiddlewares)
	eb.fPageProviders = checkRegisterFeature[feature.PageProvider](f, eb.fPageProviders)
	eb.fFileProviders = checkRegisterFeature[feature.FileProvider](f, eb.fFileProviders)
	eb.fQueryIndexFeatures = checkRegisterFeature[indexing.QueryIndexFeature](f, eb.fQueryIndexFeatures)
	eb.fPageContextProviders = checkRegisterFeature[indexing.PageContextProvider](f, eb.fPageContextProviders)
	eb.fAuthProviders = checkRegisterFeature[userbase.AuthProvider](f, eb.fAuthProviders)
	eb.fUserActionsProviders = checkRegisterFeature[userbase.UserActionsProvider](f, eb.fUserActionsProviders)
	eb.fEnjinContextProvider = checkRegisterFeature[feature.EnjinContextProvider](f, eb.fEnjinContextProvider)
	eb.fPageShortcodeProcessors = checkRegisterFeature[feature.PageShortcodeProcessor](f, eb.fPageShortcodeProcessors)

	return eb
}

func (eb *EnjinBuilder) AddFeatureNotes(tag feature.Tag, notes ...string) feature.Builder {
	eb.notes[tag] = append(eb.notes[tag], notes...)
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

func (eb *EnjinBuilder) MakeEnvKey(name string) (key string) {
	key = globals.MakeEnvKey(name)
	return
}

func (eb *EnjinBuilder) MakeEnvKeys(names ...string) (keys []string) {
	keys = globals.MakeEnvKeys(names...)
	return
}

func (eb *EnjinBuilder) AddModifyHeadersFn(fn headers.ModifyHeadersFn) feature.Builder {
	if fn != nil {
		eb.headers = append(eb.headers, fn)
	}
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