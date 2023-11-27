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

import (
	"regexp"
	"strings"
)

// inspired by https://github.com/byn9826/words-count/blob/master/src/globalWordsCount.js

var (
	DefaultPunctuation = []rune{
		',', '，', '.', '。', ':', '：', ';', '；', '[', ']', '【', ']', '】', '{', '｛', '}', '｝',
		'(', '（', ')', '）', '<', '《', '>', '》', '$', '￥', '!', '！', '?', '？', '~', '～',
		'\'', '’', '"', '“', '”',
		'*', '/', '\\', '&', '%', '@', '#', '^', '、', '、', '、', '、',
	}

	RxSpaces   = regexp.MustCompile(`\s+`)
	RxSymbols  = regexp.MustCompile(`\p{S}`)
	RxCharSets = regexp.MustCompile(`[\p{Han}\p{Katakana}\p{Hiragana}\p{Hangul}]`)
)

type Config struct {
	PunctuationAsBreaker      bool
	DisableDefaultPunctuation bool
	Punctuation               []rune
}

func DefaultConfig() (cfg *Config) {
	return &Config{}
}

func (cfg *Config) List(input string) (list []string) {
	var work, spacer string
	if work = strings.TrimSpace(input); work == "" {
		return
	} else if cfg.PunctuationAsBreaker {
		spacer = " "
	}

	var combined []rune
	if !cfg.DisableDefaultPunctuation {
		combined = append(combined, DefaultPunctuation...)
	}
	if len(cfg.Punctuation) > 0 {
		combined = append(combined, cfg.Punctuation...)
	}

	// replace all punctuation with spacers
	if len(combined) > 0 {
		for _, r := range combined {
			work = strings.ReplaceAll(work, string(r), spacer)
		}
	}

	work = RxSymbols.ReplaceAllString(work, "")
	work = RxSpaces.ReplaceAllString(work, " ")
	work = strings.TrimSpace(work)

	list = strings.Split(work, " ")
	return
}

func (cfg *Config) Range(input string, fn func(word string)) {
	list := cfg.List(input)

	for _, word := range list {
		var carry []string
		if RxCharSets.MatchString(word) {
			for _, r := range word {
				if char := string(r); RxCharSets.MatchString(char) {
					carry = append(carry, char)
				}
			}
		}
		if carried := len(carry); carried == 0 {
			fn(word)
		} else {
			for _, w := range carry {
				fn(w)
			}
		}
	}

}

func (cfg *Config) Count(input string) (count int) {
	cfg.Range(input, func(_ string) {
		count += 1
	})
	return
}

func (cfg *Config) Parse(input string) (words []string) {
	cfg.Range(input, func(word string) {
		words = append(words, word)
	})
	return
}

func (cfg *Config) Search(query, content string) (score int, found []string) {
	keywords := cfg.Parse(strings.ToLower(query))
	keywordCount := len(keywords)
	haystack := cfg.Parse(strings.ToLower(content))
	unique := map[string]struct{}{}

	for _, word := range haystack {
		for idx, keyword := range keywords {
			if word == keyword {
				weight := keywordCount - idx
				score += weight
				if _, present := unique[word]; !present {
					found = append(found, word)
				}
			}
		}
	}

	return
}
