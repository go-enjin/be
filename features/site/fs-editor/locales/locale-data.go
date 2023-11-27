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
	"strings"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/lang/catalog"
)

type LocaleData struct {
	FSID  string                                     `json:"fsid"`
	Code  string                                     `json:"code"`
	Data  map[string]map[language.Tag]*LocaleMessage `json:"data"`
	Order []string                                   `json:"order"`
}

func ConvertToPlaceholders(translation string, placeholders []*catalog.Placeholder) (modified string) {
	modified = translation
	for _, placeholder := range placeholders {
		modified = strings.ReplaceAll(modified, placeholder.String, "{"+placeholder.ID+"}")
	}
	return
}

func ConvertFromPlaceholders(translation string, placeholders []*catalog.Placeholder) (modified string) {
	modified = translation
	for _, placeholder := range placeholders {
		modified = strings.ReplaceAll(modified, "{"+placeholder.ID+"}", placeholder.String)
	}
	return
}

func (l *LocaleData) ConvertAllPlaceholders() {
	for _, txs := range l.Data {
		for _, tx := range txs {
			if tx.Translation.Select != nil {
				for k, v := range tx.Translation.Select.Cases {
					tx.Translation.Select.Cases[k] = ConvertFromPlaceholders(strings.TrimSpace(v), tx.Placeholders)
				}
			} else {
				tx.Translation.String = ConvertFromPlaceholders(strings.TrimSpace(tx.Translation.String), tx.Placeholders)
			}
		}
	}
}

func (l *LocaleData) SetStringTranslation(tag language.Tag, shasum, translation string) (err error) {
	var found bool
	if txs, ok := l.Data[shasum]; ok {
		if _, found = txs[tag]; found {
			l.Data[shasum][tag].Translation.String = ConvertFromPlaceholders(translation, l.Data[shasum][tag].Placeholders)
		}
	}
	if !found {
		err = fmt.Errorf("message not found")
	}
	return
}

func (l *LocaleData) SetPluralTranslation(tag language.Tag, shasum, arg string, cases ...string) (err error) {
	if len(cases)%2 != 0 {
		err = fmt.Errorf("unbalanced pluralization cases list")
		return
	}
	var found bool
	if txs, ok := l.Data[shasum]; ok {
		if _, found = txs[tag]; found {
			if l.Data[shasum][tag].Translation == nil {
				l.Data[shasum][tag].Translation = &LocaleTranslation{
					String: "",
					Select: &Select{
						Feature: "plural",
					},
				}
			} else if l.Data[shasum][tag].Translation.Select == nil {
				l.Data[shasum][tag].Translation.Select = &Select{
					Feature: "plural",
				}
			}
			l.Data[shasum][tag].Translation.Select.Arg = arg
			l.Data[shasum][tag].Translation.Select.Cases = make(map[string]string)
			for i := 0; i < len(cases); i += 2 {
				l.Data[shasum][tag].Translation.Select.Cases[cases[i]] = ConvertFromPlaceholders(cases[i+1], l.Data[shasum][tag].Placeholders)
			}
		}
	}
	if !found {
		err = fmt.Errorf("message not found")
	}
	return
}

func (l *LocaleData) MakeGoTextData() (lookup map[language.Tag]*catalog.GoText) {
	unique := map[language.Tag]map[string]struct{}{}
	lookup = map[language.Tag]*catalog.GoText{}

	for _, shasum := range l.Order {
		for tag, msg := range l.Data[shasum] {
			if _, present := lookup[tag]; !present {
				lookup[tag] = &catalog.GoText{Language: tag.String()}
				unique[tag] = map[string]struct{}{}
			}
			if _, present := unique[tag][msg.ID]; present {
				continue
			}
			unique[tag][msg.ID] = struct{}{}
			var plural *catalog.Select
			if msg.Translation.Select != nil {
				plural = &catalog.Select{
					Arg:     msg.Translation.Select.Arg,
					Feature: msg.Translation.Select.Feature,
					Cases:   map[string]catalog.SelectCase{},
				}
				for k, v := range msg.Translation.Select.Cases {
					plural.Cases[k] = catalog.SelectCase{Msg: v}
				}
			}
			lookup[tag].Messages = append(lookup[tag].Messages, &catalog.Message{
				BaseMessage: msg.BaseMessage,
				Translation: &catalog.Translation{
					String: msg.Translation.String,
					Select: plural,
				},
			})
		}
	}

	return
}

func (l *LocaleData) AddMissingTranslations(defTag language.Tag, locales []language.Tag) {
	for _, shasum := range l.Order {
		if txs, ok := l.Data[shasum]; ok {
			if defMsg, ok := txs[defTag]; ok {
				for _, tag := range locales {
					if _, present := txs[tag]; present {
						continue
					}
					txs[tag] = defMsg.Copy()
				}
			}
		}
	}
}