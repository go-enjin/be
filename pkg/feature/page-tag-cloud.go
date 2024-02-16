// Copyright (c) 2024  The Go-Enjin Authors
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

package feature

import (
	"sort"
	"time"

	"github.com/maruel/natural"
)

type TagCloudProvider interface {
	Feature
	GetTagCloud() (tc TagCloud)
	GetPageTagCloud(shasum string) (tags TagCloud)
}

type CloudTag struct {
	Word   string
	Link   string
	Count  int
	Weight int
}

type CloudTagPage struct {
	Title   string
	Url     string
	Created time.Time
	Updated time.Time
	Tags    TagCloud
}

type TagCloud []*CloudTag

func (c TagCloud) Update(baseurl string) {
	if baseurl != "" {
		for _, tag := range c {
			tag.Link = baseurl + "/:" + tag.Word
		}
	}
	// weight is a scale from 1 to 10 where 1 is the least count and 10 is the most count
	var most int
	for _, tag := range c {
		if most < tag.Count {
			most = tag.Count
		}
	}
	for _, tag := range c {
		switch tag.Count {
		case 1:
			tag.Weight = 1
		default:
			tag.Weight = int(float64(tag.Count) / float64(most) * 10.0)
		}
	}
}

func (c TagCloud) Rank() {
	sort.Slice(c, func(i, j int) (less bool) {
		less = c[i].Count > c[j].Count
		return
	})
}

func (c TagCloud) Sort() {
	sort.Slice(c, func(i, j int) (less bool) {
		less = natural.Less(c[i].Word, c[j].Word)
		return
	})
}
