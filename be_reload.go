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

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/signals"
)

func (e *Enjin) HotReloading() (enabled bool) {
	enabled = e.eb.hotReload
	return
}

func (e *Enjin) performHotReload(r *http.Request) (err error) {
	if e.eb.hotReload {
		log.DebugRF(r, "hot-reloading locales")
		e.ReloadLocales()
		e.Emit(signals.PreHotReloadFeatures, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
		for _, f := range feature.FilterTyped[feature.HotReloadableFeature](e.eb.features.List()) {
			log.DebugRF(r, "hot-reloading %v feature", f.Tag())
			if err = f.HotReload(); err != nil {
				err = fmt.Errorf("error hot-reloading feature: %v - %v", f.Tag(), err)
				return
			}
		}
		e.Emit(signals.PostHotReloadFeatures, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
	}
	return
}

func (e *Enjin) hotReloadMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := e.performHotReload(r); err != nil {
			log.ErrorF("error performing hot-reload: %v", err)
		}
		next.ServeHTTP(w, r)
	})
}
