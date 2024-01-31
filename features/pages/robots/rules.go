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

	"github.com/go-corelibs/slices"
	"github.com/go-enjin/be/pkg/log"
)

type RuleGroup interface {
	String() string
}

type cRuleGroup struct {
	userAgents []string
	allowed    []string
	disallowed []string
}

type MakeRuleGroup interface {
	AddUserAgent(userAgent string) MakeRuleGroup
	AddAllowed(allow string) MakeRuleGroup
	AddDisallowed(disallow string) MakeRuleGroup

	Make() (r RuleGroup)
}

func NewRuleGroup() (rule MakeRuleGroup) {
	r := new(cRuleGroup)
	rule = r
	return
}

func (r *cRuleGroup) AddUserAgent(userAgent string) MakeRuleGroup {
	userAgent = strings.TrimSpace(userAgent)
	if !slices.Within(userAgent, r.userAgents) {
		r.userAgents = append(r.userAgents, userAgent)
	}
	return r
}

func (r *cRuleGroup) AddAllowed(allow string) MakeRuleGroup {
	allow = strings.TrimSpace(allow)
	if !slices.Within(allow, r.allowed) {
		r.allowed = append(r.allowed, allow)
	}
	return r
}

func (r *cRuleGroup) AddDisallowed(disallow string) MakeRuleGroup {
	disallow = strings.TrimSpace(disallow)
	if !slices.Within(disallow, r.disallowed) {
		r.disallowed = append(r.disallowed, disallow)
	}
	return r
}

func (r *cRuleGroup) Make() RuleGroup {
	if len(r.userAgents) == 0 {
		log.FatalDF(1, "at least one user-agent is required per robots.txt rule group")
	}
	if len(r.allowed) == 0 && len(r.disallowed) == 0 {
		log.FatalDF(1, "at least one allowed and/or disallowed is required per robots.txt rule group")
	}
	return r
}

func (r *cRuleGroup) String() (grouped string) {
	for _, userAgent := range r.userAgents {
		grouped += "User-Agent: " + userAgent + "\n"
	}
	if len(r.allowed) > 0 {
		for _, allow := range r.allowed {
			grouped += "Allow: " + allow + "\n"
		}
	}
	if len(r.disallowed) > 0 {
		for _, disallow := range r.disallowed {
			grouped += "Disallow: " + disallow + "\n"
		}
	}
	return
}
