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

package auth

import (
	"fmt"

	"github.com/go-enjin/be/pkg/forms"
)

func (f *CFeature) loadEnvironment() (err error) {

	if named, ok := f.env.GetSiteEnviron("secret-keys"); ok {
		for name, value := range named {
			f.secretKeys[name] = []byte(forms.StrictSanitize(value))
			if len(f.secretKeys[name]) < 64 {
				err = fmt.Errorf("%q secret key has less than 64 bytes", name)
				return
			}
		}
	}

	if named, ok := f.env.GetSiteEnviron("audience-keys"); ok {
		for audience, value := range named {
			f.audienceKeys[audience] = []byte(forms.StrictSanitize(value))
			if len(f.audienceKeys[audience]) < 32 {
				err = fmt.Errorf("%q audience key has less than 32 bytes", audience)
				return
			}
		}
	}

	return
}