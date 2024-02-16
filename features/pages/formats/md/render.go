//go:build !exclude_pages_formats && !exclude_pages_format_md

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
	"strings"

	"github.com/gomarkdown/markdown"
	mdHtml "github.com/gomarkdown/markdown/html"
	mdParser "github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"

	"github.com/go-corelibs/slices"
)

func PreRenderMarkdown(content string) (parsedBytes []byte) {
	normalizedNewlines := markdown.NormalizeNewlines([]byte(content))
	extensions := mdParser.CommonExtensions |
		mdParser.AutoHeadingIDs |
		mdParser.Strikethrough |
		mdParser.Attributes |
		mdParser.Footnotes |
		mdParser.FencedCode |
		mdParser.OrderedListStart
	pageParser := mdParser.NewWithExtensions(extensions)
	mdHtmlFlags := mdHtml.FootnoteReturnLinks
	opts := mdHtml.RendererOptions{Flags: mdHtmlFlags}
	pageRenderer := mdHtml.NewRenderer(opts)
	parsedBytes = markdown.ToHTML(normalizedNewlines, pageParser, pageRenderer)
	return
}

func RenderExcerpt(content string) (excerpt string) {
	parsedBytes := PreRenderMarkdown(content)
	policy := bluemonday.StrictPolicy()
	policy.AllowComments()
	sane := policy.SanitizeBytes(parsedBytes)
	plain := string(sane)
	if before, middle, _, found := slices.CarveString(plain, "<!--", "-->"); found {
		if strings.ToLower(strings.TrimSpace(middle)) == "more" {
			excerpt = strings.TrimPrefix(before, "<p>")
		}
	}
	return
}

func RenderMarkdown(content string) (text string) {
	parsedBytes := PreRenderMarkdown(content)
	sanitizedBytes := bluemonday.UGCPolicy().SanitizeBytes(parsedBytes)
	text = string(sanitizedBytes)
	return
}
