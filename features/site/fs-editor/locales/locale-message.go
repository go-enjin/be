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

package locales

import (
	"maps"

	"github.com/go-enjin/be/pkg/hash/sha"
	"github.com/go-enjin/be/pkg/lang/catalog"
)

type Select struct {
	Arg     string            `json:"arg"`
	Feature string            `json:"feature"`
	Cases   map[string]string `json:"cases"`
}

type LocaleTranslation struct {
	String string  `json:"string"`
	Select *Select `json:"select"`
}

type LocaleMessage struct {
	ID                string               `json:"id"`
	Key               string               `json:"key"`
	Message           string               `json:"message"`
	Translation       *LocaleTranslation   `json:"translation"`
	TranslatorComment string               `json:"translatorComment,omitempty"`
	Placeholders      catalog.Placeholders `json:"placeholders,omitempty"`
	Fuzzy             bool                 `json:"fuzzy,omitempty"`
	Shasum            string               `json:"shasum"`
}

func (l *LocaleMessage) Copy() (cloned *LocaleMessage) {
	var plural *Select
	if l.Translation.Select != nil {
		plural = &Select{
			Arg:     l.Translation.Select.Arg,
			Feature: "plural",
			Cases:   make(map[string]string),
		}
		maps.Copy(plural.Cases, l.Translation.Select.Cases)
	}
	cloned = &LocaleMessage{
		ID:      l.ID,
		Key:     l.Key,
		Message: l.Message,
		Translation: &LocaleTranslation{
			String: l.Translation.String,
			Select: plural,
		},
		TranslatorComment: l.TranslatorComment,
		Placeholders:      l.Placeholders[:],
		Fuzzy:             l.Fuzzy,
		Shasum:            l.Shasum,
	}
	return
}

func ParseNewMessage(key, comment string) (m *LocaleMessage) {
	replaced, labelled, placeholders := catalog.ParseMessagePlaceholders(key)
	shasum, _ := sha.DataHash10(key)
	m = &LocaleMessage{
		ID:                labelled,
		Key:               key,
		Message:           replaced,
		Translation:       &LocaleTranslation{String: replaced},
		TranslatorComment: comment,
		Placeholders:      placeholders,
		Fuzzy:             true,
		Shasum:            shasum,
	}
	return
}