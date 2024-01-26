//go:build driver_fs_local || drivers_fs || locals || all

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
	"path/filepath"
	"runtime"

	"github.com/go-enjin/be/features/fs/locale"
	"github.com/go-enjin/be/pkg/feature"
)

func MakeLocalLocales() (f feature.Feature) {
	_, fn, _, _ := runtime.Caller(0)
	path, _ := filepath.Abs(filepath.Join(filepath.Dir(fn), "locales"))
	f = locale.NewTagged(feature.EnjinLocalesTag).
		MountLocalPath("/", path).
		Make()
	return
}
