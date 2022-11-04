//go:build page_robots || pages || all

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

package robots

import (
	"strings"

	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

type RuleGroup struct {
	userAgent  string
	allowed    []string
	disallowed []string
	sitemaps   []string
}

type MakeRuleGroup interface {
	AddAllowed(allow string) MakeRuleGroup
	AddDisallowed(disallow string) MakeRuleGroup
	AddSitemap(sitemap string) MakeRuleGroup

	Make() (r *RuleGroup)
}

func NewRuleGroup(userAgent string) (rule MakeRuleGroup) {
	r := new(RuleGroup)
	if userAgent == "" {
		r.userAgent = "*"
	} else {
		r.userAgent = userAgent
	}
	rule = r
	return
}

func (r *RuleGroup) AddAllowed(allow string) MakeRuleGroup {
	allow = strings.TrimSpace(allow)
	if !beStrings.StringInSlices(allow, r.allowed) {
		r.allowed = append(r.allowed, allow)
	}
	return r
}

func (r *RuleGroup) AddDisallowed(disallow string) MakeRuleGroup {
	disallow = strings.TrimSpace(disallow)
	if !beStrings.StringInSlices(disallow, r.disallowed) {
		r.disallowed = append(r.disallowed, disallow)
	}
	return r
}

func (r *RuleGroup) AddSitemap(sitemap string) MakeRuleGroup {
	sitemap = strings.TrimSpace(sitemap)
	if !beStrings.StringInSlices(sitemap, r.sitemaps) {
		r.sitemaps = append(r.sitemaps, sitemap)
	}
	return r
}

func (r *RuleGroup) Make() *RuleGroup {
	if len(r.allowed) == 0 && len(r.disallowed) == 0 && len(r.sitemaps) == 0 {
		log.FatalF("at least one allowed, disallowed or sitemap is required per robots.txt rule group")
	}
	return r
}

func (r *RuleGroup) String() (grouped string) {
	hasAllowed, hasDisallowed, hasSitemaps := len(r.allowed) > 0, len(r.disallowed) > 0, len(r.sitemaps) > 0
	grouped += "User-Agent: " + r.userAgent + "\n\n"
	if hasAllowed {
		for _, allow := range r.allowed {
			grouped += "Allow: " + allow + "\n"
		}
		if hasDisallowed || hasSitemaps {
			grouped += "\n"
		}
	}
	if hasDisallowed {
		for _, disallow := range r.disallowed {
			grouped += "Disallow: " + disallow + "\n"
		}
		if hasSitemaps {
			grouped += "\n"
		}
	}
	if hasSitemaps {
		for _, sitemap := range r.sitemaps {
			grouped += "Sitemap: " + sitemap + "\n"
		}
	}
	return
}