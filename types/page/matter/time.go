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

package matter

import (
	"fmt"
	"sync"
	"time"
)

var (
	knownDateTimeLayouts = []string{
		"2006-01-02 15:04 MST",
		"Jan 01, 2006",
	}
	knownDateTimeLayoutsLock = &sync.RWMutex{}
)

func AddDateTimeLayout(layout string) {
	knownDateTimeLayoutsLock.Lock()
	defer knownDateTimeLayoutsLock.Unlock()
	knownDateTimeLayouts = append(knownDateTimeLayouts, layout)
}

func ParseDateTime(value string) (t time.Time, err error) {
	if value != "" {
		knownDateTimeLayoutsLock.RLock()
		defer knownDateTimeLayoutsLock.RUnlock()
		for _, format := range knownDateTimeLayouts {
			if v, e := time.Parse(format, value); e == nil {
				t = v
				return
			}
		}
	}
	err = fmt.Errorf("failed to parse")
	return
}