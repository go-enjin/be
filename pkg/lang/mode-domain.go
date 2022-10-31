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

package lang

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/log"
)

var _ Mode = (*DomainMode)(nil)

type DomainMode struct {
	domains map[language.Tag]*url.URL
}

type DomainModeBuilder interface {
	Set(tag language.Tag, domain string) DomainModeBuilder

	Make() Mode
}

func NewDomainMode() (p DomainModeBuilder) {
	p = &DomainMode{
		domains: make(map[language.Tag]*url.URL),
	}
	return
}

func (p *DomainMode) Set(tag language.Tag, domain string) DomainModeBuilder {
	if u, e := url.ParseRequestURI(domain); e == nil {
		p.domains[tag] = u
		log.DebugF("set domain name: [%v] %v", tag, p.domains[tag])
	} else {
		log.FatalF("invalid domain name given: [%v] %v - %v", tag, domain, e)
	}
	return p
}

func (p *DomainMode) Make() Mode {
	return p
}

func (p *DomainMode) ToUrl(defaultTag, tag language.Tag, path string) (translated string) {

	translated = path

	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	if domain, ok := p.domains[tag]; ok {
		translated = fmt.Sprintf("%v/%v", domain.String(), path)
		// log.DebugF("found [%v] domain: %v - %v", tag, domain, path)
	} else if defaultDomain, ok := p.domains[defaultTag]; ok {
		log.ErrorF("%v language domain not found, using default: %v", tag, defaultTag)
		translated = fmt.Sprintf("%v/%v", defaultDomain.String(), path)
	} else {
		log.ErrorF("%v and %v (default) language domains not found", tag, defaultTag)
	}

	// log.DebugF("translated domain url: %v", translated)
	return
}

func (p *DomainMode) FromRequest(defaultTag language.Tag, r *http.Request) (tag language.Tag, path string) {
	tag = language.Und
	path = r.URL.Path

	for domainTag, domain := range p.domains {
		if r.Host == domain.Host {
			tag = domainTag
			// log.DebugF("found [%v] %v - %v", tag, domain.Host, path)
			break
		}
	}
	return
}