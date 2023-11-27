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

type StringMessage struct {
	BaseMessage
	Translation string `json:"translation"`
}

func (s *StringMessage) Make() (m Message) {
	m = Message{
		BaseMessage: s.BaseMessage,
		Translation: &Translation{
			String: s.Translation,
		},
	}
	return
}

func (s *StringMessage) MarshalJSON() (data []byte, err error) {
	sm := struct {
		BaseMessage
		Translation string `json:"translation"`
	}{
		BaseMessage: s.BaseMessage,
		Translation: s.Translation,
	}
	data, err = json.Marshal(sm)
	return
}
