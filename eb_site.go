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

package be

import (
	"strings"

	"github.com/go-enjin/be/pkg/feature"
)

func (eb *EnjinBuilder) SiteTag(tag string) feature.Builder {
	eb.tag = strings.ToLower(tag)
	return eb
}

func (eb *EnjinBuilder) SiteName(name string) feature.Builder {
	eb.name = name
	return eb
}

func (eb *EnjinBuilder) SiteTagLine(tagLine string) feature.Builder {
	eb.tagLine = tagLine
	return eb
}

func (eb *EnjinBuilder) SiteCopyrightName(name string) feature.Builder {
	eb.copyrightName = name
	return eb
}

func (eb *EnjinBuilder) SiteCopyrightNotice(notice string) feature.Builder {
	eb.copyrightNotice = notice
	return eb
}
