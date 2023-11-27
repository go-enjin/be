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

package convert

import (
	"sort"
)

type DigitValue struct {
	Value int
	Digit string
}

type DigitValues []DigitValue

func (r DigitValues) Sort() (sorted DigitValues) {
	sorted = append(sorted, r...)
	sort.Slice(sorted, func(i, j int) (less bool) {
		less = sorted[i].Value < sorted[j].Value
		return
	})
	return
}
