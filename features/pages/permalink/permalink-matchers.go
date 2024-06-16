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

package permalink

import (
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

func (f *CFeature) _parsePath(path string) (permalink string, short, ok bool) {
	if permalink, short, ok = ParsePermalink(path); ok {
		if short {
			log.TraceDF(1, "found permalink slug: %v - %v", path, permalink)
		} else {
			log.TraceDF(1, "found permalink root: %v - %v", path, permalink)
		}
	}
	return
}

func (f *CFeature) _permalink(r *http.Request, input interface{}) (url string, err error) {
	var permalink uuid.UUID
	if vs, ok := input.(string); ok {
		if permalink, err = uuid.FromString(vs); err != nil {
			return
		}
	} else if vu, ok := input.(uuid.UUID); ok {
		permalink = vu
	} else {
		err = fmt.Errorf("expected uuid.UUID or string; received %T", input)
		return
	}
	if permalink != uuid.Nil {
		url = "/" + permalink.String()
		for _, tag := range f.Enjin.SiteLocales() {
			if f.Enjin.SiteSupportsLanguage(tag) {
				if p := f.Enjin.FindPage(r, tag, url); p != nil {
					url = p.Url() + "-" + p.PermalinkSha()
					return
				}
			}
		}
	}
	return
}

func (f *CFeature) _permalinkMatcher(path string, p feature.Page) (found string, ok bool) {
	found = path
	if p.Permalink() != uuid.Nil {
		if parsed, short, valid := f._parsePath(path); valid {
			if short {
				// brief shasum
				ok = parsed == p.PermalinkSha()
			} else {
				// e0f7ae8b-85e0-4c3f-b6c7-4c84b59bd3e7
				if parsedUuid := uuid.FromStringOrNil(parsed); parsedUuid != uuid.Nil {
					ok = parsedUuid.String() == parsed
				}
			}
		}
	}
	return
}
