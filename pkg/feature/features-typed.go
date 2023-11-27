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

type TypedFeatures[T interface{}] []T

func (tf TypedFeatures[T]) Len() (count int) {
	return len(tf)
}

func (tf TypedFeatures[T]) Tags() (tags Tags) {
	for _, v := range tf {
		if f, ok := interface{}(v).(Feature); ok {
			tags = append(tags, f.Tag())
		}
	}
	return
}

func (tf TypedFeatures[T]) Has(tag Tag) (present bool) {
	for _, v := range tf {
		if f, ok := interface{}(v).(Feature); ok {
			if present = f.Tag() == tag; present {
				return
			}
		}
	}
	return
}

func (tf TypedFeatures[T]) Find(name string) (tag Tag) {
	for _, v := range tf {
		if f, ok := interface{}(v).(Feature); ok {
			t := f.Tag()
			if t.String() == name {
				tag = t
				return
			} else if t.Kebab() == name {
				tag = t
				return
			}
		}
	}
	return
}

func (tf TypedFeatures[T]) Get(tag Tag) (found T) {
	if tag.IsNil() {
		return
	}
	for _, v := range tf {
		if f, ok := interface{}(v).(Feature); ok && f.Tag() == tag {
			found, _ = f.This().(T)
			return
		}
	}
	return
}

func (tf TypedFeatures[T]) AsFeatures() (features Features) {
	for _, v := range tf {
		if f, ok := interface{}(v).(Feature); ok {
			features = append(features, f)
		}
	}
	return
}
