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
	"io/fs"
	"sort"

	"github.com/maruel/natural"

	"github.com/go-enjin/golang-org-x-text/feature/plural"
	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/go-enjin/golang-org-x-text/message/catalog"

	beFs "github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
)

var (
	_ Catalog = (*CCatalog)(nil)
)

var (
	DefaultFileName = "out.gotext.json"
)

type Catalog interface {
	AddCatalog(others ...catalog.Catalog)
	AddLocalesFromFS(defaultTag language.Tag, efs beFs.FileSystem)
	AddLocalesFromJsonBytes(tag language.Tag, src string, contents []byte)

	LocaleTags() (tags []language.Tag)
	LocaleTagsWithDefault(d language.Tag) (tags []language.Tag)

	MakeGoTextCatalog() (gtc catalog.Catalog, err error)
}

type CCatalog struct {
	catalog  *catalog.Builder
	catalogs []catalog.Catalog
}

func New() (c Catalog) {
	c = &CCatalog{
		catalog:  catalog.NewBuilder(),
		catalogs: make([]catalog.Catalog, 0),
	}
	return
}

func (c *CCatalog) AddCatalog(others ...catalog.Catalog) {
	c.catalogs = append(c.catalogs, others...)
}

func (c *CCatalog) AddLocalesFromJsonBytes(tag language.Tag, src string, contents []byte) {
	if len(contents) == 0 {
		return
	}

	var err error
	var gt *GoText
	var parsed language.Tag

	if gt, parsed, err = ParseGoText(contents); err != nil {
		log.ErrorF("error parsing gotext.json: [%v] %v - %v", tag, src, err)
		return
	}

	for _, msg := range gt.Messages {
		if msg.Translation.Select != nil {
			var argNum int = -1
			for _, p := range msg.Placeholders {
				if p.ID == msg.Translation.Select.Arg {
					argNum = p.ArgNum
					break
				}
			}
			if argNum <= -1 {
				log.ErrorF("error correlating placeholder argNum: [%v] %v - %#+v", tag, src, msg)
				continue
			}
			var cases []interface{}
			for k, v := range msg.Translation.Select.Cases {
				cases = append(cases, k, v.Msg)
			}
			if err = c.catalog.Set(parsed, msg.Key, plural.Selectf(argNum, "", cases...)); err != nil {
				log.ErrorF("error setting gotext.json select: [%v] %v - %q - %v", tag, src, msg.Translation.Select, err)
			}
		} else if err = c.catalog.Set(parsed, msg.Key, catalog.String(msg.Translation.String)); err != nil {
			log.ErrorF("error setting gotext.json string: [%v] %v - %q - %v", tag, src, msg.Translation.String, err)
		}
	}

}

func (c *CCatalog) AddLocalesFromFS(defaultTag language.Tag, efs beFs.FileSystem) {
	var err error
	var entries []fs.DirEntry
	if entries, err = efs.ReadDir("."); err != nil {
		log.ErrorF("error read dir: %v", err)
		return
	}

	sort.Slice(entries, func(i, j int) (less bool) {
		less = natural.Less(entries[i].Name(), entries[j].Name())
		return
	})

	for _, entry := range entries {
		name := entry.Name()
		var entryTag language.Tag
		if entryTag, err = language.Parse(name); err != nil {
			log.ErrorF("invalid language: %v", entry.Name())
			continue
		}

		var tagEntries []fs.DirEntry
		if tagEntries, err = efs.ReadDir(entry.Name()); err != nil {
			log.ErrorF("error read dir %v: %v", name, err)
			continue
		}

		var filename string
		for _, te := range tagEntries {
			switch te.Name() {
			case DefaultFileName:
				filename = DefaultFileName
				break
			}
		}
		if filename == "" {
			log.DebugF("locale (%v) not found, expected: %v", name, DefaultFileName)
			continue
		}

		src := name + "/" + filename
		//log.DebugF("locale source found: %v", src)
		if contents, eeee := efs.ReadFile(src); eeee != nil {
			log.ErrorF("error reading: %v - %v", src, eeee)
		} else {
			c.AddLocalesFromJsonBytes(entryTag, src, contents)
		}

	}
}

func (c *CCatalog) LocaleTags() (tags []language.Tag) {
	tags = c.catalog.Languages()
	tags = lang.SortLanguageTags(tags)
	return
}

func (c *CCatalog) LocaleTagsWithDefault(d language.Tag) (tags []language.Tag) {
	tags = c.catalog.Languages()
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
	b := catalog.NewBuilder()
	b.Include(c.catalogs...)
	b.Include(c.catalog)
	gtc = b
	return
}
