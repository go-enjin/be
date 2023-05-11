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

package globals

import "github.com/iancoleman/strcase"

func MakeEnvKey(name string) (key string) {
	key = name
	if EnvPrefix != "" {
		key = EnvPrefix + "_" + name
	}
	key = strcase.ToScreamingSnake(key)
	return
}

func MakeEnvKeys(names ...string) (keys []string) {
	for _, name := range names {
		keys = append(keys, MakeEnvKey(name))
	}
	return
}

func MakeFlagName(tag, name string) (actual string) {
	actual = strcase.ToKebab(tag + "-" + name)
	return
}

func MakeFlagEnvKey(tag, name string) (actual string) {
	actual = MakeEnvKey(strcase.ToScreamingSnake(MakeFlagName(tag, name)))
	return
}

func MakeFlagEnvKeys(tag, name string) (actual []string) {
	actual = MakeEnvKeys(strcase.ToScreamingSnake(MakeFlagName(tag, name)))
	return
}