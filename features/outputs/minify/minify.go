//go:build output_minify || outputs || minify || all

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
	"regexp"

	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/slices"
	clStrings "github.com/go-corelibs/strings"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
	beMinify "github.com/go-enjin/be/pkg/net/minify"
	bePath "github.com/go-enjin/be/pkg/path"
)

const Tag feature.Tag = "outputs-minify"

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

	AddMimeType(mime string) MakeFeature
	SetMimeTypes(mimeTypes ...string) MakeFeature
	Ignore(paths ...string) MakeFeature
}

type CFeature struct {
	feature.CFeature

	mimeTypes []string
	ignore    []string
	ignored   []*regexp.Regexp
}

var _minifyInstance = beMinify.NewInstance()

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
}

func (f *CFeature) AddMimeType(mime string) MakeFeature {
	if !slices.Present(mime, f.mimeTypes...) {
		f.mimeTypes = append(f.mimeTypes, mime)
	}
	return f
}

func (f *CFeature) SetMimeTypes(mimeTypes ...string) MakeFeature {
	f.mimeTypes = mimeTypes
	return f
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
	if len(f.mimeTypes) == 0 {
		f.mimeTypes = []string{
			"text/css",
			"text/javascript",
		}
	}
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
			log.TraceRF(r, "minify ignoring (path or pattern): (%v) - %v", f.ignore[idx], urlPath)
			return
		}
	}
	if len(f.mimeTypes) > 0 {
		basicMime := clStrings.GetBasicMime(mime)
		for _, supported := range f.mimeTypes {
			switch supported {
			case mime, basicMime:
				ok = true
				log.TraceRF(r, "minify transforming: %v", urlPath)
				return
			}
		}
		log.TraceRF(r, "minify ignoring (mime type): (%v) - %v", basicMime, urlPath)
	}
	return
}

func (f *CFeature) TransformOutput(mime string, input []byte) (output []byte) {
	var err error
	if output, err = _minifyInstance.Bytes(mime, input); err == nil {
		return
	}
	log.ErrorF("error minifying %v: %v", clStrings.GetBasicMime(mime), err)
	output = input
	return
}
