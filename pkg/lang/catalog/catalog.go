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

package catalog

import (
	"fmt"

	"github.com/goccy/go-json"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/go-enjin/golang-org-x-text/message/catalog"

	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
)

var (
	_ Catalog = (*CCatalog)(nil)
)

var (
	OutputGoTextName   = "out.gotext.json"
	MessagesGoTextName = "messages.gotext.json"
)

type Catalog interface {
	AddCatalog(others ...catalog.Catalog)
	AddLocalesFromFS(defaultTag language.Tag, efs fs.FileSystem)
	AddLocalesFromJsonBytes(tag language.Tag, src string, contents []byte)

	LocaleTags() (tags []language.Tag)
	LocaleTagsWithDefault(d language.Tag) (tags []language.Tag)

	MakeGoTextCatalog() (gtc catalog.Catalog, err error)
}

type CCatalog struct {
	table    map[language.Tag]*dictionaries
	catalogs []catalog.Catalog
}

func New() (c Catalog) {
	c = &CCatalog{
		table:    make(map[language.Tag]*dictionaries),
		catalogs: make([]catalog.Catalog, 0),
	}
	return
}

func (c *CCatalog) AddCatalog(others ...catalog.Catalog) {
	c.catalogs = append(c.catalogs, others...)
}

func (c *CCatalog) AddLocalesFromJsonBytes(tag language.Tag, src string, contents []byte) {
	var data map[string]interface{}
	if err := json.Unmarshal(contents, &data); err != nil {
		log.ErrorF("error parsing json locale: [%v] %v - %v", tag, src, err)
	} else {
		if v, ok := data["messages"]; ok && v == nil {
			log.TraceDF(1, "skipping nil messages: [%v] %v", tag, src)
			return
		}
		if _, ok := c.table[tag]; !ok {
			c.table[tag] = newDictionaries(tag)
		}
		d := newDictionaryFromJsonData(tag, src, data)
		c.table[tag].Append(d)
	}
}

func (c *CCatalog) AddLocalesFromFS(defaultTag language.Tag, efs fs.FileSystem) {
	if entries, err := efs.ReadDir("."); err != nil {
		log.ErrorF("error read dir: %v", err)
	} else {
		for _, entry := range entries {
			name := entry.Name()
			if entryTag, ee := language.Parse(name); ee != nil {
				log.ErrorF("invalid language: %v", entry.Name())
			} else {
				if tagEntries, eee := efs.ReadDir(name); eee != nil {
					log.ErrorF("error read dir %v: %v", name, eee)
				} else {
					var filename string
					for _, te := range tagEntries {
						switch te.Name() {
						case OutputGoTextName:
							filename = OutputGoTextName
							break
						}
					}

					if filename == "" {
						if language.Compare(entryTag, defaultTag) {
							log.DebugF("locale (%v) not found, expected: %v", name, OutputGoTextName)
						} else {
							log.DebugF("locale (%v) not found, expected: %v", name, MessagesGoTextName)
						}
						continue
					}

					src := name + "/" + filename
					log.DebugF("locale source found: %v", src)
					if contents, eeee := efs.ReadFile(src); eeee != nil {
						log.ErrorF("error reading: %v - %v", src, eeee)
					} else {
						c.AddLocalesFromJsonBytes(entryTag, src, contents)
					}
				}
			}
		}
	}
}

func (c *CCatalog) LocaleTags() (tags []language.Tag) {
	for tag, _ := range c.table {
		tags = append(tags, tag)
	}
	tags = lang.SortLanguageTags(tags)
	return
}

func (c *CCatalog) LocaleTagsWithDefault(d language.Tag) (tags []language.Tag) {
	for tag, _ := range c.table {
		tags = append(tags, tag)
	}
	tags = lang.SortLanguageTags(tags)
	found := -1
	for idx, tag := range tags {
		if language.Compare(d, tag) {
			found = idx
			break
		}
	}
	if found > -1 {
		tag := tags[found]
		tags = append(tags[:found], tags[found+1:]...)
		tags = append([]language.Tag{tag}, tags...)
	}
	return
}

func (c *CCatalog) MakeGoTextCatalog() (gtc catalog.Catalog, err error) {
	//var opts []catalog.Option
	//if len(c.catalogs) > 0 {
	//	opts = append(opts, catalog.Include(c.catalogs...))
	//}
	//b := catalog.NewBuilder(opts...)
	b := catalog.NewBuilder()
	b.Include(c.catalogs...)
	for tag, dict := range c.table {
		for _, d := range dict.list {
			for k, v := range d.key {
				if err = b.SetString(tag, k, v); err != nil {
					err = fmt.Errorf("error setting message string: [%v] %v: %v", tag, k, v)
					return
				}
			}
		}
	}
	gtc = b
	return
}