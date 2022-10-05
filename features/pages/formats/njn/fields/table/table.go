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

package table

import (
	"fmt"

	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/maps"
)

const (
	Tag feature.Tag = "NjnTableField"
)

var (
	_ Field     = (*CField)(nil)
	_ MakeField = (*CField)(nil)
)

var _instance *CField

type Field interface {
	feature.EnjinField
}

type MakeField interface {
	Make() Field
}

type CField struct {
	feature.CEnjinField
}

func New() (field MakeField) {
	if _instance == nil {
		_instance = new(CField)
		_instance.Init(_instance)
	}
	field = _instance
	return
}

func (f *CField) Tag() feature.Tag {
	return Tag
}

func (f *CField) Init(this interface{}) {
	f.CEnjinField.Init(this)
}

func (f *CField) Make() Field {
	return f
}

func (f *CField) NjnTagClass() (tagClass feature.NjnTagClass) {
	tagClass = feature.ContainerNjnTag
	return
}

func (f *CField) NjnFieldNames() (name []string) {
	name = append(name, "table")
	return
}

func (f *CField) PrepareNjnData(re feature.EnjinRenderer, tagName string, field map[string]interface{}) (data map[string]interface{}, err error) {
	if tagName != "table" {
		err = fmt.Errorf(`%v feature does not support tags named: "%v"`, Tag, tagName)
		return
	}

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
							if t, e := re.RenderInlineFieldList(vv); e != nil {
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
						if attrs, _, _, e := maps.ParseNjnFieldAttributes(bodyRowMap); e != nil {
							err = e
							return
						} else {
							if data["Attributes"], err = maps.FinalizeNjnFieldAttributes(attrs); err != nil {
								return
							}
						}
						var rowCells []interface{}
						if bodyRowDataList, ok := bodyRowMap["data"].([]interface{}); ok {
							for _, bodyRowDataItem := range bodyRowDataList {
								if bodyRowDataMap, ok := bodyRowDataItem.(map[string]interface{}); ok {
									if bodyRowDataType, ok := bodyRowDataMap["type"].(string); ok {
										switch bodyRowDataType {
										case "td":
											rowData := make(map[string]interface{})
											rowData["Type"] = "td"
											if attrs, _, _, e := maps.ParseNjnFieldAttributes(bodyRowDataMap); e != nil {
												err = e
												return
											} else {
												if data["Attributes"], err = maps.FinalizeNjnFieldAttributes(attrs); err != nil {
													return
												}
											}
											if rowData["Data"], err = re.RenderContainerFieldText(bodyRowDataMap); err != nil {
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

	err = maps.FinalizeNjnFieldData(data, field, "type", "head", "body", "foot")
	return
}