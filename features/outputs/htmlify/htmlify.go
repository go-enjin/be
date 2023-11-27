//go:build output_htmlify || outputs || htmlify || all

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

package htmlify

import (
	"net/http"
	"regexp"

	"github.com/urfave/cli/v2"
	"github.com/yosssi/gohtml"
	"golang.org/x/net/html"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/slices"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

const Tag feature.Tag = "outputs-htmlify"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	feature.OutputTransformer
}

type MakeFeature interface {
	Make() Feature

	Ignore(paths ...string) MakeFeature
}

type CFeature struct {
	feature.CFeature

	ignore  []string
	ignored []*regexp.Regexp
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
}

func (f *CFeature) Ignore(pathsOrPatterns ...string) MakeFeature {
	for _, pathOrPattern := range pathsOrPatterns {
		if !slices.Present(pathOrPattern, f.ignore...) {
			f.ignore = append(f.ignore, pathOrPattern)
		}
	}
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	for _, path := range f.ignore {
		if rx, e := regexp.Compile(path); e != nil {
			f.ignored = append(f.ignored, nil)
		} else {
			f.ignored = append(f.ignored, rx)
		}
	}
	return
}

func (f *CFeature) CanTransform(mime string, r *http.Request) (ok bool) {
	urlPath := bePath.TrimSlash(forms.TrimQueryParams(r.URL.Path))
	for idx, rx := range f.ignored {
		ignore := false
		if rx != nil {
			ignore = rx.MatchString(urlPath)
		} else {
			ignore = urlPath == f.ignore[idx]
		}
		if ignore {
			log.TraceRF(r, "htmlify ignoring (path or pattern): (%v) - %v", f.ignore[idx], urlPath)
			return
		}
	}
	basicMime := beStrings.GetBasicMime(mime)
	switch basicMime {
	case "text/html":
		ok = true
		log.TraceRF(r, "htmlify transforming: %v", urlPath)
	default:
		log.TraceRF(r, "htmlify ignoring (mime type): (%v) - %v", basicMime, urlPath)
	}
	return
}

func (f *CFeature) TransformOutput(_ string, input []byte) (output []byte) {
	gohtml.Condense = true
	gohtml.InlineTagMaxLength = 0
	gohtml.LineWrapMaxSpillover = 0
	gohtml.InlineTags = map[string]bool{
		"a":      true,
		"b":      true,
		"i":      true,
		"q":      true,
		"s":      true,
		"u":      true,
		"em":     true,
		"img":    true,
		"var":    true,
		"sub":    true,
		"sup":    true,
		"abbr":   true,
		"cite":   true,
		"code":   true,
		"span":   true,
		"input":  true,
		"small":  true,
		"strong": true,
	}
	gohtml.IsPreformatted = func(token html.Token) bool {
		switch token.Data {
		case "pre", "textarea", "code", "p":
			return true
		}
		return false
	}
	output = []byte(gohtml.Format(string(input)))
	return
}
