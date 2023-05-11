//go:build fs_drivers_local || fs_drivers || all

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

package mountable

import (
	"github.com/go-enjin/be/pkg/fs/drivers/local"
	"github.com/go-enjin/be/pkg/log"
)

type LocalPathSupport[MakeTypedFeature interface{}] interface {
	MountLocalPath(mount, path string) MakeTypedFeature
}

func (f *CFeature[MakeTypedFeature]) MountLocalPath(mount, path string) MakeTypedFeature {
	if lfs, err := local.New(path); err != nil {
		log.FatalDF(1, "error mounting path: %v", err)
	} else {
		f.MountPathROFS(path, mount, lfs)
	}
	v, _ := f.This().(MakeTypedFeature)
	return v
}