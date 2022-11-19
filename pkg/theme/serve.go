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

package theme

import (
	"fmt"
	"net/http"

	"github.com/go-enjin/be/pkg/log"
	bePath "github.com/go-enjin/be/pkg/path"
)

func (t *Theme) Middleware(next http.Handler) http.Handler {
	log.DebugF("including %v theme static middleware", t.Config.Name)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := fmt.Sprintf("static/%v", bePath.TrimSlashes(r.URL.Path))
		var err error
		var data []byte
		var mime string
		if data, err = t.FileSystem.ReadFile(path); err != nil {
			// log.DebugF("theme statics middleware skip: %v", path)
			next.ServeHTTP(w, r)
			return
		}
		mime, _ = t.FileSystem.MimeType(path)
		w.Header().Set("Content-Type", mime)
		if t.Config.CacheControl == "" {
			w.Header().Set("Cache-Control", DefaultCacheControl)
			// log.WarnF("default cache control: %v", DefaultCacheControl)
		} else {
			w.Header().Set("Cache-Control", t.Config.CacheControl)
			// log.WarnF("custom cache control: %v", t.Config.CacheControl)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
		log.DebugF("%v served: %v (%v)", t.Config.Name, path, mime)
	})
}