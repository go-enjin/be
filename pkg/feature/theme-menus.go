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

type MenuSupports []MenuSupport

type MenuSupport struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Area string `json:"area"`
}

func ParseMenuSupports(input interface{}) (supports MenuSupports) {

	process := func(vm map[string]interface{}) {
		var present bool
		var support = MenuSupport{}
		if support.Key, present = vm["key"].(string); !present {
			return
		} else if support.Name, present = vm["name"].(string); !present {
			return
		} else if support.Area, present = vm["area"].(string); !present {
			return
		}
		supports = append(supports, support)
		return
	}

	switch t := input.(type) {
	case []interface{}:
		for _, v := range t {
			if vm, ok := v.(map[string]interface{}); ok {
				process(vm)
			}
		}

	case []map[string]interface{}:
		for _, vm := range t {
			process(vm)
		}
	}

	return
}

func (m MenuSupports) Append(others MenuSupports) (merged MenuSupports) {
	unique := map[string]struct{}{}
	for _, s := range m {
		unique[s.Key] = struct{}{}
	}
	merged = append(merged, m...)
	for _, other := range others {
		if _, exists := unique[other.Key]; exists {
			continue
		}
		merged = append(merged, other)
	}
	return
}

func (m MenuSupports) Keys() (keys []string) {
	for _, s := range m {
		keys = append(keys, s.Key)
	}
	return
}

func (m MenuSupports) Has(key string) (supported bool) {
	for _, s := range m {
		if supported = s.Key == key; supported {
			return
		}
	}
	return
}