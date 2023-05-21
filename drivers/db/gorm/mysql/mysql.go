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

const (
	Dialect             = "mysql"
	Tag     feature.Tag = "drivers-db-gorm-" + Dialect
)

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

	loggerConfig map[string]logger.Config
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

	// SetLogging configures the gorm.Config.Logger setting for the given
	// connection, the default is to not log anything for all connections
	SetLogging(connection string, level logger.LogLevel, slowThreshold time.Duration, ignoreRecordNotFound, parameterizedQueries bool) MakeFeature
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
		log.FatalDF(1, "a gorm %v connection already exists with the given tag: %v", Dialect, tag)
	}
	f.flags[tag] = fmt.Sprintf(gFlagFormat, tag)
	return f
}

func (f *CFeature) SetLogging(connection string, level logger.LogLevel, slowThreshold time.Duration, ignoreRecordNotFound, parameterizedQueries bool) MakeFeature {
	f.loggerConfig[connection] = logger.Config{
		Colorful:                  false, // always disable color
		LogLevel:                  level,
		SlowThreshold:             slowThreshold,
		IgnoreRecordNotFoundError: ignoreRecordNotFound,
		ParameterizedQueries:      parameterizedQueries,
	}
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.flags = make(map[string]string)
	f.conns = make(map[string]*gorm.DB)
	f.loggerConfig = make(map[string]logger.Config)
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	log.DebugDF(1, "building db gorm %v feature", Dialect)
	for tag, flag := range f.flags {
		b.AddFlags(
			&cli.StringFlag{
				Name:    flag,
				Usage:   fmt.Sprintf("db gorm %v %v connection string", Dialect, tag),
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

	for tag, flag := range f.flags {
		var uri string
		if uri = ctx.String(flag); uri == "" {
			err = fmt.Errorf("db gorm %v startup error: missing --%v", Dialect, flag)
			return
		}

		var gormConfig = &gorm.Config{}
		if config, ok := f.loggerConfig[tag]; ok {
			gormConfig.Logger = logger.New(log.PrefixedLogger("(db|"+Dialect+"|"+tag+") - "), config)
		}

		if f.conns[tag], err = gorm.Open(mysql.Open(uri), gormConfig); err != nil {
			err = fmt.Errorf("db gorm %v startup error: %v - %v", Dialect, tag, err)
			return
		}
		log.InfoF("db gorm %v connected: %v", Dialect, tag)
	}
	return
}

func (f *CFeature) Shutdown() {
	for tag, conn := range f.conns {
		if db, err := conn.DB(); err != nil {
			log.DebugF("error getting db gorm %v: %v - %v", Dialect, tag, err)
		} else {
			if err = db.Close(); err != nil {
				log.ErrorF("error closing db gorm %v: %v - %v", Dialect, tag, err)
			} else {
				log.InfoF("closed db gorm %v: %v", Dialect, tag)
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
	err = fmt.Errorf("db gorm %v connection %v not found", Dialect, tag)
	return
}

func (f *CFeature) MustDB(tag string) (db interface{}) {
	if v, ok := f.conns[tag]; ok {
		db = v
		return
	}
	log.FatalDF(1, "db gorm %v connection %v not found", Dialect, tag)
	return
}
