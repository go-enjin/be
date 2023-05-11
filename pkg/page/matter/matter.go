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
	"encoding/gob"
	"fmt"
	"time"

	"github.com/go-enjin/golang-org-x-text/language"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/lang"
)

func init() {
	gob.Register(&PageMatter{})
}

type FrontMatterType string

const (
	TomlMatter FrontMatterType = "toml"
	JsonMatter FrontMatterType = "json"
	YamlMatter FrontMatterType = "yaml"
	NoneMatter FrontMatterType = "none"
)

type PageMatter struct {
	Path   string
	Shasum string

	Body   string
	Matter beContext.Context
	Locale language.Tag

	Created time.Time
	Updated time.Time

	FrontMatter     string
	FrontMatterType FrontMatterType

	Stub *PageStub
}

func ParsePageMatter(path string, created, updated time.Time, raw []byte) (pm *PageMatter, err error) {
	var shasum string
	if shasum, err = sha.DataHash10(raw); err != nil {
		err = fmt.Errorf("error hashing page data: %v", err)
		return
	}

	cleaned := lang.StripTranslatorComments(string(raw))

	var ctx beContext.Context
	matter, content, matterType := ParseFrontMatterContent(cleaned)
	switch matterType {
	case JsonMatter:
		if ctx, err = ParseJson(matter); err != nil {
			err = fmt.Errorf("error parsing JSON front-matter: %v", err)
			return
		}
	case TomlMatter:
		if ctx, err = ParseToml(matter); err != nil {
			err = fmt.Errorf("error parsing TOML front-matter: %v", err)
			return
		}
	case YamlMatter:
		if ctx, err = ParseYaml(matter); err != nil {
			err = fmt.Errorf("error parsing YAML front-matter: %v", err)
			return
		}
	default:
		ctx = beContext.New()
	}
	if v := ctx.String("Created", ""); v != "" {
		if t, e := ParseDateTime(v); e == nil {
			created = t
		}
	}
	if v := ctx.String("Updated", ""); v != "" {
		if t, e := ParseDateTime(v); e == nil {
			updated = t
		}
	}
	pm = &PageMatter{
		Path:            path,
		Shasum:          shasum,
		Body:            content,
		Matter:          ctx,
		Created:         created,
		Updated:         updated,
		FrontMatter:     matter,
		FrontMatterType: matterType,
	}
	return
}

// Bytes rebuilds the page matter's file data, overriding FrontMatter with the
// Matter content, in the FrontMatterType format
func (pm *PageMatter) Bytes() (data []byte, err error) {
	var matter []byte
	switch pm.FrontMatterType {
	case JsonMatter:
		data = append(data, "{{{\n"...)
		if matter, err = pm.Matter.AsJSON(); err != nil {
			return
		}
		data = append(data, matter...)
		data = append(data, "}}}\n"...)
	case TomlMatter:
		data = append(data, "+++\n"...)
		if matter, err = pm.Matter.AsTOML(); err != nil {
			return
		}
		data = append(data, matter...)
		data = append(data, "+++\n"...)
	case YamlMatter:
		data = append(data, "---\n"...)
		if matter, err = pm.Matter.AsYAML(); err != nil {
			return
		}
		data = append(data, matter...)
		data = append(data, "---\n"...)
	}
	data = append(data, pm.Body...)
	return
}