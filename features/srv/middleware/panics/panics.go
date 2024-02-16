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

package panics

import (
	"net/http"
	"runtime"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/signals"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "srv-middleware-panics"

type Feature interface {
	feature.Feature
	feature.PanicHandler
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.CFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) PanicHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 1<<16)
				n := runtime.Stack(buf, false)
				buf = buf[:n]
				log.ErrorRF(r, "recovering from panic: %v\n(begin stacktrace)\n%s\n(end stacktrace)", err, buf)
				defer func() {
					if ee := recover(); ee != nil {
						f.Enjin.Emit(signals.EnjinSecondaryPanicRecovery, feature.EnjinTag.String(), w, r, err, ee)
						log.ErrorRF(r, "recovering from secondary panic (without stacktrace)")
						f.Enjin.Serve500(w, r)
					}
				}()
				f.Enjin.Emit(signals.EnjinPanicRecovery, feature.EnjinTag.String(), w, r, err)
				f.Enjin.ServeInternalServerError(w, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
