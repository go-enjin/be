//go:build minify || outputs || all

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

package minify

import (
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net"
	beMinify "github.com/go-enjin/be/pkg/net/minify"
	bePath "github.com/go-enjin/be/pkg/path"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var _minify *Feature

var _minifyInstance = beMinify.NewInstance()

var _ MakeFeature = (*Feature)(nil)

var _ feature.Feature = (*Feature)(nil)

var _ feature.OutputTransformer = (*Feature)(nil)

const Tag feature.Tag = "OutputMinify"

type Feature struct {
	feature.CFeature

	mimeTypes []string
	ignored   []string
}

type MakeFeature interface {
	feature.MakeFeature

	AddMimeType(mime string) MakeFeature
	SetMimeTypes(mimeTypes ...string) MakeFeature
	Ignore(paths ...string) MakeFeature
}

func New() MakeFeature {
	if _minify == nil {
		_minify = new(Feature)
		_minify.Init(_minify)
	}
	return _minify
}

func (f *Feature) AddMimeType(mime string) MakeFeature {
	if !beStrings.StringInStrings(mime, f.mimeTypes...) {
		f.mimeTypes = append(f.mimeTypes, mime)
	}
	return f
}

func (f *Feature) SetMimeTypes(mimeTypes ...string) MakeFeature {
	f.mimeTypes = mimeTypes
	return f
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
	if len(f.mimeTypes) == 0 {
		f.mimeTypes = []string{
			"text/css",
			"text/javascript",
		}
	}
	return
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *Feature) CanTransform(mime string, r *http.Request) (ok bool) {
	urlPath := bePath.TrimSlash(net.TrimQueryParams(r.URL.Path))
	for _, path := range f.ignored {
		if urlPath == path {
			log.DebugF("minify ignoring: %v", path)
			return
		}
	}
	basicMime := beStrings.GetBasicMime(mime)
	for _, supported := range f.mimeTypes {
		switch supported {
		case mime, basicMime:
			ok = true
			return
		}
	}
	return
}

func (f *Feature) TransformOutput(mime string, input []byte) (output []byte) {
	var err error
	if output, err = _minifyInstance.Bytes(mime, input); err == nil {
		return
	}
	log.ErrorF("error minifying %v: %v", beStrings.GetBasicMime(mime), err)
	output = input
	return
}