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

package fs

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type QueryFileSystem interface {
	FindPathsWithContextKey(path, key string) (found []string, err error)
	FindPathsWhereContextKeyEquals(path, key string, value interface{}) (found []string, err error)
	FindPathsWhereContextEquals(path string, conditions map[string]interface{}) (found []string, err error)
	FindPathsWhereContext(path string, orJsonConditions ...map[string]interface{}) (found []string, err error)
}

type GormFileSystem interface {
	GormTx() (tx *gorm.DB)
	FindPathsWhere(path string, expressions ...clause.Expression) (found []string, err error)
}