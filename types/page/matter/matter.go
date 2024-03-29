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
	"encoding/json"
	"fmt"
	"time"

	cllang "github.com/go-corelibs/lang"
	"github.com/go-corelibs/x-text/language"

	"github.com/go-enjin/be/pkg/editor"

	clPath "github.com/go-corelibs/path"
	sha "github.com/go-corelibs/shasum"
	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
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
	Origin string

	Path   string
	Shasum string

	Body   string
	Matter beContext.Context
	Locale language.Tag

	Created time.Time
	Updated time.Time

	FrontMatter     string
	FrontMatterType FrontMatterType

	Stub interface{}
}

func NewPageMatter(origin string, path, body string, frontMatterType FrontMatterType, matter beContext.Context) (pm *PageMatter) {
	stanza := MakeStanza(frontMatterType, matter)
	content := stanza + "\n" + body
	now := time.Now()
	var err error
	if pm, err = ParsePageMatter(origin, path, now, now, []byte(content)); err != nil {
		panic(err)
	}
	return
}

func ParsePageMatter(origin string, path string, created, updated time.Time, raw []byte) (pm *PageMatter, err error) {
	var shasum string
	if shasum, err = sha.BriefSum(raw); err != nil {
		err = fmt.Errorf("error hashing page data: %v", err)
		return
	}

	if modified, _, ok := editor.ParseEditorWorkFile(path); ok {
		path = modified
	}
	path = clPath.CleanWithSlash(path)
	cleaned := cllang.PruneTranslatorComments(string(raw))

	var ctx beContext.Context
	matter, content, matterType := ParseContent(cleaned)
	switch matterType {
	case JsonMatter:
		if ctx, err = beContext.ParseJson(matter); err != nil {
			err = fmt.Errorf("error parsing JSON front-matter: %v", err)
			return
		}
	case TomlMatter:
		if ctx, err = beContext.ParseToml(matter); err != nil {
			err = fmt.Errorf("error parsing TOML front-matter: %v", err)
			return
		}
	case YamlMatter:
		if ctx, err = beContext.ParseYaml(matter); err != nil {
			err = fmt.Errorf("error parsing YAML front-matter: %v", err)
			return
		}
	case NoneMatter:
		fallthrough
	default:
		ctx = beContext.New()
	}

	locale := language.Und
	if parsedTag, modifiedPath, ok := lang.ParseLangPath(path); ok {
		path = modifiedPath
		locale = parsedTag
	} else if v := ctx.String("Language", ""); v != "" {
		if parsedLang, ee := language.Parse(v); ee == nil {
			locale = parsedLang
		} else {
			log.ErrorF("error parsing language from page context: %v - %v", v, ee)
			ctx.SetSpecific("Language", "")
		}
	}

	for _, key := range []string{"created", "updated", "deleted"} {
		var t time.Time
		if t = ctx.Time(key, time.Time{}); t.Unix() == 0 {
			if v := ctx.String(key, ""); v != "" {
				var e error
				if t, e = beContext.ParseTimeStructure(v); e != nil {
					log.ErrorF("error parsing page matter .%s timestamp: %v", key, e)
					continue
				}
			}
		}
		if t.Unix() > 0 {
			ctx.SetSpecific(key, t)
			switch key {
			case "created":
				created = t
			case "updated":
				updated = t
			}
		}
	}

	pm = &PageMatter{
		Origin:          origin,
		Path:            path,
		Shasum:          shasum,
		Body:            content,
		Locale:          locale,
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
	case NoneMatter, JsonMatter:
		data = append(data, "{{{\n"...)
		if matter, err = pm.Matter.AsJSON(); err != nil {
			return
		}
		// json front-matter must remove opening and closing braces
		data = append(data, matter[1:len(matter)-1]...)
		data = append(data, "\n}}}\n"...)
	case TomlMatter:
		data = append(data, "+++\n"...)
		if matter, err = pm.Matter.AsTOML(); err != nil {
			return
		}
		data = append(data, matter...)
		data = append(data, "\n+++\n"...)
	case YamlMatter:
		data = append(data, "---\n"...)
		if matter, err = pm.Matter.AsYAML(); err != nil {
			return
		}
		data = append(data, matter...)
		data = append(data, "\n---\n"...)
	}
	data = append(data, pm.Body...)
	return
}

func (pm *PageMatter) DecodeJsonBodyWith(v interface{}) (err error) {
	err = json.Unmarshal([]byte(pm.Body), v)
	return
}

// DecodeJsonBody decodes the .Body from JSON and returns the context data
func (pm *PageMatter) DecodeJsonBody() (data beContext.Context, err error) {
	data = beContext.New()
	err = json.Unmarshal([]byte(pm.Body), &data)
	return
}

func (pm *PageMatter) Copy() (copied *PageMatter) {
	copied = &PageMatter{
		Origin:          pm.Origin,
		Path:            pm.Path,
		Shasum:          pm.Shasum,
		Body:            pm.Body,
		Matter:          pm.Matter.Copy(),
		Locale:          pm.Locale,
		Created:         pm.Created,
		Updated:         pm.Updated,
		FrontMatter:     pm.FrontMatter,
		FrontMatterType: pm.FrontMatterType,
		Stub:            pm.Stub,
	}
	return
}
