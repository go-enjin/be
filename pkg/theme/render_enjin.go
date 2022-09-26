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

package theme

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
)

type renderEnjin struct {
	theme *Theme
	ctx   context.Context

	blockCount   int
	headingLevel int
	headingCount int

	cache map[string]string

	sync.RWMutex
}

func newNjnRenderer(ctx context.Context, t *Theme) (re *renderEnjin) {
	re = new(renderEnjin)
	re.theme = t
	re.ctx = ctx
	re.headingLevel = 0
	re.cache = make(map[string]string)
	return
}

func (re *renderEnjin) render(ctx context.Context, data interface{}) (html template.HTML, err error) {

	switch v := data.(type) {

	case []interface{}:
		for _, c := range v {
			if h, e := re.render(ctx, c); e != nil {
				err = e
				return
			} else {
				html += h
			}
		}

	case map[string]interface{}:
		html, err = re.processBlock(ctx, v)

	default:
		err = fmt.Errorf("unsupported njn data received: %T", v)
	}

	return
}

func (re *renderEnjin) getNjnTemplateContent(name string) (contents string, err error) {
	// TODO: use the already prepared templating?
	if v, ok := re.cache[name]; ok {
		log.TraceF("found cached njn template: %v", name)
		contents = v
		return
	}
	path := bePath.JoinWithSlashes("layouts", "partials", "njn", name)
	log.TraceF("looking for njn template: %v - %v", name, path)
	var data []byte
	if data, err = re.theme.FileSystem.ReadFile(path); err == nil {
		contents = string(data)
		re.cache[name] = contents
	} else {
		err = fmt.Errorf("njn template not found: %v", name)
	}
	return
}

func (re *renderEnjin) renderNjnTemplate(tag string, data map[string]interface{}) (html template.HTML, err error) {
	var tmplContent string
	if tmplContent, err = re.getNjnTemplateContent(tag + ".tmpl"); err != nil {
		return
	} else {
		var tt *template.Template
		if tt, err = re.theme.NewHtmlTemplate(tag).Parse(tmplContent); err == nil {
			var w bytes.Buffer
			if err = tt.Execute(&w, data); err == nil {
				html = template.HTML(w.Bytes())
			} else {
				err = fmt.Errorf("error rendering template: %v", err)
			}
		} else {
			err = fmt.Errorf("error parsing template: %v", err)
		}
	}
	return
}