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

package theme

import (
	"bytes"
	"sort"
	"strings"

	"github.com/maruel/natural"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/types/page/matter"
)

func (t *CTheme) ListArchetypes() (names []string) {
	if parent := t.GetParent(); parent != nil {
		names = parent.ListArchetypes()
	}
	if paths, err := t.ThemeFS().ListFiles("archetypes"); err == nil {
		for _, file := range paths {
			if cleaned := strings.TrimPrefix(file, "archetypes/"); !strings.Contains(cleaned, "/") {
				names = append(names, cleaned)
			}
		}
	}
	sort.Sort(natural.StringSlice(names))
	return
}

func (t *CTheme) MakeArchetype(enjin feature.Internals, name string) (format string, data []byte, err error) {
	var paths []string
	if paths, err = t.ThemeFS().ListFiles("archetypes"); err == nil {
		var names []string
		for _, file := range paths {
			names = append(names, strings.TrimPrefix(file, "archetypes/"))
		}
		sort.Sort(natural.StringSlice(names))
		var source string
		for _, file := range names {
			if file == name {
				source = "archetypes/" + file
				for _, fp := range t.formatProviders {
					if frmt, match := fp.MatchFormat(file); frmt != nil {
						format = match
						break
					}
				}
				if format != "" {
					break
				}
			}
		}
		if source != "" && format != "" {
			if d, e := t.ThemeFS().ReadFile(source); e == nil {
				parsedMatter, parsedContent, parsedMatterType := matter.ParseContent(string(d))
				if parsedMatterType == matter.NoneMatter {
					parsedMatterType = matter.TomlMatter
					parsedMatter = `archetype = "` + name + `"`
				}
				ctx := enjin.Context()
				if tmpl, ee := t.NewTextTemplate(enjin, name+"-archetype.tmpl", ctx); ee == nil {
					if tmpl, ee = tmpl.Parse(parsedMatter); ee == nil {
						var w bytes.Buffer
						if ee = tmpl.Execute(&w, ctx); ee == nil {
							if mCtx, eee := matter.UnmarshalFrontMatter(w.Bytes(), parsedMatterType); eee == nil {
								mCtx["archetype"] = name
								stanza := matter.MakeStanza(parsedMatterType, mCtx)
								data = []byte(stanza + "\n" + parsedContent)
								err = nil
								return
							}
						}
					}
				}
			}
		}
	}
	// fallback to parent
	if parent := t.GetParent(); parent != nil {
		format, data, err = parent.MakeArchetype(enjin, name)
	}
	return
}