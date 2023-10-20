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

package casting

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/values"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-funcmaps-casting"

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
		"asURL":       AsURL,
		"asHTML":      AsHTML,
		"asHTMLAttr":  AsHTMLAttr,
		"asCSS":       AsCSS,
		"asJS":        AsJS,
		"safeHTML":    AsHTML,
		"typeOf":      values.TypeOf,
		"typeOfSlice": TypeOfSlice,
	}
	return
}

func AsURL(input interface{}) template.URL {
	switch v := input.(type) {
	case string:
		return template.URL(v)
	case template.URL:
		return v
	default:
		return template.URL(fmt.Sprintf("%v", v))
	}
}

func AsJS(input interface{}) template.JS {
	switch v := input.(type) {
	case string:
		return template.JS(v)
	case template.JS:
		return v
	default:
		return template.JS(fmt.Sprintf("%v", v))
	}
}

func AsCSS(input interface{}) template.CSS {
	switch v := input.(type) {
	case string:
		return template.CSS(v)
	case template.CSS:
		return v
	default:
		return template.CSS(fmt.Sprintf("%v", v))
	}
}

func AsHTML(input interface{}) template.HTML {
	switch v := input.(type) {
	case string:
		return template.HTML(v)
	case template.HTML:
		return v
	default:
		return template.HTML(fmt.Sprintf("%v", v))
	}
}

func AsHTMLAttr(input interface{}) template.HTMLAttr {
	switch v := input.(type) {
	case string:
		return template.HTMLAttr(v)
	case template.HTML:
		return template.HTMLAttr(v)
	default:
		return template.HTMLAttr(fmt.Sprintf("%v", v))
	}
}

func TypeOfSlice(input interface{}) (yes bool) {
	yes = strings.Contains(values.TypeOf(input), "[]")
	return
}