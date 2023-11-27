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

package words

func List(input string, cfg *Config) (list []string) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	list = cfg.List(input)
	return
}

func Range(input string, cfg *Config, fn func(word string)) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	cfg.Range(input, fn)
	return
}

func Count(input string, cfg *Config) (count int) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	count = cfg.Count(input)
	return
}

func Parse(input string, cfg *Config) (words []string) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	words = cfg.Parse(input)
	return
}

func Search(query, content string, cfg *Config) (score int, present []string) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	score, present = cfg.Search(query, content)
	return
}
