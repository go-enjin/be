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

	"github.com/urfave/cli/v2"
	"github.com/yosssi/gohtml"

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

	ignored []string
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
func (f *Feature) Ignore(paths ...string) MakeFeature {
	for _, path := range paths {
		path = bePath.TrimSlash(net.TrimQueryParams(path))
		if !beStrings.StringInStrings(path, f.ignored...) {
			f.ignored = append(f.ignored, path)
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
	return
}

func (f *Feature) CanTransform(mime string, r *http.Request) (ok bool) {
	urlPath := bePath.TrimSlash(net.TrimQueryParams(r.URL.Path))
	for _, path := range f.ignored {
		if urlPath == path {
			log.DebugF("htmlify ignoring: %v", path)
			return
		}
	}
	switch beStrings.GetBasicMime(mime) {
	case "text/html":
		ok = true
	}
	return
}

func (f *Feature) TransformOutput(_ string, input []byte) (output []byte) {
	output = []byte(gohtml.Format(string(input)))
	return
}