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

package page

import (
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/theme/types"
)

var (
	knownFormats = make(map[string]types.Format)
	Extensions   = make([]string, 0)
)

func GetFormat(name string) (format types.Format) {
	if v, ok := knownFormats[name]; ok {
		format = v
	}
	return
}

func AddFormat(format types.Format) {
	knownFormats[format.Name()] = format
	if !beStrings.StringInStrings(format.Name(), Extensions...) {
		Extensions = append(Extensions, format.Name())
	}
}

func RemoveFormat(name string) {
	delete(knownFormats, name)
	if idx := beStrings.StringIndexInStrings(name, Extensions...); idx >= 0 {
		Extensions = beStrings.RemoveIndexFromStrings(idx, Extensions)
	}
}