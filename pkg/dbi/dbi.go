// Copyright (c) 2024  The Go-Enjin Authors
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

// Package dbi is an implementation of feature.DataBase
package dbi

import (
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/go-corelibs/context"
	"github.com/go-corelibs/go-sqlbuilder"
	"github.com/go-corelibs/go-sqlbuilder/dialects"
	clValues "github.com/go-corelibs/values"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

var (
	_ feature.DataBase = (*database)(nil)
)

var (
	ErrNotFound     = errors.New("not found")
	ErrNotSupported = errors.New("not supported")
)

type database struct {
	sql   *sql.DB
	gorm  *gorm.DB
	build sqlbuilder.Buildable
}

func New(db *gorm.DB) (dbi feature.DataBase, err error) {
	var d sqlbuilder.Dialect
	if parsed, ok := dialects.Parse(db.Name()); ok {
		d = parsed
	} else {
		err = fmt.Errorf("%q %w", db.Name(), ErrNotSupported)
		return
	}
	var sdb *sql.DB
	if sdb, err = db.DB(); err != nil {
		return
	}
	dbi = &database{
		sql:   sdb,
		gorm:  db,
		build: sqlbuilder.NewBuildable(d),
	}
	return
}

func (d *database) SqlDB() (db *sql.DB) {
	return d.sql
}

func (d *database) GormDB() (db *gorm.DB) {
	return d.gorm
}

func (d *database) Build() (b sqlbuilder.Buildable) {
	return d.build
}

func (d *database) Get(column, sql string, argv ...interface{}) (value interface{}, err error) {
	tx := d.gorm.Raw(sql, argv...)
	if err = tx.Error; err != nil {
		return
	}
	data := make([]map[string]interface{}, 0)
	if err = tx.Scan(&data).Error; err != nil {
		return
	}
	if len(data) > 0 {
		value, _ = data[0][column]
	} else {
		err = ErrNotFound
	}
	return
}

func (d *database) GetInt(column, sql string, argv ...interface{}) (value int64, err error) {
	var ok bool
	var v interface{}
	if v, err = d.Get(column, sql, argv...); err != nil {
		return
	} else if value, ok = clValues.RecastValue[int64](v); !ok {
		err = fmt.Errorf("expected int64, received: %T", v)
	}
	return
}

func (d *database) GetInts(column, sql string, argv ...interface{}) (values []int64, err error) {
	tx := d.gorm.Raw(sql, argv...)
	if err = tx.Error; err != nil {
		return
	}
	data := make([]map[string]interface{}, 0)
	if err = tx.Scan(&data).Error; err != nil {
		return
	}
	for _, item := range data {
		if value, ok := clValues.RecastValue[int64](item[column]); ok {
			values = append(values, value)
		} else {
			log.ErrorDF(1, "%q expected int64, received: %T", column, item[column])
		}
	}
	return
}

func (d *database) GetFloat(column, sql string, argv ...interface{}) (value float64, err error) {
	var ok bool
	var v interface{}
	if v, err = d.Get(column, sql, argv...); err != nil {
		return
	} else if value, ok = clValues.RecastValue[float64](v); !ok {
		err = fmt.Errorf("expected float64, received: %T", v)
	}
	return
}

func (d *database) GetFloats(column, sql string, argv ...interface{}) (values []float64, err error) {
	tx := d.gorm.Raw(sql, argv...)
	if err = tx.Error; err != nil {
		return
	}
	data := make([]map[string]interface{}, 0)
	if err = tx.Scan(&data).Error; err != nil {
		return
	}
	for _, item := range data {
		if value, ok := clValues.RecastValue[float64](item[column]); ok {
			values = append(values, value)
		} else {
			log.ErrorDF(1, "%q expected float64, received: %T", column, item[column])
		}
	}
	return
}

func (d *database) GetString(column, sql string, argv ...interface{}) (value string, err error) {
	var ok bool
	var v interface{}
	if v, err = d.Get(column, sql, argv...); err != nil {
		return
	} else if value, ok = clValues.RecastValue[string](v); !ok {
		err = fmt.Errorf("expected string, received: %T", v)
	}
	return
}

func (d *database) GetStrings(column, sql string, argv ...interface{}) (values []string, err error) {
	tx := d.gorm.Raw(sql, argv...)
	if err = tx.Error; err != nil {
		return
	}
	data := make([]map[string]interface{}, 0)
	if err = tx.Scan(&data).Error; err != nil {
		return
	}
	for _, item := range data {
		if value, ok := clValues.RecastValue[string](item[column]); ok {
			values = append(values, value)
		} else {
			log.ErrorDF(1, "%q expected string, received: %T", column, item[column])
		}
	}
	return
}

func (d *database) SelectAll(sql string, argv ...interface{}) (rows []context.Context, err error) {
	tx := d.gorm.Raw(sql, argv...)
	if err = tx.Error; err != nil {
		return
	}
	rows = make([]context.Context, 0)
	err = tx.Scan(&rows).Error
	return
}

func (d *database) SelectRows(query string, fn func(row context.Context) (stop bool), argv ...interface{}) (err error) {
	tx := d.gorm.Raw(query, argv...)
	if err = tx.Error; err != nil {
		return
	}
	var rows *sql.Rows
	if rows, err = tx.Rows(); err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		row := context.New()
		if err = rows.Scan(&row); err != nil {
			return
		} else if stop := fn(row); stop {
			return
		}
	}
	return
}
