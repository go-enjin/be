//go:build database || all

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

package database

import (
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/database"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

var (
	DefaultDatabaseType = "sqlite3"
	DefaultDatabaseUri  = "db.sqlite"
)

var _ feature.Feature = (*Feature)(nil)

const Tag feature.Tag = "Database"

type Feature struct {
	feature.CFeature

	dbType string
	dbUri  string
}

type MakeFeature interface {
	feature.MakeFeature

	DatabaseType(dbType string) MakeFeature
	DatabaseUri(dbUri string) MakeFeature
}

func New() MakeFeature {
	f := new(Feature)
	f.Init(f)
	return f
}

func (f *Feature) DatabaseType(dbType string) MakeFeature {
	f.dbType = dbType
	return f
}

func (f *Feature) DatabaseUri(dbUri string) MakeFeature {
	f.dbUri = dbUri
	return f
}

func (f *Feature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.dbType = DefaultDatabaseType
	f.dbUri = DefaultDatabaseUri
}

func (f *Feature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *Feature) Build(b feature.Buildable) (err error) {
	log.DebugDF(1, "building database feature")
	b.AddFlags(
		&cli.StringFlag{
			Name:    "db-type",
			Usage:   "the type of database to use",
			Value:   DefaultDatabaseType,
			Aliases: []string{"t"},
			EnvVars: b.MakeEnvKeys("DB_TYPE"),
		},
		&cli.StringFlag{
			Name:    "db-uri",
			Usage:   "the database connection string to use",
			Value:   DefaultDatabaseUri,
			Aliases: []string{"d"},
			EnvVars: b.MakeEnvKeys("DB_URI"),
		},
	)
	return
}

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	if ctx.IsSet("db-type") {
		f.dbType = ctx.String("db-type")
	}
	if ctx.IsSet("db-uri") {
		f.dbUri = ctx.String("db-uri")
	}
	if err = database.Connect(f.dbType, f.dbUri); err != nil {
		return
	}
	log.InfoF("%v database connected", f.dbType)
	return
}

func (f *Feature) Shutdown() {
	if gdb, err := database.Get(); err == nil {
		if db, err := gdb.DB(); err == nil {
			_ = db.Close()
			log.InfoF("database connection closed")
		}
	}
}