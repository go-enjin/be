//go:build htmlify || outputs || all

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
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	bePath "github.com/go-enjin/be/pkg/path"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var _htmlify *Feature

var _ MakeFeature = (*Feature)(nil)

var _ feature.Feature = (*Feature)(nil)

var _ feature.OutputTransformer = (*Feature)(nil)

const Tag feature.Tag = "OutputHtmlify"

type Feature struct {
	feature.CFeature

	ignore  []string
	ignored []*regexp.Regexp
}

type MakeFeature interface {
	feature.MakeFeature

	Ignore(paths ...string) MakeFeature
}

func New() MakeFeature {
	if _htmlify == nil {
		_htmlify = new(Feature)
		_htmlify.Init(_htmlify)
	}
	return _htmlify
}

func (f *Feature) Ignore(pathsOrPatterns ...string) MakeFeature {
	for _, pathOrPattern := range pathsOrPatterns {
		if !beStrings.StringInStrings(pathOrPattern, f.ignore...) {
			f.ignore = append(f.ignore, pathOrPattern)
		}
	}
	return f
}

func (f *Feature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *Feature) Build(b feature.Buildable) (err error) {
	return
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	for _, path := range f.ignore {
		if rx, e := regexp.Compile(path); e != nil {
			f.ignored = append(f.ignored, nil)
		} else {
			f.ignored = append(f.ignored, rx)
		}
	}
	return
}

func (f *Feature) CanTransform(mime string, r *http.Request) (ok bool) {
	urlPath := bePath.TrimSlash(net.TrimQueryParams(r.URL.Path))
	for idx, rx := range f.ignored {
		ignore := false
		if rx != nil {
			ignore = rx.MatchString(urlPath)
		} else {
			ignore = urlPath == f.ignore[idx]
		}
		if ignore {
			log.TraceF("htmlify ignoring (path or pattern): (%v) - %v", f.ignore[idx], urlPath)
			return
		}
	}
	basicMime := beStrings.GetBasicMime(mime)
	switch basicMime {
	case "text/html":
		ok = true
		log.TraceF("htmlify transforming: %v", urlPath)
	default:
		log.TraceF("htmlify ignoring (mime type): (%v) - %v", basicMime, urlPath)
	}
	return
}

func (f *Feature) TransformOutput(_ string, input []byte) (output []byte) {
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
		"code":   true,
		"span":   true,
		"input":  true,
		"small":  true,
		"strong": true,
	}
	gohtml.IsPreformatted = func(token html.Token) bool {
		return token.Data == "pre" || token.Data == "textarea" || token.Data == "code" || token.Data == "p"
	}
	output = []byte(gohtml.Format(string(input)))
	return
}