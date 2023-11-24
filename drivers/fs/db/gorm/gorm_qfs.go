//go:build driver_fs_db_gorm || drivers_fs_db || drivers_fs || drivers || all

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
	"os"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/go-enjin/be/pkg/maps"
)

func (f *DBFileSystem) GormTx() (tx *gorm.DB) {
	return f.tableScopedOrTx()
}

func (f *DBFileSystem) FindPathsWithContextKey(path, key string) (found []string, err error) {
	found, err = f.FindPathsWhere(path, datatypes.JSONQuery("context").HasKey(key))
	return
}

func (f *DBFileSystem) FindPathsWhereContextKeyEquals(path, key string, value interface{}) (found []string, err error) {
	found, err = f.FindPathsWhere(path, datatypes.JSONQuery("context").Equals(value, key))
	return
}

func (f *DBFileSystem) FindPathsWhereContextEquals(path string, conditions map[string]interface{}) (found []string, err error) {
	var expressions []clause.Expression
	for k, v := range conditions {
		expressions = append(expressions, datatypes.JSONQuery("context").Equals(v, k))
	}
	found, err = f.FindPathsWhere(path, expressions...)
	return
}

func (f *DBFileSystem) FindPathsWhereContext(path string, orJsonConditions ...map[string]interface{}) (found []string, err error) {
	var orExpressions []clause.Expression
	for _, andConditions := range orJsonConditions {
		var andExpressions []clause.Expression
		for _, k := range maps.SortedKeys(andConditions) {
			andExpressions = append(
				andExpressions,
				datatypes.JSONQuery("context").
					Equals(andConditions[k], k),
			)
		}
		orExpressions = append(orExpressions, clause.And(andExpressions...))
	}
	found, err = f.FindPathsWhere(path, clause.Or(orExpressions...))
	return
}

func (f *DBFileSystem) FindPathsWhere(path string, expressions ...clause.Expression) (found []string, err error) {
	f.RLock()
	defer f.RUnlock()

	realpath := strings.TrimPrefix(path, "/")
	var results []*entryStub

	query := f.tableScopedOrTx().
		Where(`path LIKE ?`, realpath+"%").
		Where(clause.And(expressions...))
	if err = query.Find(&results).Error; err != nil {
		err = fmt.Errorf("error querying for path LIKE %v AND %#+v", path+"%", query.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx
		}))
		return
	}

	if len(results) > 0 {
		for _, result := range results {
			found = append(found, result.Path)
		}
		return
	}

	err = os.ErrNotExist
	return
}
