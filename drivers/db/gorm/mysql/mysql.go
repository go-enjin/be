//go:build driver_db_gorm_mysql || drivers_db || drivers || mysql || all

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

package mysql

import (
	"fmt"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

const Tag feature.Tag = "drivers-db-gorm-mysql"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const (
	gFlagFormat string = string(Tag) + "-%v-uri"
)

type Feature interface {
	feature.Database
}

type CFeature struct {
	feature.CFeature

	flags map[string]string
	conns map[string]*gorm.DB
}

type MakeFeature interface {
	Make() Feature

	// AddConnection adds a new connection tag to the enjin and provides a
	// specific command-line connection URI flag
	//
	// Example - if the tag is "your-tag" then the flag is:
	//
	//   --db-gorm-mysql-your-tag-uri
	//
	// the environment variable is:
	//
	//   DB_GORM_MYSQL_YOUR_TAG_URI
	//
	// Note: given tag is always converted to lower-kebab-case format
	AddConnection(tag string) MakeFeature
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = tag
	return f
}

func (f *CFeature) AddConnection(tag string) MakeFeature {
	tag = strcase.ToKebab(tag)
	if _, exists := f.flags[tag]; exists {
		log.FatalDF(1, "a gorm mysql connection already exists with the given tag: %v", tag)
	}
	f.flags[tag] = fmt.Sprintf(gFlagFormat, tag)
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.flags = make(map[string]string)
	f.conns = make(map[string]*gorm.DB)
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	log.DebugDF(1, "building db gorm mysql feature")
	for tag, flag := range f.flags {
		b.AddFlags(
			&cli.StringFlag{
				Name:    flag,
				Usage:   fmt.Sprintf("db gorm mysql %v connection string", tag),
				EnvVars: b.MakeEnvKeys(strcase.ToScreamingSnake(flag)),
			},
		)
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	config := logger.Config{
		SlowThreshold:             time.Second,   // Slow SQL threshold
		LogLevel:                  logger.Silent, // Log level
		IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound
		Colorful:                  false,         // Disable color
	}
	for tag, flag := range f.flags {
		var uri string
		if uri = ctx.String(flag); uri == "" {
			err = fmt.Errorf("db gorm mysql startup error: missing --%v", flag)
			return
		}
		if f.conns[tag], err = gorm.Open(mysql.Open(uri), &gorm.Config{Logger: logger.New(log.Logger(), config)}); err != nil {
			err = fmt.Errorf("db gorm mysql startup error: %v - %v", tag, err)
			return
		}
		log.InfoF("db gorm mysql connected: %v", tag)
	}
	return
}

func (f *CFeature) Shutdown() {
	for tag, conn := range f.conns {
		if db, err := conn.DB(); err != nil {
			log.DebugF("error getting db gorm mysql: %v - %v", tag, err)
		} else {
			if err = db.Close(); err != nil {
				log.ErrorF("error closing db gorm mysql: %v - %v", tag, err)
			} else {
				log.InfoF("closed db gorm mysql: %v", tag)
			}
		}
	}
}

func (f *CFeature) ListDB() (tags []string) {
	tags = maps.SortedKeys(f.conns)
	return
}

func (f *CFeature) DB(tag string) (db interface{}, err error) {
	if v, ok := f.conns[tag]; ok {
		db = v
		return
	}
	err = fmt.Errorf("db gorm mysql connection %v not found", tag)
	return
}

func (f *CFeature) MustDB(tag string) (db interface{}) {
	if v, ok := f.conns[tag]; ok {
		db = v
		return
	}
	log.FatalDF(1, "db gorm mysql connection %v not found", tag)
	return
}