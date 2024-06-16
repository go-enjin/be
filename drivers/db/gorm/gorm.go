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

	"github.com/go-corelibs/maps"
	"github.com/go-enjin/be/pkg/dbi"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

const Tag feature.Tag = "drivers-db-gorm"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.DatabaseFeature
}

type MakeFeature interface {
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

	// AddConnectionWith is like AddConnection with specifying the underlying
	// sql.DB and logger.Config settings via a buildable Config instance used
	// when opening Gorm instances during startup
	AddConnectionWith(tag string, c Config) MakeFeature

	// SetPreset configures a new connection with a default dialect and URI
	//
	// Note: cli flags and environment variables override any presets
	SetPreset(tag, dialect, uri string) MakeFeature

	// SetPresetWith is like AddConnectionWith but with presets
	SetPresetWith(tag, dialect, uri string, c Config) MakeFeature

	// SetConfig specifies the db config for the connection tag
	SetConfig(tag string, c Config) MakeFeature

	// SetLogging configures the gorm.Config.Logger setting for the given
	// connection, the default is to not log anything for all connections
	SetLogging(connection string, level logger.LogLevel, slowThreshold time.Duration, ignoreRecordNotFound, parameterizedQueries bool) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	flags  map[string][]string
	cfg    map[string]*config
	preset map[string]*preset
	dbh    map[string]feature.DataBase
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
	f.cfg = make(map[string]*config)
	f.preset = make(map[string]*preset)
	f.dbh = make(map[string]feature.DataBase)
}

func (f *CFeature) setConn(depth int, tag string, c Config) {
	tag = strcase.ToKebab(tag)
	if _, exists := f.flags[tag]; exists {
		log.FatalDF(depth+1, "connection exists already: %q", tag)
	}
	f.flags[tag] = append(f.flags[tag],
		fmt.Sprintf("%v-%v-type", f.Tag().String(), tag),
		fmt.Sprintf("%v-%v-uri", f.Tag().String(), tag),
	)
	f.cfg[tag] = c.make()
}

func (f *CFeature) setPreset(depth int, tag, dialect, uri string) (kebab string) {
	kebab = strcase.ToKebab(tag)
	f.setConn(depth+1, kebab, NewConfig())
	if _, present := f.preset[kebab]; present {
		log.FatalDF(depth+1, "preset exists already: %q", kebab)
	} else if !gDialects.has(DialectName(dialect)) {
		log.FatalDF(depth+1, "unknown dialect: %q", dialect)
	}
	f.preset[kebab] = &preset{
		tag:  kebab,
		name: DialectName(dialect),
		uri:  uri,
	}
	return
}

func (f *CFeature) AddConnection(tag string) MakeFeature {
	f.setConn(1, tag, NewConfig())
	return f
}

func (f *CFeature) AddConnectionWith(tag string, c Config) MakeFeature {
	f.setConn(1, tag, c)
	return f
}

func (f *CFeature) SetPreset(tag, dialect, uri string) MakeFeature {
	f.setPreset(1, tag, dialect, uri)
	return f
}

func (f *CFeature) SetPresetWith(tag, dialect, uri string, c Config) MakeFeature {
	kebab := f.setPreset(1, tag, dialect, uri)
	f.cfg[kebab] = c.make()
	return f
}

func (f *CFeature) SetConfig(tag string, c Config) MakeFeature {
	kebab := strcase.ToKebab(tag)
	if _, present := f.flags[kebab]; !present {
		log.FatalDF(1, "unknown connection: %q", kebab)
	}
	f.cfg[kebab] = c.make()
	return f
}

