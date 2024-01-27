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

package site_environ

import (
	"os"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/go-corelibs/slices"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/maps"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

type MakeFeature[M interface{}] interface {
	SetSiteEnviron(key, name, value string) M
}

type CSiteEnviron[M interface{}] struct {
	siteEnvironData map[string]map[string]string
	siteEnvironHelp map[string]string

	this interface{}
}

func New[M interface{}](this interface{}, keysHelp ...string) (c *CSiteEnviron[M]) {
	c = &CSiteEnviron[M]{}
	c.InitSiteEnviron(this, keysHelp...)
	return
}

func (c *CSiteEnviron[M]) SiteEnvironUsageNotes() (notes []string) {
	f, _ := c.this.(feature.Feature)
	tag := f.Tag().String()
	notes = []string{
		"this feature supports composite environment variables",
	}
	var ok bool
	for _, key := range maps.SortedKeys(c.siteEnvironData) {
		snake := strcase.ToScreamingSnake(key)
		var help string
		if help, ok = c.siteEnvironHelp[key]; !ok {
			help = "value"
		}
		notes = append(notes, "# export "+globals.MakeFlagEnvKey(tag, snake)+"_SOME_THING='<"+help+">'")
	}
	return
}

func (c *CSiteEnviron[M]) InitSiteEnviron(this interface{}, keysHelp ...string) {
	c.this = this
	c.siteEnvironData = make(map[string]map[string]string)
	c.siteEnvironHelp = make(map[string]string)
	if count := len(keysHelp); count%2 != 0 {
		panic("unbalanced list of keysHelp strings")
	} else {
		for i := 0; i < count; i += 2 {
			key := strcase.ToKebab(keysHelp[i])
			help := beStrings.TrimQuotes(keysHelp[i+1])
			maps.MakeTypedKey(key, c.siteEnvironData)
			c.siteEnvironHelp[key] = help
		}
	}
}

func (c *CSiteEnviron[M]) SetSiteEnviron(key, name, value string) M {
	key = strcase.ToKebab(key)
	name = strcase.ToKebab(name)
	maps.MakeTypedKey(key, c.siteEnvironData)
	c.siteEnvironData[key][name] = value
	t, _ := c.this.(M)
	return t
}

func (c *CSiteEnviron[M]) StartupSiteEnviron() (err error) {
	environ := os.Environ()
	sort.Strings(environ)
	keys := maps.SortedKeys(c.siteEnvironData)
	envKeys := slices.ToScreamingSnakes(keys)
	f, _ := c.this.(feature.Feature)
	prefix := globals.MakeFlagEnvKey(f.Tag().String(), "")

	for _, env := range environ {
		if !strings.HasPrefix(env, prefix) {
			continue
		}
		if key, value, match := strings.Cut(env, "="); match {
			for idx, envKey := range envKeys {
				envPrefix := prefix + envKey + "_"
				if strings.HasPrefix(key, envPrefix) {
					name := strings.TrimPrefix(key, envPrefix)
					name = strcase.ToKebab(name)
					value = beStrings.TrimQuotes(value)
					maps.MakeTypedKey(keys[idx], c.siteEnvironData)
					c.siteEnvironData[keys[idx]][name] = value
				}
			}
		}
	}

	return
}

func (c *CSiteEnviron[M]) GetSiteEnviron(key string) (value map[string]string, ok bool) {
	value, ok = c.siteEnvironData[key]
	return
}

func (c *CSiteEnviron[M]) GetSiteEnvironNamed(key, name string) (value string, ok bool) {
	value, ok = c.siteEnvironData[key][name]
	return
}
