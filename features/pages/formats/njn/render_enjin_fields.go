//go:build !exclude_pages_formats && !exclude_pages_format_njn

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

package njn

import (
	"html/template"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

func (re *RenderEnjin) ParseFieldAttributes(field map[string]interface{}) (attributes map[string]interface{}, classes []string, styles map[string]string, ok bool) {
	var err error
	if attributes, classes, styles, err = maps.ParseNjnFieldAttributes(field); err != nil {
		log.ErrorF("error parsing njn field attributes: %v", err)
		ok = false
		return
	}
	return
}

func (re *RenderEnjin) FinalizeFieldData(data map[string]interface{}, field map[string]interface{}, skip ...string) {
	if err := maps.FinalizeNjnFieldData(data, field, skip...); err != nil {
		log.ErrorF("error finalizing njn field data: %v", err)
	}
	return
}

func (re *RenderEnjin) FinalizeFieldAttributes(attrs map[string]interface{}) (attributes []template.HTMLAttr) {
	var err error
	if attributes, err = maps.FinalizeNjnFieldAttributes(attrs); err != nil {
		log.ErrorF("error finalizing njn field attributes: %v", err)
	}
	return
}