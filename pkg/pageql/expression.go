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

type Expression struct {
	Condition *Condition `parser:"( @@ )?" json:"condition,omitempty"`
	Operation *Operation `parser:"( @@ )?" json:"operation,omitempty"`
}

func (e *Expression) Render() (clone *Expression) {
	clone = new(Expression)
	if e.Condition != nil {
		clone.Condition = e.Condition.Render()
	}
	if e.Operation != nil {
		clone.Operation = e.Operation.Render()
	}
	return
}
