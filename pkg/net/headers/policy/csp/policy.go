// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import (
	"github.com/go-enjin/be/pkg/maps"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

type Policy interface {
	// Set overwrites any existing version of the same directives (chainable)
	Set(d Directive) Policy
	// Add appends the given directive (chainable)
	Add(d Directive) Policy
	// Value returns a string suitable for use in HTTP header responses
	Value() string
	// Collapse reduces directives of the same type and places default-src first, returns a new Policy
	Collapse() Policy
}

type cPolicy []Directive

func NewPolicy(directives ...Directive) (p Policy) {
	np := cPolicy(directives)
	p = &np
	return
}

func (p *cPolicy) Value() (value string) {
	for idx, d := range *p {
		if idx > 0 {
			value += ";"
		}
		value += d.Value()
	}
	return
}

func (p *cPolicy) Set(d Directive) Policy {
	data, order := p.makeDataMap()
	dType := d.Type()
	if !beStrings.StringInSlices(dType, order) {
		order = append(order, dType)
	}
	data[dType] = []Directive{d}
	next := make(cPolicy, 0)
	for _, name := range order {
		next = append(next, data[name]...)
	}
	*p = next
	return p
}

func (p *cPolicy) Add(d Directive) Policy {
	*p = append(*p, d)
	return p
}

func (p *cPolicy) makeDataMap() (data map[string][]Directive, order []string) {
	data = make(map[string][]Directive)
	for _, d := range *p {
		dType := d.Type()
		if !beStrings.StringInSlices(dType, order) {
			order = append(order, dType)
		}
		data[dType] = append(data[dType], d)
	}
	return
}

func (p *cPolicy) Collapse() Policy {
	collapsed := make(cPolicy, 0)
	data := make(map[string]Directive)
	for _, d := range *p {
		dType := d.Type()
		if dSourceDirective, ok := d.(SourceDirective); ok {
			if existingSourceDirective, ok := data[dType].(SourceDirective); ok {
				existingSourceDirective.Append(dSourceDirective.Sources()...)
				data[dType] = existingSourceDirective
			} else {
				data[dType] = dSourceDirective
			}
		} else {
			data[dType] = d
		}
	}
	if defaultSrc, ok := data["default-src"]; ok {
		collapsed = append(collapsed, defaultSrc)
		delete(data, "default-src")
	}
	for _, name := range maps.SortedKeys(data) {
		collapsed = append(collapsed, data[name])
	}
	return &collapsed
}