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

package feature

import (
	"encoding/json"
	"fmt"
	"html/template"
)

type EnjinError struct {
	Title   string
	Summary string
	Content string
}

func NewEnjinError(title, summary, content string) *EnjinError {
	return &EnjinError{
		Title:   title,
		Summary: summary,
		Content: content,
	}
}

func (e *EnjinError) Error() (message string) {
	message = fmt.Sprintf("%v: %v", e.Title, e.Summary)
	return
}

func (e *EnjinError) NjnData() (data map[string]interface{}) {
	b, _ := json.MarshalIndent(e.Content, "", "  ")
	data = map[string]interface{}{
		"type": "content",
		"content": map[string]interface{}{
			"section": []interface{}{
				map[string]interface{}{
					"type":    "details",
					"summary": e.Summary,
					"text": []interface{}{
						map[string]interface{}{
							"type": "pre",
							"text": string(b),
						},
					},
				},
			},
		},
	}
	return
}

func (e *EnjinError) Html() (html template.HTML) {
	html = `
<article
    class="block"
    data-block-tag="theme-enjin-error"
    data-block-type="content"
    data-block-profile="outer--inner"
    data-block-padding="both"
    data-block-margins="both"
    data-block-jump-top="true"
    data-block-jump-link="true">
    <div class="content">`
	if e.Title != "" {
		html += `
        <header>
            <h1>` + template.HTML(e.Title) + `</h1>
            <a class="jump-link" href="#theme-enjin-error">
                <span class="sr-only">permalink</span>
            </a>
        </header>`
	}
	html += `
        <section>
            <details>
                <summary>` + template.HTML(e.Summary) + `</summary>
                <pre>` + template.HTML(e.Content) + `</pre>
            </details>
        </section>
        <a class="jump-top" href="#top">top</a>
    </div>
</article>
`
	return
}