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

package feature

import (
	"fmt"

	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

type FeaturesCache struct {
	list Features
	tags map[Tag]Feature
}

func NewFeaturesCache() (cache *FeaturesCache) {
	cache = &FeaturesCache{
		list: make(Features, 0),
		tags: make(map[Tag]Feature),
	}
	return
}

func (c *FeaturesCache) checkDupe(tag Tag) (err error) {
	if _, present := c.tags[tag]; present {
		err = fmt.Errorf("duplicate feature.Tag: %v", tag)
		return
	}
	return
}

func (c *FeaturesCache) Prepend(f Feature) (err error) {
	tag := f.Tag()
	if err = c.checkDupe(tag); err != nil {
		return
	}
	c.list = append(Features{f}, c.list...)
	c.tags[tag] = f
	return
}

func (c *FeaturesCache) Add(f Feature) (err error) {
	tag := f.Tag()
	if err = c.checkDupe(tag); err != nil {
		return
	}
	c.list = append(c.list, f)
	c.tags[tag] = f
	return
}

func (c *FeaturesCache) Tags() (list Tags) {
	list = maps.TypedKeys(c.tags)
	return
}

func (c *FeaturesCache) List() (list Features) {
	list = append(list, c.list...)
	return
}

func (c *FeaturesCache) Get(tag Tag) (f Feature, ok bool) {
	f, ok = c.tags[tag]
	return
}

func (c *FeaturesCache) MustGet(tag Tag) (f Feature) {
	if v, ok := c.tags[tag]; ok {
		f = v
		return
	}
	log.FatalDF(1, "%v feature not found", tag)
	return
}