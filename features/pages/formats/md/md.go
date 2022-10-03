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

package md

import (
	"html/template"

	"github.com/gomarkdown/markdown"
	mdHtml "github.com/gomarkdown/markdown/html"
	mdParser "github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
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
	Make() Feature
}

type CFeature struct {
	feature.CFeature
}

func New() MakeFeature {
	if _instance == nil {
		_instance = new(CFeature)
		_instance.Init(_instance)
	}
	return _instance
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = "PageFormatMarkdown"
	return
}

func (f *CFeature) Name() (name string) {
	name = "md"
	return
}

func (f *CFeature) Label() (label string) {
	label = "Markdown"
	return
}

func (f *CFeature) Process(ctx context.Context, t types.Theme, content string) (html template.HTML, err error) {
	normalizedNewlines := markdown.NormalizeNewlines([]byte(content))
	extensions := mdParser.CommonExtensions |
		mdParser.AutoHeadingIDs |
		mdParser.NoIntraEmphasis |
		mdParser.Strikethrough |
		mdParser.Attributes
	pageParser := mdParser.NewWithExtensions(extensions)
	mdHtmlFlags := mdHtml.CommonFlags | mdHtml.HrefTargetBlank | mdHtml.FootnoteReturnLinks
	opts := mdHtml.RendererOptions{Flags: mdHtmlFlags}
	pageRenderer := mdHtml.NewRenderer(opts)
	parsedBytes := markdown.ToHTML(normalizedNewlines, pageParser, pageRenderer)
	sanitizedBytes := bluemonday.UGCPolicy().SanitizeBytes(parsedBytes)
	html = template.HTML(sanitizedBytes)
	return
}