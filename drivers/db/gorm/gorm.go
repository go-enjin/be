//go:build driver_db_gorm || drivers_db || drivers || all

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
	"fmt"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

const Tag feature.Tag = "drivers-db-gorm"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Database
}

type CFeature struct {
	feature.CFeature

	flags map[string][]string
	conns map[string]*gorm.DB

	loggerConfig map[string]logger.Config
}

type MakeFeature interface {
	Make() Feature

	// AddConnection adds a new connection tag to the enjin and provides a
	// pair of specific command-line connection flags
	//
	// Example - if the tag is "your-tag" then the flag is:
	//
	//   --db-gorm-your-tag-uri
	//   --db-gorm-your-tag-type
	//
	// the environment variable is:
	//
	//   DB_GORM_YOUR_TAG_URI
	//   DB_GORM_YOUR_TAG_TYPE
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
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.CFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.flags = make(map[string][]string)
	f.conns = make(map[string]*gorm.DB)
	f.loggerConfig = make(map[string]logger.Config)
}

func (f *CFeature) AddConnection(tag string) MakeFeature {
	tag = strcase.ToKebab(tag)
	if _, exists := f.flags[tag]; exists {
		log.FatalDF(1, "a gorm connection already exists with the given tag: %v", tag)
	}
	f.flags[tag] = append(f.flags[tag],
		fmt.Sprintf("%v-%v-type", f.Tag().String(), tag),
		fmt.Sprintf("%v-%v-uri", f.Tag().String(), tag),
	)
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

func (f *CFeature) Build(b feature.Buildable) (err error) {
	log.DebugDF(1, "building gorm db feature")
	fTag := f.Tag().String()
	for tag, flags := range f.flags {
		known := maps.SortedKeys(gKnownDialects)
		b.AddFlags(
			&cli.StringFlag{
				Category: fTag,
				Name:     flags[0],
				Usage:    fmt.Sprintf("gorm db %v connection type %v", tag, known),
				EnvVars:  b.MakeEnvKeys(strcase.ToScreamingSnake(flags[0])),
			},
			&cli.StringFlag{
				Category: fTag,
				Name:     flags[1],
				Usage:    fmt.Sprintf("gorm db %v connection URI", tag),
				EnvVars:  b.MakeEnvKeys(strcase.ToScreamingSnake(flags[1])),
			},
		)
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	for tag, flags := range f.flags {
		dbTypeFlag, dbUriFlag := flags[0], flags[1]
		var dialect *gormDialect

		var dbType string
		if dbType = ctx.String(dbTypeFlag); dbType == "" {
			err = fmt.Errorf("gorm db startup error: missing --%v", dbTypeFlag)
			return
		} else {
			var known bool
			if dialect, known = gKnownDialects[dbType]; !known {
				err = fmt.Errorf("gorm db startup error: unknown type --%v=%q", dbTypeFlag, dbType)
				return
			}
		}

		var uri string
		if uri = ctx.String(dbUriFlag); uri == "" {
			err = fmt.Errorf("gorm db startup error: %v - missing --%v", tag, dbUriFlag)
			return
		}

		var gormConfig = &gorm.Config{}
		if config, ok := f.loggerConfig[tag]; ok {
			gormConfig.Logger = logger.New(log.PrefixedLogger("(db|"+dbType+"|"+tag+") - "), config)
		}

		if f.conns[tag], err = gorm.Open(dialect.openFn(uri), gormConfig); err != nil {
			err = fmt.Errorf("gorm db startup connection error: %v - %v", tag, err)
			return
		}

		log.InfoF("connected: %v - %v", tag, dbType)
	}
	return
}

func (f *CFeature) Shutdown() {
	for tag, conn := range f.conns {
		if db, err := conn.DB(); err != nil {
			log.DebugF("error getting gorm db: %v - %v", tag, err)
		} else {
			if err = db.Close(); err != nil {
				log.ErrorF("error closing gorm db: %v - %v", tag, err)
			} else {
				log.InfoF("closed gorm db: %v", tag)
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
	err = fmt.Errorf("gorm db connection %v not found", tag)
	return
}

func (f *CFeature) MustDB(tag string) (db interface{}) {
	if v, ok := f.conns[tag]; ok {
		db = v
		return
	}
	log.PanicDF(1, "gorm db connection %v not found", tag)
	return
}