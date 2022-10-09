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
	"fmt"
	"html/template"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

func (re *RenderEnjin) PrepareInlineFieldText(field map[string]interface{}) (combined []interface{}, err error) {
	if typeName, ok := re.ParseTypeName(field); ok {

		if njnField, ok := re.Njn.FindField(feature.AnyNjnClass, typeName); ok {
			if textItem, ok := field["text"]; ok {

				switch t := textItem.(type) {
				case []interface{}:
					for _, item := range t {
						if childText, ok := item.(string); ok {
							combined = append(combined, childText)
						} else {
							if _, child, _, e := re.CheckInlineFieldText(njnField, typeName, item); e != nil {
								err = e
								return
							} else {
								combined = append(combined, child)
							}
						}
					}

				case interface{}:
					if childText, ok := t.(string); ok {
						if prepared, e := re.PrepareInlineFieldList([]interface{}{childText}); e != nil {
							err = e
							return
						} else {
							combined = append(combined, prepared)
						}
					} else {
						if _, child, _, e := re.CheckInlineFieldText(njnField, typeName, t); e != nil {
							err = e
							return
						} else {
							if prepared, ee := re.PrepareInlineFieldList([]interface{}{child}); ee != nil {
								err = ee
								return
							} else {
								combined = append(combined, prepared)
							}
						}
					}

				default:
					err = fmt.Errorf("unsupported inline field text structure: %T", field["text"])
				}
			} else {
				err = fmt.Errorf("inline field missing text")
			}
		} else {
			err = fmt.Errorf("inline field not found: %v", typeName)
		}
	} else {
		err = fmt.Errorf("inline field missing type")
	}

	return
}

func (re *RenderEnjin) PrepareInlineFieldList(list []interface{}) (combined []interface{}, err error) {
	for _, item := range list {
		switch typedItem := item.(type) {
		case string:

			// TODO: explore how to cross field list string boundaries
			if parsed, e := re.PrepareStringTags(typedItem); e != nil {
				err = fmt.Errorf("error parsing shortcodes: %v", e)
				return
			} else {
				combined = append(combined, parsed...)
			}

			// if idx > 0 {
			// 	if _, ok := list[idx-1].(string); ok {
			// 		if lastIndex := len(combined) - 1; lastIndex >= 0 {
			// 			if v, ok := combined[lastIndex].(string); ok {
			// 				combined[lastIndex] = v + " " + typedItem
			// 			} else {
			// 				combined = append(combined, typedItem)
			// 			}
			// 		} else {
			// 			combined = append(combined, typedItem)
			// 		}
			// 		continue
			// 	}
			// }
			// // combined = append(combined, template.HTML(typedItem))
			// combined = append(combined, typedItem)

		case []interface{}:
			if prepared, e := re.PrepareInlineFieldList(typedItem); e != nil {
				err = e
				return
			} else {
				combined = append(combined, prepared...)
			}

		case map[string]interface{}:
			if prepared, e := re.PrepareInlineField(typedItem); e != nil {
				err = e
				return
			} else {
				combined = append(combined, prepared)
			}

		default:
			err = fmt.Errorf("unsupported inline field text structure: %T", item)
			return
		}
	}
	return
}

func (re *RenderEnjin) PrepareInlineFields(fields []interface{}) (combined []interface{}, err error) {
	for _, field := range fields {
		switch fieldTyped := field.(type) {
		case string:
			// combined = append(combined, fieldTyped)
			if parsed, e := re.PrepareStringTags(fieldTyped); e != nil {
				err = fmt.Errorf("error parsing shortcodes: %v", e)
				return
			} else {
				combined = append(combined, parsed...)
			}
		case map[string]interface{}:
			if c, e := re.PrepareInlineField(fieldTyped); e != nil {
				err = e
				return
			} else {
				combined = append(combined, c)
			}
		case []interface{}:
			if c, e := re.PrepareInlineFieldList(fieldTyped); e != nil {
				err = e
				return
			} else {
				combined = append(combined, c...)
			}
		default:
			err = fmt.Errorf("unsupported inline field structure: %T", field)
			return
		}
	}
	return
}

func (re *RenderEnjin) PrepareInlineField(field map[string]interface{}) (prepared map[string]interface{}, err error) {
	if ft, ok := re.ParseTypeName(field); ok {

		if inlineField, ok := re.Njn.FindField(feature.InlineNjnClass, ft); ok {
			if prepared, err = inlineField.PrepareNjnData(re, ft, field); err != nil {
				return
			}
		} else {
			err = fmt.Errorf("unsupported inline field type: %v", ft)
			return
		}

	} else {
		err = fmt.Errorf("inline field missing type")
	}
	return
}

