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

package feature

import (
	"fmt"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/values"
)

func AsTyped[T interface{}](f Feature) (t T, ok bool) {
	t, ok = f.This().(T)
	return
}

func MustTyped[T interface{}](f Feature) (t T) {
	if v, ok := f.This().(T); ok {
		t = v
		return
	}
	var check *T
	log.FatalDF(1, "%v feature is not %v", f.Tag(), values.TypeOf(check)[1:])
	return
}

func FirstTyped[T interface{}](list Features) (found T) {
	for _, f := range list {
		if fT, ok := f.This().(T); ok {
			found = fT
			return
		}
	}
	return
}

func FilterTyped[T interface{}](list Features) (found []T) {
	for _, f := range list {
		if fT, ok := f.This().(T); ok {
			found = append(found, fT)
		}
	}
	return
}

func GetTyped[T interface{}](tag Tag, list Features) (f T, err error) {
	if found := list.Get(tag); found == nil {
		err = fmt.Errorf("%q feature not found", tag.String())
	} else if sup, ok := found.This().(T); ok {
		f = sup
	} else {
		var t *T
		err = fmt.Errorf("%q is not a %T", tag.String(), values.TypeOf(t)[1:])
	}
	return
}