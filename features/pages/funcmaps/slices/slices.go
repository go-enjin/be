//go:build page_funcmaps || pages || all

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

package slices

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/maruel/natural"
	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/maths"
	"github.com/go-enjin/be/pkg/slices"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-funcmaps-slices"

type Feature interface {
	feature.Feature
	feature.FuncMapProvider
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.CFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *CFeature) Shutdown() {

}

func (f *CFeature) MakeFuncMap(ctx beContext.Context) (fm feature.FuncMap) {
	fm = feature.FuncMap{
		"joinStrings":                strings.Join,
		"splitString":                strings.Split,
		"filterStrings":              FilterStrings,
		"stringsAsList":              StringsAsList,
		"reverseStrings":             ReverseStrings,
		"sortedStrings":              SortedStrings,
		"sortedKeys":                 SortedKeys,
		"sortedFirstLetters":         SortedFirstLetters,
		"sortedLastNameFirstLetters": SortedLastNameFirstLetters,
		"iterate":                    Iterate,
		"makeSlice":                  MakeSlice[interface{}],
		"makeStringSlice":            MakeStringSlice,
		"appendSlice":                AppendSlice[interface{}],
		"appendStringSlice":          AppendStringSlice,
		"stringInStrings":            slices.Present[string],
		"withinStrings":              slices.Within[string, []string],
		"anyWithinStrings":           slices.AnyWithin[string, []string],
	}
	return
}

func MakeSlice[T interface{}](values ...T) (output []T) {
	output = append(output, values...)
	return
}

func MakeStringSlice(values ...interface{}) (output []string) {
	for _, value := range values {
		output = append(output, fmt.Sprintf("%v", value))
	}
	return
}

func AppendSlice[T interface{}](slice []T, values ...T) (combined []T) {
	combined = append(slice, values...)
	return
}

func AppendStringSlice(slice []string, values ...interface{}) (output []string) {
	output = append(output, slice...)
	for _, value := range values {
		output = append(output, fmt.Sprintf("%v", value))
	}
	return
}

func SortedKeys(v interface{}) (keys []string) {
	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Map:
		if kt := t.Key(); kt.Kind() == reflect.String {
			value := reflect.ValueOf(v)
			mapKeys := value.MapKeys()
			for _, k := range mapKeys {
				keys = append(keys, k.String())
			}
			sort.Sort(natural.StringSlice(keys))
			return
		}
		log.WarnF("unsupported sortedKeys map key type: %T", v)
	default:
		log.WarnF("unsupported sortedKeys type: %T", v)
	}
	return
}

func FilterStrings(pattern string, values []string) (filtered []string) {
	if rx, err := regexp.Compile(pattern); err == nil {
		for _, value := range values {
			if rx.MatchString(value) {
				filtered = append(filtered, value)
			}
		}
	} else {
		log.ErrorF("error compiling filterStrings pattern: m!%v! - %v", pattern, err)
	}
	return
}

func StringsAsList(v ...string) (list []string) {
	list = v
	return
}

func ReverseStrings(v []string) (reversed []string) {
	for i := len(v) - 1; i >= 0; i-- {
		reversed = append(reversed, v[i])
	}
	return
}

func SortedStrings(values []string) (sorted []string) {
	sorted = values[:]
	sort.Sort(natural.StringSlice(sorted))
	return
}

func SortedFirstLetters(values []interface{}) (firsts []string) {
	cache := make(map[string]bool)
	for _, v := range values {
		if value, ok := v.(string); ok && value != "" {
			char := strings.ToLower(string(value[0]))
			cache[char] = true
		}
	}
	firsts = maps.SortedKeys(cache)
	return
}

func SortedLastNameFirstLetters(values []interface{}) (firsts []string) {
	cache := make(map[string]bool)
	for _, v := range values {
		if value, ok := v.(string); ok && value != "" {
			if word := beStrings.LastName(value); word != "" {
				char := strings.ToLower(string(word[0]))
				cache[char] = true
			}
		}
	}
	firsts = maps.SortedKeys(cache)
	return
}

func Iterate(argv ...interface{}) (ch chan int) {
	var argvInt []int
	var from, inc, to int

	for _, arg := range argv {
		if v := maths.ToInt(arg, math.MinInt); v > math.MinInt {
			argvInt = append(argvInt, v)
		} else {
			panic(fmt.Sprintf("expected any number type, received: %q (%T)", arg, arg))
		}
	}

	ch = make(chan int)
	switch len(argvInt) {
	case 0:
		log.ErrorF("template is trying to iterate over nothing")
	case 1:
		from, inc, to = 0, 1, argvInt[0]
	case 2:
		from, inc, to = argvInt[0], 1, argvInt[1]
	default:
		from, inc, to = argvInt[0], argvInt[1], argvInt[2]
	}
	go func() {
		for i := from; i < to; i += inc {
			ch <- i
		}
		close(ch)
	}()
	return
}
