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
	"strings"

	"github.com/go-enjin/be/pkg/maths"
)

func ToLetters[T maths.Integers](number T) (letters string) {
	letters = ToCharacters(number, "abcdefghijklmnopqrstuvwxyz")
	return
}

func ToCharacters[T maths.Integers](number T, base string) (letters string) {
	var builder strings.Builder
	defer func() { letters = builder.String() }()
	size := T(len(base))

	if base == "" {
		return
	} else if number >= size {
		builder.WriteString(ToCharacters((number/size)-1, base))
	}

	builder.WriteByte(base[number%size])
	return
}
