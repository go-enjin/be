//go:build driver_kvs_gorm || drivers_kvs || all

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

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"gorm.io/gorm"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/gob"
	"github.com/go-enjin/be/pkg/log"
)

const Tag feature.Tag = "drivers-kvs-gorm"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	feature.KeyValueStoreAny
}

type MakeFeature interface {
	Make() Feature

	SetTableName(name string) MakeFeature
	SetDatabaseConnection(name string) MakeFeature
}

type CFeature struct {
	feature.CFeature

	db        *gorm.DB
	dbcName   string
	tableName string
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.dbcName = "default"
}

func (f *CFeature) SetTableName(name string) MakeFeature {
	f.tableName = name
	return f
}

func (f *CFeature) SetDatabaseConnection(name string) MakeFeature {
	f.dbcName = name
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	if f.tableName == "" {
		f.tableName = strcase.ToSnake(f.Tag().String())
	}
	if f.dbcName == "" {
		err = fmt.Errorf("%v requires a database connection name set", f.Tag())
		return
	}
	if dbi, ee := f.Enjin.DB(f.dbcName); ee != nil {
		err = fmt.Errorf("error getting database connection by name: %v", f.dbcName)
		return
	} else if db, ok := dbi.(*gorm.DB); !ok {
		err = fmt.Errorf("error getting database connection: %v, expected *gorm.DB, received: %T", f.dbcName, dbi)
	} else {
		f.db = db.Scopes(func(tx *gorm.DB) *gorm.DB {
			tx.Table(f.tableName)
			return tx
		})
		err = f.db.AutoMigrate(&Entry{})
	}
	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) Get(key interface{}) (value interface{}, ok bool) {
	log.DebugF("getting: %T - %#+v", key, key)
	var err error
	var keyGob []byte
	if keyGob, err = gob.Encode(key); err != nil {
		log.ErrorDF(1, "error encoding key data: %v", err)
		return
	}
	var entry Entry
	if err = f.db.Where(`key = ?`, keyGob).First(&entry).Error; err != nil {
		log.ErrorDF(1, "error retrieving key value: %v", err)
		return
	}
	if value, err = gob.Decode(entry.Value); err != nil {
		log.ErrorDF(1, "error decoding value data: %v", err)
	} else {
		ok = true
	}
	return
}

func (f *CFeature) Set(key, value interface{}) {
	log.DebugF("setting: %T - %#+v => %T - %#+v", key, key, value, value)
	var err error
	var keyGob, valueGob []byte
	if keyGob, err = gob.Encode(key); err != nil {
		log.FatalDF(1, "error encoding key data: %v", err)
	}
	if valueGob, err = gob.Encode(value); err != nil {
		log.FatalDF(1, "error encoding value data: %v", err)
	}
	entry := &Entry{Key: keyGob, Value: valueGob}
	if err = f.db.Save(entry).Error; err != nil {
		log.FatalDF(1, "error saving key-value entry: %v", err)
	}
	return
}
