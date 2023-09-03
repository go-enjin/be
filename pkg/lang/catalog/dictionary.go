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

package catalog

import (
	"fmt"
	"strings"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/go-enjin/golang-org-x-text/message/catalog"

	"github.com/go-enjin/be/pkg/log"
)

var (
	_ catalog.Dictionary = (*dictionary)(nil)
)

type dictionary struct {
	tag language.Tag
	src string
	msg map[string]string
	key map[string]string
}

func newDictionary(tag language.Tag, src string) (d *dictionary) {
	d = &dictionary{
		tag: tag,
		src: src,
		msg: make(map[string]string),
		key: make(map[string]string),
	}
	return
}

func newDictionaryFromJsonData(tag language.Tag, src string, data map[string]interface{}) (d *dictionary) {
	d = newDictionary(tag, src)
	if messages, ok := data["messages"].([]interface{}); ok {
		for _, m := range messages {
			if msg, ok := m.(map[string]interface{}); ok {
				if id, ok := msg["id"].(string); ok {
					keyOrId := id
					if key, ok := msg["key"].(string); ok {
						keyOrId = key
					}
					if msgTranslation, ok := msg["translation"]; ok {
						if tx, ok := msgTranslation.(string); !ok {
							log.WarnF("translation is not a string: (%T) %#+v", msgTranslation, msgTranslation)
						} else {
							translated := tx
							if phList, ok := msg["placeholders"].([]interface{}); ok {
								rpl := map[string]string{}
								for _, phItem := range phList {
									if placeholder, ok := phItem.(map[string]interface{}); ok {
										if phId, ok := placeholder["id"].(string); ok {
											if phString, ok := placeholder["string"].(string); ok {
												rpl[phId] = phString
											}
										}
									}
								}
								if len(rpl) > 0 {
									for k, v := range rpl {
										translated = strings.ReplaceAll(translated, fmt.Sprintf("{%v}", k), v)
									}
								}
							}
							d.msg[id] = tx
							d.key[keyOrId] = translated
						}
					}
				}
			}
		}
	}
	return
}

func (d *dictionary) Lookup(key string) (data string, ok bool) {
	if data, ok = d.msg[key]; ok {
		log.DebugF("found [%v]: %v", d.tag.String(), data)
	} else {
		log.DebugF("not found [%v]: %v\n%#v", d.tag.String(), key, d.msg)
	}
	return
}