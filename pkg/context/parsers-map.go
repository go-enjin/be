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

package context

import (
	"errors"
	"strings"

	"github.com/microcosm-cc/bluemonday"

	"github.com/go-corelibs/values"
)

func StringMapParser(spec *Field, input interface{}) (parsed interface{}, err error) {
	strict := bluemonday.StrictPolicy()
	switch t := input.(type) {
	case map[string]string:
		cleaned := make(map[string]string)
		for dk, dv := range t {
			cleaned[dk] = strict.Sanitize(strings.TrimSpace(dv))
		}
		parsed = cleaned
	default:
		err = errors.New(spec.Printer.Sprintf("unsupported type: %[1]s", values.TypeOf(input)))
	}
	return
}
