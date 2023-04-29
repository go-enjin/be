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

package database

import (
	"net/http"

	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
)

const RequestDefaultDbKey context.RequestKey = "DefaultDb"

func GetRequestDefaultDB(r *http.Request) (db interface{}) {
	db = r.Context().Value(RequestDefaultDbKey)
	return
}

func GetRequestDB(tag string, r *http.Request) (db interface{}) {
	key := MakeRequestDbKey(tag)
	db = r.Context().Value(key)
	return
}

func GetRequestSpecificDB(fTag feature.Tag, tag string, r *http.Request) (db interface{}) {
	_, spec := MakeRequestDbSpecificKeys(fTag, tag)
	db = r.Context().Value(spec)
	return
}

func MakeRequestDbKey(tag string) (key context.RequestKey) {
	key = context.RequestKey(strcase.ToCamel(tag))
	return
}

func MakeRequestDbSpecificKeys(fTag feature.Tag, tag string) (tagKey, specificKey context.RequestKey) {
	fKebab := strcase.ToKebab(fTag.String())
	kebab := strcase.ToKebab(tag)
	tagKey = context.RequestKey(strcase.ToCamel(tag))
	specificKey = context.RequestKey(strcase.ToCamel(fKebab + "-" + kebab))
	return
}