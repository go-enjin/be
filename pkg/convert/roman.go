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
	"strconv"
	"strings"
)

var (
	MaximumRomanNumber = 3999
)

func ListRomanNumerals() (numerals DigitValues) {
	numerals = DigitValues{
		{1000, "M"},
		{900, "CM"},
		{500, "D"},
		{400, "CD"},
		{100, "C"},
		{90, "XC"},
		{50, "L"},
		{40, "XL"},
		{10, "X"},
		{9, "IX"},
		{5, "V"},
		{4, "IV"},
		{1, "I"},
	}
	return
}

func ToRoman(number int) (numerals string) {
	if number > MaximumRomanNumber {
		return strconv.Itoa(number)
	}

	var roman strings.Builder
	for _, conversion := range ListRomanNumerals().Sort() {
		for number >= conversion.Value {
			roman.WriteString(conversion.Digit)
			number -= conversion.Value
		}
	}

	return roman.String()
}
