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

package signals

import (
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/signaling"
)

const (
	ContentAddIndexing    signaling.Signal = "enjin-add-page-indexing"
	ContentRemoveIndexing signaling.Signal = "enjin-remove-page-indexing"
)

// UnpackContentIndexing is a signal listener helper for extracting the typed
// arguments passed to the signal handler func
func UnpackContentIndexing(argv []interface{}) (filePath string, theme feature.Theme, stub *feature.PageStub, p feature.Page, ok bool) {
	// filePath string, theme feature.Theme, stub *feature.PageStub, p feature.Page
	if ok = len(argv) == 4; ok {
		if filePath, ok = argv[0].(string); ok {
			if theme, ok = argv[1].(feature.Theme); ok {
				if stub, ok = argv[2].(*feature.PageStub); ok {
					if p, ok = argv[3].(feature.Page); ok {
						return
					}
				}
			}
		}
	}
	return
}
