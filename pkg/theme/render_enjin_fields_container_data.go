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
	"strings"

	"github.com/iancoleman/strcase"
)

func (re *renderEnjin) prepareLiteralFieldData(tag string, field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = tag
	re.finalizeFieldData(data, field, "type")
	return
}

func (re *renderEnjin) preparePreFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = "pre"
	var dataText template.HTML
	if dataText, err = re.renderInlineFieldText(field); err != nil {
		return
	}
	data["Text"] = template.HTML(template.HTMLEscapeString(string(dataText)))
	re.finalizeFieldData(data, field, "type", "text")
	return
}

func (re *renderEnjin) prepareTableFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = "table"

	if hl, ok := field["head"].([]interface{}); ok {
		dataHead := make([]map[string]interface{}, 0)
		for idx, hi := range hl {
			if hm, ok := hi.(map[string]interface{}); ok {
				if hmt, ok := hm["type"].(string); ok {
					switch hmt {
					case "th":
						heading := make(map[string]interface{})
						for k, v := range hm {
							key := strcase.ToCamel(k)
							var vv []interface{}
							if vv, ok = v.([]interface{}); !ok {
								vv = []interface{}{v}
							}
							if t, e := re.renderInlineFieldList(vv); e != nil {
								err = e
								return
							} else {
								heading[key] = t
							}
						}
						dataHead = append(dataHead, heading)
					default:
						err = fmt.Errorf("unsupported table heading type: %v [index=%d]", hmt, idx)
						return
					}
				}
			}
		}
		data["Head"] = dataHead
	}

	if hl, ok := field["body"].([]interface{}); ok {
		var dataBody []interface{}
		for idx, hi := range hl {
			if bodyRowMap, ok := hi.(map[string]interface{}); ok {
				if bodyRowType, ok := bodyRowMap["type"].(string); ok {
					switch bodyRowType {
					case "tr":

						row := make(map[string]interface{})
						row["Type"] = "tr"
						row["Attributes"], _, _, _ = re.parseFieldAttributes(bodyRowMap)
						var rowCells []interface{}
						if bodyRowDataList, ok := bodyRowMap["data"].([]interface{}); ok {
							for _, bodyRowDataItem := range bodyRowDataList {
								if bodyRowDataMap, ok := bodyRowDataItem.(map[string]interface{}); ok {
									if bodyRowDataType, ok := bodyRowDataMap["type"].(string); ok {
										switch bodyRowDataType {
										case "td":
											rowData := make(map[string]interface{})
											rowData["Type"] = "td"
											rowData["Attributes"], _, _, _ = re.parseFieldAttributes(bodyRowDataMap)
											if rowData["Data"], err = re.renderContainerFieldText(bodyRowDataMap); err != nil {
												return
											}
											rowCells = append(rowCells, rowData)
										}
									} else {
										err = fmt.Errorf("body row data map missing type: %+v", bodyRowDataItem)
										return
									}
								} else {
									err = fmt.Errorf("body row data map is not a map: %T %+v", bodyRowDataItem, bodyRowDataItem)
									return
								}
							}
						}
						row["Data"] = rowCells
						dataBody = append(dataBody, row)

					default:
						err = fmt.Errorf("unsupported table heading type: %v [index=%d]", bodyRowType, idx)
						return
					}
				}
			}
		}
		data["Body"] = dataBody
	}

	re.finalizeFieldData(data, field, "type", "head", "body", "foot")
	return
}

func (re *renderEnjin) prepareDescriptionListFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = "dl"
	if data["Text"], err = re.renderContainerFieldText(field); err != nil {
		return
	}
	re.finalizeFieldData(data, field, "type", "text")
	return
}

func (re *renderEnjin) prepareCodeFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = "code"
	decorated := false
	if v, ok := field["decorated"].(string); ok {
		lv := strings.ToLower(v)
		decorated = lv == "true" || lv == "1"
	}
	data["Decorated"] = decorated

	if lines, ok := field["code"].([]interface{}); ok {
		if decorated {
			data["Lines"] = lines
		} else {
			text := ""
			for _, line := range lines {
				if v, ok := line.(string); ok {
					if text != "" {
						text += "\n"
					}
					text += v
				}
			}
			data["Text"] = text
		}
	}

	if attrs, classes, _, ok := re.parseFieldAttributes(field); ok {
		if decorated {
			classes = append(classes, "decorated")
			attrs["class"] = strings.Join(classes, " ")
		}
		data["Attributes"] = attrs
	} else if decorated {
		data["Attributes"] = map[string]interface{}{
			"class": "decorated",
		}
	}

	re.finalizeFieldData(data, field, "type", "decorated", "code", "attributes")
	return
}

func (re *renderEnjin) prepareListFieldData(tag string, field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = tag
	var ok bool
	var list []interface{}
	if list, ok = field["list"].([]interface{}); !ok {
		err = fmt.Errorf("ordered list missing list: %+v", field)
		return
	}
	var combined []template.HTML
	if combined, err = re.renderInlineFields(list); err != nil {
		return
	}
	data["Items"] = combined

	re.finalizeFieldData(data, field, "type", "list")
	return
}

func (re *renderEnjin) prepareFieldsetFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = "fieldset"
	if list, ok := field["legend"].([]interface{}); ok {
		if data["Legend"], err = re.renderInlineFields(list); err != nil {
			err = fmt.Errorf("error rendering fieldset legend: %v", err)
			return
		}
	}
	if list, ok := field["fields"].([]interface{}); ok {
		if data["Fields"], err = re.renderContainerFields(list); err != nil {
			err = fmt.Errorf("error rendering fieldset fields: %v", err)
			return
		}
	}
	re.finalizeFieldData(data, field, "type", "legend", "fields")
	return
}

func (re *renderEnjin) prepareDetailsFieldData(field map[string]interface{}) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	data["Type"] = "details"
	data["Summary"], _ = field["summary"]
	if data["Text"], err = re.renderContainerFieldText(field); err != nil {
		return
	}
	re.finalizeFieldData(data, field, "type", "summary", "text")
	return
}