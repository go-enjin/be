//go:build !driver_fs_db_gorm && !drivers_fs_db && !drivers_fs && !dbs && !all

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

package filesystem

import (
	"github.com/urfave/cli/v2"
)

type GormDBPathSupport[MakeTypedFeature interface{}] interface {
}

type CGormDBPathSupport[MakeTypedFeature interface{}] struct {
}

func (s CGormDBPathSupport[MakeTypedFeature]) initGormDBPathSupport(f *CFeature[MakeTypedFeature]) (err error) {
	return
}

func (s CGormDBPathSupport[MakeTypedFeature]) startupGormDBPathSupport(f *CFeature[MakeTypedFeature], ctx *cli.Context) (err error) {
	return
}

func (s CGormDBPathSupport[MakeTypedFeature]) handleFindPageMatterPathGormDBPathSupport(f *CFeature[MakeTypedFeature], ctx *cli.Context) (handled bool, path string, err error) {
	return
}