func (re *RenderEnjin) CheckInlineFieldText(parent feature.EnjinField, parentName string, child interface{}) (njn feature.EnjinField, field map[string]interface{}, name string, err error) {
	if childField, childName, ok := re.ParseFieldAndTypeName(child); ok {
		if childNjnField, ok := re.Njn.FindField(feature.InlineNjnClass, childName); ok {
			if parent.NjnCheckTag(childName) && parent.NjnCheckClass(childNjnField.NjnClass()) {
				njn = childNjnField
				field = childField
				name = childName
			} else {
				log.TraceF("%v denied as child of %v", childName, parentName)
			}
		} else {
			err = fmt.Errorf("inline njn field not found: %v", childName)
			return
		}
	} else {
		err = fmt.Errorf("inline field missing type or unsupported structure: %T", child)
	}
	return
}

func (re *RenderEnjin) RenderInlineFields(fields []interface{}) (combined []template.HTML, err error) {
	for _, field := range fields {
		switch fieldTyped := field.(type) {
		case string:
			// combined = append(combined, template.HTML(fieldTyped))
			if parsed, e := re.PrepareStringTags(fieldTyped); e != nil {
				err = fmt.Errorf("error parsing shortcodes: %v", e)
				return
			} else {
				if c, e := re.RenderInlineFields(parsed); e != nil {
					err = e
					return
				} else {
					combined = append(combined, c...)
				}
			}
		case map[string]interface{}:
			if c, e := re.RenderInlineField(fieldTyped); e != nil {
				err = e
				return
			} else {
				combined = append(combined, c...)
			}
		case []interface{}:
			if c, e := re.RenderInlineFieldList(fieldTyped); e != nil {
				err = e
				return
			} else {
				combined = append(combined, c)
			}
		default:
			err = fmt.Errorf("unsupported inline field structure: %T", field)
			return
		}
	}
	return
}

func (re *RenderEnjin) RenderInlineField(field map[string]interface{}) (combined []template.HTML, err error) {
	if ft, ok := re.ParseTypeName(field); ok {
		var data map[string]interface{}

		if inlineField, ok := re.Njn.FindField(feature.InlineNjnClass, ft); ok {
			if data, err = inlineField.PrepareNjnData(re, ft, field); err != nil {
				return
			}
		} else {
			err = fmt.Errorf("unsupported inline field type: %v", ft)
			return
		}

		if html, e := re.RenderNjnTemplate("field/"+ft, data); e != nil {
			err = e
			return
		} else {
			combined = append(combined, html)
		}
	} else {
		err = fmt.Errorf("inline field missing type")
	}
	return
}

func (re *RenderEnjin) RenderInlineFieldText(field map[string]interface{}) (html template.HTML, err error) {
	if typeName, ok := re.ParseTypeName(field); ok {
		if njnField, ok := re.Njn.FindField(feature.AnyNjnClass, typeName); ok {
			if textItem, ok := field["text"]; ok {
				switch t := textItem.(type) {
				case []interface{}:
					var allowed []interface{}
					for _, item := range t {
						if childText, ok := item.(string); ok {
							allowed = append(allowed, childText)
						} else {
							if _, child, _, e := re.CheckInlineFieldText(njnField, typeName, item); e != nil {
								err = e
								return
							} else {
								allowed = append(allowed, child)
							}
						}
					}
					html, err = re.RenderInlineFieldList(allowed)
				case interface{}:
					if childText, ok := t.(string); ok {
						html, err = re.RenderInlineFieldList([]interface{}{childText})
					} else {
						if _, child, _, e := re.CheckInlineFieldText(njnField, typeName, t); e != nil {
							err = e
							return
						} else {
							html, err = re.RenderInlineFieldList([]interface{}{child})
						}
					}
				default:
					err = fmt.Errorf("unsupported inline field text structure: %T", field["text"])
				}
			} else {
				err = fmt.Errorf("inline field missing text")
			}
		} else {
			err = fmt.Errorf("inline field not found: %v", typeName)
		}
	} else {
		err = fmt.Errorf("inline field missing type")
	}

	return
}

func (re *RenderEnjin) RenderInlineFieldList(list []interface{}) (html template.HTML, err error) {
	for idx, item := range list {
		if itemString, ok := item.(string); ok {
			if idx > 0 {
				if _, ok := list[idx-1].(string); ok {
					html += " "
				}
			}
			html += template.HTML(itemString)
		} else if itemMap, ok := item.(map[string]interface{}); ok {
			var rendered []template.HTML
			if rendered, err = re.RenderInlineField(itemMap); err != nil {
				html = ""
				return
			}
			for _, rend := range rendered {
				html += rend
			}
		} else if itemList, ok := item.([]interface{}); ok {
			html, err = re.RenderInlineFieldList(itemList)
		} else {
			html = ""
			err = fmt.Errorf("unsupported inline field text structure: %T", item)
			return
		}
	}
	return
}