func (f *CFeature) SetLogging(connection string, level logger.LogLevel, slowThreshold time.Duration, ignoreRecordNotFound, parameterizedQueries bool) MakeFeature {
	kebab := strcase.ToKebab(connection)
	if cfg, present := f.cfg[kebab]; !present {
		log.FatalDF(1, "%q connection not found", kebab)
	} else {
		cfg.logLevel = level
		cfg.slowThreshold = slowThreshold
		cfg.ignoreRecordNotFoundError = ignoreRecordNotFound
		cfg.parameterizedQueries = parameterizedQueries
	}
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	log.DebugDF(1, "building database feature")
	fTag := f.Tag().String()
	for tag, flags := range f.flags {
		b.AddFlags(
			&cli.StringFlag{
				Category: fTag,
				Name:     flags[0],
				Usage:    fmt.Sprintf("%v connection type (supported: %v)", tag, gDialects.supports()),
				EnvVars:  b.MakeEnvKeys(strcase.ToScreamingSnake(flags[0])),
			},
			&cli.StringFlag{
				Category: fTag,
				Name:     flags[1],
				Usage:    fmt.Sprintf("%v connection URI", tag),
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
		var dbUri string
		var dbType DialectName
		var dialect *dbDialect
		dbTypeFlag, dbUriFlag := flags[0], flags[1]

		if p, ok := f.preset[tag]; ok {
			// preset present, cli flags are optional
			dbUri = p.uri
			dbType = p.name
			dialect = gDialects.get(p.name)

			if ctx.IsSet(dbUriFlag) {
				if uri := ctx.String(dbUriFlag); uri != "" {
					dbUri = uri
				}
			}
			if ctx.IsSet(dbTypeFlag) {
				if name := ctx.Generic(dbTypeFlag).(DialectName); name != "" {
					dbType = name
				}
			}

		} else {
			// no preset, all cli flags are required
			if dbUri = ctx.String(dbUriFlag); dbUri == "" {
				err = fmt.Errorf("database startup error: %v - missing --%v", tag, dbUriFlag)
				return
			} else if dbType = DialectName(ctx.String(dbTypeFlag)); dbType == "" {
				err = fmt.Errorf("database startup error: --%v is missing", dbTypeFlag)
				return
			} else if dialect = gDialects.get(dbType); dialect == nil {
				err = fmt.Errorf("database startup error: unknown type --%v=%q", dbTypeFlag, dbType)
				return
			}
		}

		var ok bool
		var cfg *config
		var gormConfig = &gorm.Config{}
		if cfg, ok = f.cfg[tag]; ok {
			gormConfig.Logger = logger.New(log.PrefixedLogger("(db|"+string(dbType)+"|"+tag+") - "), cfg.loggerConfig())
		}

		// ensure that loc, parseTime and charset are set for the specific dialect
		if dbUri, err = dbi.UpdateURI(dbUri, dialect.dbUriParams()); err != nil {
			return
		}

		var db *gorm.DB
		if db, err = gorm.Open(dialect.openFn(dbUri), gormConfig); err != nil {
			err = fmt.Errorf("database startup connection error: %v - %v", tag, err)
			return
		}

		if f.dbh[tag], err = dbi.New(db); err != nil {
			err = fmt.Errorf("dbi startup error: %v", err)
			return
		}

		log.InfoF("connected: %v - %v", tag, dbType)
	}
	return
}

func (f *CFeature) Shutdown() {
	for tag, dbh := range f.dbh {
		if db := dbh.SqlDB(); db != nil {
			if err := db.Close(); err != nil {
				log.ErrorF("error closing database: %v - %v", tag, err)
			} else {
				log.InfoF("closed database: %v", tag)
			}
		}
	}
}

func (f *CFeature) ListDB() (tags []string) {
	tags = maps.SortedKeys(f.dbh)
	return
}

func (f *CFeature) DB(tag string) (db feature.DataBase, err error) {
	if v, ok := f.dbh[tag]; ok {
		db = v
		return
	}
	err = ErrConnectionNotFound
	return
}

func (f *CFeature) MustDB(tag string) (db feature.DataBase) {
	var err error
	if db, err = f.DB(tag); err == nil {
		return
	}
	log.PanicDF(1, f.mkErr("MustDB", tag, err).Error())
	return
}
