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
	"fmt"
	"strconv"
	"strings"
)

type Select struct {
	Arg     string                `json:"arg"`
	Feature string                `json:"feature"`
	Cases   map[string]SelectCase `json:"cases"`
}

type SelectCase struct {
	Msg string `json:"msg"`
}

type SelectMessage struct {
	BaseMessage
	Translation *Translation `json:"translation"`
}

func (s *SelectMessage) Make() (m Message) {
	m = Message{
		BaseMessage: s.BaseMessage,
		Translation: s.Translation,
	}
	return
}

func (s *SelectMessage) MarshalJSON() (data []byte, err error) {
	sm := struct {
		BaseMessage
		Translation *Select `json:"translation"`
	}{
		BaseMessage: s.BaseMessage,
		Translation: s.Translation.Select,
	}
	data, err = json.Marshal(sm)
	return
}

func ParsePluralCaseKey(key string) (real string) {
	real = strings.ToLower(strings.TrimSpace(key))
	switch real {
	case "zero", "one", "two", "few", "many", "other":
		return
	}
	if len(real) <= 1 {
		real = ""
		return
	}
	switch real[0] {
	case '=':
		if v, ee := strconv.Atoi(real[1:]); ee != nil {
			real = ""
		} else {
			real = fmt.Sprintf("=%d", v)
		}
	case '<':
		if v, ee := strconv.Atoi(real[1:]); ee != nil {
			real = ""
		} else {
			real = fmt.Sprintf("<%d", v)
		}
	default:
		real = ""
	}
	return
}
