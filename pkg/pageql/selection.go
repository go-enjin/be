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

package pageql

import "encoding/json"

type Selection struct {
	Selecting []*Selecting `parser:"'SELECT' @@ ( ',' @@ )*" json:"selecting,omitempty"`
	Statement *Statement   `parser:"( 'WITHIN' @@ )?" json:"statement,omitempty"`
}

func (s *Selection) String() (query string) {
	query = ""
	if numSelecting := len(s.Selecting); numSelecting > 0 {
		query += "SELECT"
		for idx, sel := range s.Selecting {
			if idx > 0 {
				query += ","
			}
			query += " " + sel.String()
		}
	}
	if s.Statement != nil {
		if v := s.Statement.String(); v != "" {
			if query != "" {
				query += " "
			}
			query += "WITHIN " + v
		}
	}
	return
}

func (s *Selection) Render() (out *Selection) {
	out = new(Selection)
	for _, sel := range s.Selecting {
		out.Selecting = append(out.Selecting, sel.Render())
	}
	if s.Statement != nil {
		out.Statement = s.Statement.Render()
	}
	return
}

func (s *Selection) Stringify() (out string) {
	b, _ := json.MarshalIndent(s.Render(), "", "  ")
	out = string(b)
	return
}