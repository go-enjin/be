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
	"fmt"
	"html/template"

	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/strings"
)

func (re *renderEnjin) renderSectionFields(fields []interface{}) (combined []template.HTML, err error) {
	combined, err = re.renderContainerFields(fields)
	return
}

func (re *renderEnjin) renderFooterFields(fields []interface{}) (combined []template.HTML, err error) {
	combined, err = re.renderContainerFields(fields)
	return
}

func (re *renderEnjin) finalizeFieldData(data map[string]interface{}, field map[string]interface{}, skip ...string) {
	for key, value := range field {
		switch {
		case strings.StringInStrings(key, skip...):
		default:
			if key == "attributes" {
				data["Attributes"], _, _, _ = re.parseFieldAttributes(field)
				continue
			}
			name := strcase.ToCamel(key)
			switch vv := value.(type) {
			case string:
				data[name] = template.HTML(vv)
			case int, int8, int16, int32, int64, float32, float64, bool:
				data[name] = fmt.Sprintf("%v", vv)
			default:
				data[name] = vv
			}
		}
	}
	return
}