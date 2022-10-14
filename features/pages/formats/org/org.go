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

package org

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/niklasfasching/go-org/org"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/search"
	"github.com/go-enjin/be/pkg/theme/types"
)

var (
	_ Feature      = (*CFeature)(nil)
	_ MakeFeature  = (*CFeature)(nil)
	_ types.Format = (*CFeature)(nil)
)

var _instance *CFeature

type Feature interface {
	feature.Feature
	types.Format
}

type MakeFeature interface {
	// SetDefault updates org.Configuration.DefaultSettings
	SetDefault(key, value string) MakeFeature

	// SetDefaults replaces org.Configuration.DefaultSettings
	SetDefaults(defaults map[string]string) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	settings map[string]string
	replaced map[string]string
}

func New() MakeFeature {
	if _instance == nil {
		_instance = new(CFeature)
		_instance.Init(_instance)
	}
	return _instance
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.settings = make(map[string]string)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = "PageFormatOrgMode"
	return
}

func (f *CFeature) SetDefault(key, value string) MakeFeature {
	f.settings[key] = value
	return f
}

func (f *CFeature) SetDefaults(settings map[string]string) MakeFeature {
	f.replaced = settings
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Name() (name string) {
	name = "org"
	return
}

func (f *CFeature) Label() (label string) {
	label = "Org-Mode"
	return
}

func (f *CFeature) Process(ctx context.Context, t types.Theme, content string) (html template.HTML, err *types.EnjinError) {
	input := strings.NewReader(content)
	orgConfig := org.New()
	if f.replaced != nil {
		orgConfig.DefaultSettings = f.replaced
	} else if len(f.settings) > 0 {
		for k, v := range f.settings {
			orgConfig.DefaultSettings[k] = v
			log.DebugF(`setting default: %v = "%v"`, k, v)
		}
	}
	if text, e := orgConfig.Parse(input, "./").Write(org.NewHTMLWriter()); e != nil {
		err = types.NewEnjinError(
			"org-mode parse error",
			e.Error(),
			content,
		)
		return
	} else {
		html = template.HTML(text)
	}
	return
}

func (f *CFeature) IndexDocument(ctx context.Context, content string) (doc search.Document, err error) {
	var url, title string
	if url = ctx.String("Url", ""); url == "" {
		err = fmt.Errorf("index document missing Url")
		return
	}
	if title = ctx.String("Title", ""); url == "" {
		err = fmt.Errorf("index document missing Title")
		return
	}

	doc = search.NewDocument(url, title)
	doc.AddContent(content)
	return
}