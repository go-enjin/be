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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"

	"github.com/go-enjin/be/pkg/context"
)

func UnmarshalFrontMatter(data []byte, matterType FrontMatterType) (matter context.Context, err error) {
	matter = context.Context{}
	switch matterType {
	case TomlMatter:
		err = toml.Unmarshal(data, &matter)
	case YamlMatter:
		err = yaml.Unmarshal(data, &matter)
	case JsonMatter:
		err = json.Unmarshal(data, &matter)
	default:
		err = fmt.Errorf("unsupported front-matter type: %v", matterType)
	}
	return
}

func ParseContent(raw string) (matter, content string, matterType FrontMatterType) {
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

func MakeStanza(fmt FrontMatterType, ctx context.Context) (stanza string) {
	switch fmt {
	case JsonMatter:
		stanza += "{{{\n"
		if data, err := json.MarshalIndent(ctx, "", "\t"); err == nil {
			content := string(data)
			if size := len(content); size >= 2 {
				if content[size-1] == '}' {
					content = content[:size-1]
				}
				if content[0] == '{' {
					content = content[1:]
				}
			}
			stanza += content + "\n"
		}
		stanza += "}}}"
	case YamlMatter:
		stanza += "---\n"
		if data, err := yaml.Marshal(ctx); err == nil {
			content := string(data)
			if len(content) > 4 {
				if content[:4] == "---\n" {
					content = content[4:]
				}
			}
			stanza += content + "\n"
		}
		stanza += "---"
	case TomlMatter:
		fallthrough
	default:
		stanza += "+++\n"
		var buf bytes.Buffer
		enc := toml.NewEncoder(&buf)
		if err := enc.Encode(ctx); err == nil {
			content := buf.String()
			if len(content) > 4 {
				if content[:4] == "---\n" {
					content = content[4:]
				}
			}
			stanza += content + "\n"
		}
		stanza += "+++"
	}
	return
}