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

package be

import (
	"fmt"
	"net/http"

	"github.com/go-enjin/be/pkg/log"
)

func (e *Enjin) performHotReload() (err error) {
	t := e.MustGetTheme()
	if e.eb.hotReload {
		log.DebugF("hot-reloading theme: %v", t.Name())
		if err = t.Reload(); err != nil {
			err = fmt.Errorf("error reloading theme: %v", err)
			return
		}
		log.DebugF("hot-reloading locales")
		e.reloadLocales()
	}
	return
}

func (e *Enjin) hotReloadMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := e.performHotReload(); err != nil {
			log.ErrorF("error performing hot-reload: %v", err)
		}
		next.ServeHTTP(w, r)
	})
}