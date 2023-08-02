//go:build page_shortcodes || pages || all

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

package shortcodes

import (
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/go-stdlib-text-scanner"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
)

const Tag feature.Tag = "page-shortcodes"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	feature.PageShortcodeProcessor
}

type CFeature struct {
	feature.CFeature

	known   map[string]Shortcode
	aliases map[string]string
}

type MakeFeature interface {
	Make() Feature

	Defaults() MakeFeature

	Add(shortcodes ...Shortcode) MakeFeature
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.known = make(map[string]Shortcode)
	f.aliases = make(map[string]string)
}

func (f *CFeature) Add(shortcodes ...Shortcode) MakeFeature {
	for _, shortcode := range shortcodes {
		f.known[shortcode.Name] = shortcode
		for _, alias := range shortcode.Aliases {
			if _, present := f.known[alias]; !present {
				f.aliases[alias] = shortcode.Name
			}
		}
	}
	return f
}

func (f *CFeature) Defaults() MakeFeature {
	f.Add(
		BoldShortcode,
		ItalicShortcode,
		UnderlineShortcode,
		ColorShortcode,
		StrikethroughShortcode,
		SuperscriptShortcode,
		SubscriptShortcode,
		UrlShortcode,
		CodeShortcode,
		QuoteShortcode,
	)
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)

}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) LookupShortcode(name string) (shortcode Shortcode, ok bool) {
	if shortcode, ok = f.known[name]; !ok {
		if actual, aliased := f.aliases[name]; aliased {
			shortcode, ok = f.known[actual]
		}
	}
	return
}

func (f *CFeature) TranslateShortcodes(content string, ctx beContext.Context) (modified string) {
	p := parser{
		feature: f,
		errHandler: func(s *scanner.Scanner, msg string) {
			// nop
		},
	}
	modified = p.process(content).Render(ctx)
	return
}