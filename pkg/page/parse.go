// Copyright (c) 2022  The Go-Enjin Authors
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

package page

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/gofrs/uuid"
	"gopkg.in/yaml.v3"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/log"
)

type FrontMatterType string

const (
	TomlMatter FrontMatterType = "toml"
	JsonMatter FrontMatterType = "json"
	YamlMatter FrontMatterType = "yaml"
	NoneMatter FrontMatterType = "none"
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

func (p *Page) parseContext(ctx context.Context) {
	ctx.CamelizeKeys()
	ctx.DeleteKeys("Path", "Section", "Slug", "Content")

	p.Type = ctx.String("Type", "page")

	p.SetSlugUrl(ctx.String("Url", p.Url))

	if ctxLang := ctx.String("Language", ""); ctxLang != "" {
		if tag, err := language.Parse(ctxLang); err == nil {
			p.SetLanguage(tag)
		} else {
			p.SetLanguage(language.Und)
		}
	} else {
		p.SetLanguage(language.Und)
	}

	if ctxTranslates := ctx.String("Translates", ""); ctxTranslates != "" {
		p.Translates = ctxTranslates
	}

	if ctxPermalinkId := ctx.String("Permalink", ""); ctxPermalinkId != "" {
		if id, e := uuid.FromString(ctxPermalinkId); e != nil {
			log.ErrorF("error parsing permalink id: %v - %v", ctxPermalinkId, e)
		} else if sum, ee := sha.DataHash10([]byte(ctxPermalinkId)); ee != nil {
			log.ErrorF("error getting permalink sha: %v - %v", id, ee)
		} else {
			p.Permalink = id
			p.PermalinkSha = sum
			ctx.SetSpecific("Permalink", id)
			ctx.SetSpecific("PermalinkSha", sum)
			ctx.SetSpecific("Permalinked", true)
			ctx.SetSpecific("PermalinkUrl", fmt.Sprintf("%v-%v", p.Url, p.PermalinkSha))
			ctx.SetSpecific("PermalinkLongUrl", fmt.Sprintf("%v-%v", p.Url, p.Permalink))
		}
	} else {
		ctx.SetSpecific("Permalink", p.Permalink)
		ctx.SetSpecific("PermalinkSha", "")
		ctx.SetSpecific("Permalinked", false)
		ctx.SetSpecific("PermalinkUrl", p.Url)
		ctx.SetSpecific("PermalinkLongUrl", p.Url)
	}

	p.Title = ctx.String("Title", p.Title)
	ctx.Set("Title", p.Title)

	p.Summary = ctx.String("Summary", p.Summary)
	ctx.Set("Summary", p.Summary)

	p.Description = ctx.String("Description", p.Description)
	ctx.Set("Description", p.Description)

	p.Layout = ctx.String("Layout", p.Layout)
	ctx.Set("Layout", p.Layout)

	p.Archetype = ctx.String("Archetype", p.Archetype)
	ctx.Set("Archetype", p.Archetype)
	if format := ctx.String("Format", ""); format != "" {
		p.Format = format
	}

	// context content is not "source" content, do not populate "from" context,
	// only set it so that it's current
	// ctx.Set("Content", p.Content)

	p.Context.Apply(ctx)
}