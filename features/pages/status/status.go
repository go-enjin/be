//go:build page_status || pages || all

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

package status

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
)

var (
	DefaultStatusCode = 404
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "page-status"

type Feature interface {
	feature.Feature
	feature.PageTypeProcessor
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
	if err = f.CFeature.Build(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) ProcessRequestPageType(r *http.Request, p feature.Page) (pg feature.Page, redirect string, processed bool, err error) {
	if p.Type() != "status" {
		return
	}

	var statusCode int
	if statusCode = p.Context().Int("StatusCode", 0); statusCode == 0 {
		// no specific StatusCode set
		if v, e := strconv.Atoi(p.Title()); e == nil {
			// given the page type is status and the title is actually an integer, use that
			statusCode = v
		} else {
			// default to something sane (404)
			statusCode = DefaultStatusCode
		}
	}
	var statusText string
	if statusText = http.StatusText(statusCode); statusText == "" {
		err = fmt.Errorf("unknown status code: %d", statusCode)
		return
	}
	statusText = p.Context().String("StatusText", statusText)

	p.Context().SetSpecific("StatusCode", statusCode)
	p.Context().SetSpecific("StatusText", statusText)

	pg = p
	processed = true
	return
}