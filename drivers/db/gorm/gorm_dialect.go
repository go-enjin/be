//go:build driver_db_gorm || drivers_db || gorm || all

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

package gorm

import "gorm.io/gorm"

var (
	gKnownDialects = make(map[string]*gormDialect)
)

type gormDialect struct {
	dbType string
	openFn func(dsn string) gorm.Dialector
}
