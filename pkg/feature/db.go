// Copyright (C) 2023  RunesGambit.com - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package feature

import (
	"database/sql"

	"gorm.io/gorm"

	"github.com/go-corelibs/context"
	"github.com/go-corelibs/go-sqlbuilder"
)

// DatabaseFeature provides enjin supports for one or more Standard Query
// Language database services such as Postgres, MySQL and Sqlite
type DatabaseFeature interface {
	Feature

	// ListDB returns a sorted list of connected db tags for use with DB and
	// MustDB
	ListDB() (tags []string)

	// DB returns the database connection
	DB(tag string) (db DataBase, err error)

	// MustDB returns the database connection, panicking on any error
	MustDB(tag string) (db DataBase)
}

// DataBase is a specific DatabaseFeature connection instance
type DataBase interface {
	// SqlDB returns the underlying [sql.DB] instance
	SqlDB() (db *sql.DB)
	// GormDB returns the [gorm.DB] instance
	GormDB() (db *gorm.DB)
	// Build returns a [sqlbuilder.Buildable]
	// pre-configured with the underlying database's dialect
	Build() (b sqlbuilder.Buildable)

	// Get returns the first `column` value
	Get(column, query string, argv ...interface{}) (value interface{}, err error)
	// GetInt returns the first `column` value, as an int64
	GetInt(column, query string, argv ...interface{}) (value int64, err error)
	// GetInts returns the all `column` values, as an int64 slice
	GetInts(column, query string, argv ...interface{}) (value []int64, err error)
	// GetFloat returns the first `column` value, as a float64
	GetFloat(column, query string, argv ...interface{}) (value float64, err error)
	// GetFloats returns the all `column` values, as a float64 slice
	GetFloats(column, query string, argv ...interface{}) (value []float64, err error)
	// GetString returns the first `column` value, as a string
	GetString(column, query string, argv ...interface{}) (value string, err error)
	// GetStrings returns the all `column` values, as a string slice
	GetStrings(column, query string, argv ...interface{}) (value []string, err error)

	SelectAll(sql string, argv ...interface{}) (rows []context.Context, err error)
	SelectRows(query string, fn func(row context.Context) (stop bool), argv ...interface{}) (err error)
}
