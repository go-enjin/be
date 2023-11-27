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
	"encoding/json"
)

type BaseMessage struct {
	ID                string       `json:"id"`
	Key               string       `json:"key"`
	Message           string       `json:"message"`
	TranslatorComment string       `json:"translatorComment,omitempty"`
	Fuzzy             bool         `json:"fuzzy,omitempty"`
	Placeholders      Placeholders `json:"placeholders,omitempty"`
}

type Translation struct {
	String string  `json:"string,omitempty"`
	Select *Select `json:"select,omitempty"`
}

type Message struct {
	BaseMessage
	Translation *Translation `json:"translation"`
}

func (m *Message) UnmarshalJSON(data []byte) (err error) {
	structMessage := &SelectMessage{}
	if e := json.Unmarshal(data, &structMessage); e == nil {
		*m = structMessage.Make()
		return
	}
	stringMessage := &StringMessage{}
	if e := json.Unmarshal(data, &stringMessage); e == nil {
		*m = stringMessage.Make()
		return
	} else {
		err = e
	}
	return
}

func (m *Message) MarshalJSON() (data []byte, err error) {
	if m.Translation.Select != nil {
		data, err = json.MarshalIndent(struct {
			BaseMessage
			Translation map[string]interface{} `json:"translation"`
		}{
			BaseMessage: m.BaseMessage,
			Translation: map[string]interface{}{
				"select": m.Translation.Select,
			},
		}, "", "\t")
		return
	}
	data, err = json.MarshalIndent(struct {
		BaseMessage
		Translation string `json:"translation"`
	}{
		BaseMessage: m.BaseMessage,
		Translation: m.Translation.String,
	}, "", "\t")
	return
}

func (m *Message) Copy() (copied *Message) {
	copied = &Message{
		BaseMessage: BaseMessage{
			ID:                m.ID,
			Key:               m.Key,
			Message:           m.Message,
			TranslatorComment: m.TranslatorComment,
			Fuzzy:             m.Fuzzy,
			Placeholders:      m.Placeholders.Copy(),
		},
	}
	if m.Translation.Select != nil {
		copied.Translation.Select = &Select{
			Arg:     m.Translation.Select.Arg,
			Feature: m.Translation.Select.Feature,
			Cases:   make(map[string]SelectCase),
		}
		for k, sc := range m.Translation.Select.Cases {
			copied.Translation.Select.Cases[k] = SelectCase{
				Msg: sc.Msg,
			}
		}
	} else {
		copied.Translation = &Translation{
			String: m.Translation.String,
		}
	}
	return
}
