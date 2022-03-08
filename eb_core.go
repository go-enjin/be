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

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/headers"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func (eb *EnjinBuilder) Set(key string, value interface{}) feature.Builder {
	eb.context[key] = value
	return eb
}

func (eb *EnjinBuilder) AddDomains(domains ...string) feature.Builder {
	for _, domain := range domains {
		if !beStrings.StringInStrings(domain, eb.domains...) {
			eb.domains = append(eb.domains, domain)
		}
	}
	return eb
}

func (eb *EnjinBuilder) AddFeature(f feature.Feature) feature.Builder {
	for _, known := range eb.features {
		if known.Tag() == f.Tag() {
			return eb
		}
	}
	log.DebugF("adding feature: %v", f.Tag())
	eb.features = append(eb.features, f)
	return eb
}

func (eb *EnjinBuilder) AddFlags(flags ...cli.Flag) feature.Builder {
	eb.flags = append(
		eb.flags,
		flags...,
	)
	return eb
}

func (eb *EnjinBuilder) MakeEnvKey(name string) (key string) {
	key = name
	if globals.EnvPrefix != "" {
		key = globals.EnvPrefix + "_" + name
	}
	key = strcase.ToScreamingSnake(key)
	return
}

func (eb *EnjinBuilder) MakeEnvKeys(names ...string) (keys []string) {
	for _, name := range names {
		keys = append(keys, eb.MakeEnvKey(name))
	}
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
	for _, f := range eb.features {
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
	return
}