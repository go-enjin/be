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

type Selecting struct {
	Count      bool   `parser:"( @'COUNT' )?" json:"count,omitempty"`
	Random     bool   `parser:"( @'RANDOM' )?" json:"random,omitempty"`
	Distinct   bool   `parser:"( @'DISTINCT' )?" json:"distinct,omitempty"`
	ContextKey string `parser:" '.' @Ident" json:"context-key"`
}

func (s *Selecting) String() (query string) {
	if s.Count {
		query += "COUNT"
	}
	if s.Random {
		query += "RANDOM"
	}
	if s.Distinct {
		if query != "" {
			query += " "
		}
		query += "DISTINCT"
	}
	if query != "" {
		query += " "
	}
	query += "." + s.ContextKey
	return
}

func (s *Selecting) Render() (out *Selecting) {
	out = new(Selecting)
	out.Count = s.Count
	out.Random = s.Random
	out.Distinct = s.Distinct
	out.ContextKey = s.ContextKey
	return
}