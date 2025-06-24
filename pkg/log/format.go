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

package log

type Format int

const (
	FormatPretty Format = iota
	FormatJson
	FormatText
)

func (f Format) String() string {
	switch f {
	case FormatPretty:
		return "pretty"
	case FormatJson:
		return "json"
	case FormatText:
		return "text"
	}
	return ""
}

var formatLookup = map[string]Format{
	"pretty": FormatPretty,
	"json":   FormatJson,
	"text":   FormatText,
}
