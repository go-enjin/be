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
)

func (f *DBFileSystem) FindPathsWithContext(path, key string, value interface{}) (found []string, err error) {
	f.RLock()
	defer f.RUnlock()

	realpath := strings.TrimPrefix(path, "/")
	// realpath := f.realpath(path)
	var results []*entryStub

	if err = f.tx().
		Where(`path LIKE ?`, realpath+"%").
		Where(datatypes.JSONQuery("context").Equals(value, key)).
		Find(&results).Error; err != nil {
		err = fmt.Errorf("error querying for path LIKE %v AND context.%q == %q", path+"%", key, value)
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