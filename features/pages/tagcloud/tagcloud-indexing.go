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

package tagcloud

import (
	"strings"

	"github.com/go-corelibs/slices"
	"github.com/go-enjin/be/pkg/feature/signaling"
	"github.com/go-enjin/be/pkg/signals"
)

func (f *CFeature) addPageIndexingFn(signal signaling.Signal, _ string, _ []interface{}, argv []interface{}) (_ bool) {
	if signal == signals.ContentAddIndexing {
		if _, _, _, p, ok := signals.UnpackContentIndexing(argv); ok {
			if p.Context().Bool("NoTagIndexing", false) || p.Context().Bool("NoPageIndexing", false) {
				return
			} else if len(f.ignoreTypes) > 0 && slices.Within(p.Type(), f.ignoreTypes) {
				return
			} else if len(f.archetypes) > 0 && !slices.Within(p.Archetype(), f.archetypes) {
				return
			} else if v := p.Context().String(f.indexKey, ""); v != "" {
				var keywords []string
				for _, keyword := range strings.Split(strings.ToLower(v), " ") {
					if keyword != "" {
						keywords = append(keywords, keyword)
					}
				}
				if len(keywords) > 0 {
					f.cache.Add(p.Shasum(), keywords...)
				}
			}
		}
	}
	return
}

func (f *CFeature) removePageIndexingFn(signal signaling.Signal, _ string, _ []interface{}, argv []interface{}) (_ bool) {
	if signal == signals.ContentRemoveIndexing {
		if _, _, _, p, ok := signals.UnpackContentIndexing(argv); ok {
			if v := p.Context().String(f.indexKey, ""); v != "" {
				var keywords []string
				for _, keyword := range strings.Split(strings.ToLower(v), " ") {
					if keyword != "" {
						keywords = append(keywords, keyword)
					}
				}
				if len(keywords) > 0 {
					f.cache.Remove(p.Shasum(), keywords...)
				}
			}
		}
	}
	return
}
