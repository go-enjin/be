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

package locales

import (
	"fmt"
	"net/http"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/lang/catalog"
	"github.com/go-enjin/be/pkg/log"
)

func (f *CFeature) performDraftChanges(r *http.Request, translations map[string]interface{}, ld *LocaleData) (err error) {

	for code, v := range translations {
		var tag language.Tag
		if tag, err = language.Parse(code); err != nil {
			log.WarnRF(r, "error parsing language code: %v - %v", code, err)
			continue
		}
		if messages, ok := v.(map[string]interface{}); ok {
			for shasum, vv := range messages {
				switch t := vv.(type) {
				case string:
					if err = ld.SetStringTranslation(tag, shasum, t); err != nil {
						log.WarnRF(r, "error setting string translation: %v - %v - %v", code, shasum, err)
						continue
					}

				case map[string]interface{}:
					if arg, ok := t["arg"].(string); ok {
						var foundPlaceholder *catalog.Placeholder
						if msg, ok := ld.Data[shasum][tag]; ok {
							numerics := msg.Placeholders.Numeric()
							for _, placeholder := range numerics {
								if arg == placeholder.ID {
									foundPlaceholder = placeholder
								}
							}
						}
						if foundPlaceholder == nil {
							err = fmt.Errorf(`invalid placeholder: %q`, arg)
							return
						}
						if cases, ok := t["cases"].([]interface{}); ok {
							var list []string
							for _, v := range cases {
								if pair, ok := v.(map[string]interface{}); ok {
									if pk, ok := pair["key"].(string); ok {
										if pv, ok := pair["value"].(string); ok {
											list = append(list, pk, pv)
										}
									}
								}
							}
							if err = ld.SetPluralTranslation(tag, shasum, arg, list...); err != nil {
								log.WarnRF(r, "error setting plural translation: %v - %v - %v - %+v - %v", code, shasum, arg, list, err)
								continue
							}
						}
					}

				default:
					log.WarnRF(r, `unexpected message translation type: (%T) %#+v`, t, t)

				}
			}
		}
	}

	ld.ConvertAllPlaceholders()
	return
}