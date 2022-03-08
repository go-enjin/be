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
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/go-enjin/be/pkg/log"
)

var (
	Instance *gorm.DB = nil
)

var (
	ErrorMissingInstance = fmt.Errorf("missing database instance")
)

// New returns a newly opened database connection with inherited log settings
func New(dbType, dsn string) (db *gorm.DB, err error) {
	config := logger.Config{
		SlowThreshold:             time.Second,   // Slow SQL threshold
		LogLevel:                  logger.Silent, // Log level
		IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound
		Colorful:                  false,         // Disable color
	}

	return NewWithLoggerConfig(dbType, dsn, config)
}

// NewWithLoggerConfig returns a newly opened database connection using the
// given logger.Config
func NewWithLoggerConfig(dbType, dsn string, config logger.Config) (db *gorm.DB, err error) {

	cfg := &gorm.Config{
		Logger: logger.New(
			log.Logger(), // io writer
			config,
		),
	}

	switch dbType {
	case "sqlite", "sqlite3":
		db, err = gorm.Open(sqlite.Open(dsn), cfg)
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), cfg)
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), cfg)
	default:
		err = fmt.Errorf("unsupported database: %v", dbType)
	}
	return
}

// Connect opens a new database connection and stores it in a package instance
// for further use with the Get and MustGet functions
func Connect(dbType, dsn string) (err error) {
	Instance, err = New(dbType, dsn)
	return
}

// Get returns the package instance connection or a "missing database
// instance" error
func Get() (db *gorm.DB, err error) {
	if Instance == nil {
		err = ErrorMissingInstance
	} else {
		db = Instance
	}
	return
}

// MustGet returns the package instance connection or panics with a log.FatalF call
func MustGet() (db *gorm.DB) {
	var err error
	if db, err = Get(); err != nil {
		log.FatalF("%v", err)
	}
	return db
}