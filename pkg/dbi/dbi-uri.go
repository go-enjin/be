// Copyright (c) 2024  The Go-Enjin Authors
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

package dbi

import (
	"fmt"
	"net/url"
	"strings"
)

func UpdateURI(dbUri string, set map[string]string) (modified string, err error) {
	var found bool
	var parsed *url.URL
	var values url.Values
	var before, after string
	if before, after, found = strings.Cut(dbUri, "?"); found {
		if parsed, err = url.Parse("/?" + after); err != nil {
			err = fmt.Errorf("error parsing db uri query parameters: %w", err)
			return
		}
		values = parsed.Query()
	} else {
		values = url.Values{}
	}

	for key, value := range set {
		values.Set(key, value)
	}

	modified = before + "?" + values.Encode()
	return
}
