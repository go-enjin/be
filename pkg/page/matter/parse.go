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

package matter

import (
	"bufio"
	"encoding/json"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"

	"github.com/go-enjin/be/pkg/context"
)

func ParseJson(content string) (m context.Context, err error) {
	m = context.New()
	err = json.Unmarshal([]byte(content), &m)
	return
}

func ParseToml(content string) (m context.Context, err error) {
	m = context.New()
	_, err = toml.Decode(content, &m)
	return
}

func ParseYaml(content string) (m context.Context, err error) {
	m = context.New()
	err = yaml.Unmarshal([]byte(content), m)
	return
}

func ParseFrontMatterContent(raw string) (matter, content string, matterType FrontMatterType) {
	scanner := bufio.NewScanner(strings.NewReader(raw))
	scanner.Split(bufio.ScanLines)

	slurpEOF := func() (lines string) {
		for scanner.Scan() {
			lines += scanner.Text() + "\n"
		}
		return
	}

	slurp := func(until string) (lines string) {
		for scanner.Scan() {
			line := scanner.Text()
			if line == until {
				break
			}
			lines += line + "\n"
		}
		return
	}

	if scanner.Scan() {
		switch scanner.Text() {
		case "+++": // toml
			matter = slurp("+++")
			content = slurpEOF()
			matterType = TomlMatter
			return
		case "---": // yaml
			matter = slurp("---")
			content = slurpEOF()
			matterType = YamlMatter
			return
		case "{{{": // json
			matter = "{\n"
			matter += slurp("}}}")
			matter += "}"
			content = slurpEOF()
			matterType = JsonMatter
			return
		}
	}

	matter = ""
	content = raw
	matterType = NoneMatter
	return
}