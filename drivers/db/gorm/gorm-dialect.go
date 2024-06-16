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

import (
	"gorm.io/gorm"

	"github.com/go-corelibs/go-sqlbuilder"
)

// DialectName is a string type representing the name (or alias) of a supported
// database type
type DialectName string

type dbDialect struct {
	name    DialectName
	alias   []DialectName
	openFn  func(dsn string) gorm.Dialector
	dialect sqlbuilder.Dialect
}

func (d *dbDialect) aliases() (names []string) {
	for _, alias := range d.alias {
		names = append(names, string(alias))
	}
	return
}

func (d *dbDialect) dbUriParams() map[string]string {
	// TODO: figure out if postgres needs something more

	m := map[string]string{
		"loc":       "UTC",
		"parseTime": "true",
		"charset":   "utf8",
	}

	if d.name == "mysql" {
		m["charset"] = "utf8mb4"
	}

	return m
}
