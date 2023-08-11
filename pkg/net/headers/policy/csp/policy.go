// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import (
	"sort"

	"github.com/fvbommel/sortorder"

	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/slices"
)

type Policy interface {
	// Set overwrites any existing version of the same directives (chainable)
	Set(d Directive) Policy
	// Add appends the given directive (chainable)
	Add(d Directive) Policy
	// Value returns a string suitable for use in HTTP header responses
	Value() string
	// Find returns all directive instances of named type
	Find(name string) (found []Directive)
	// None returns true if Empty or there is only the None source present in the named directive
	None(name string) (none bool)
	// Empty returns true if there are no directives present
	Empty() (empty bool)
	// Unsafe returns true if any "unsafe" sources are present in the named directive
	Unsafe(name string) (unsafe bool)
	// Collapse reduces directives of the same type and places default-src first, returns a new Policy
	Collapse() Policy
	// Directives returns the list of directives present
	Directives() (directives []Directive)
}

type cPolicy []Directive

func NewPolicy(directives ...Directive) (p Policy) {
	np := cPolicy(directives)
	p = &np
	return
}

func StrictContentSecurityPolicy() Policy {
	return &cPolicy{
		NewDefaultSrc(Self, SchemeSource("https")),
		NewFrameAncestors(None),
		NewObjectSrc(None),
	}
}

func DefaultContentSecurityPolicy() Policy {
	return &cPolicy{
		NewDefaultSrc(Self, SchemeSource("https"), SchemeSource("data"), UnsafeInline),
		NewFrameAncestors(None),
		NewObjectSrc(None),
	}
}

func (p *cPolicy) Value() (value string) {
	var sorted []Directive
	sorted = append(sorted, *p...)
	sort.Slice(sorted, func(i, j int) (less bool) {
		a := sorted[i]
		b := sorted[j]
		aType := a.DirectiveType()
		bType := b.DirectiveType()
		aIsReport := slices.Present(aType, "report-to", "report-uri")
		bIsReport := slices.Present(bType, "report-to", "report-uri")
		aIsDefault := aType == "default-src"
		bIsDefault := bType == "default-src"
		switch {
		case aIsDefault && !bIsDefault:
			less = true
		case aIsDefault && bIsDefault:
			less = true
		case !aIsDefault && bIsDefault:
			less = false
		case aIsReport && !bIsReport:
			less = false
		case aIsReport && bIsReport:
			less = sortorder.NaturalLess(aType, bType)
		case !aIsReport && bIsReport:
			less = true
		default:
			less = sortorder.NaturalLess(aType, bType)
		}
		return
	})
	for idx, d := range sorted {
		if idx > 0 {
			value += " ; "
		}
		value += d.Value()
	}
	return
}

func (p *cPolicy) Set(d Directive) Policy {
	data, order := p.makeDataMap()
	dType := d.DirectiveType()
	if !slices.Within(dType, order) {
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
		dType := d.DirectiveType()
		if !slices.Within(dType, order) {
			order = append(order, dType)
		}
		data[dType] = append(data[dType], d)
	}
	return
}

func (p *cPolicy) Find(name string) (found []Directive) {
	for _, d := range *p {
		if d.DirectiveType() == name {
			found = append(found, d)
		}
	}
	return
}

func (p *cPolicy) None(name string) (none bool) {
	found := p.Find(name)
	if none = len(found) == 0; !none {
		var notNone bool
		for _, d := range found {
			if sd, ok := d.(SourceDirective); ok {
				sources := sd.Sources()
				switch len(sources) {
				case 1:
					none = sources[0].Value() == None.Value()
				case 0:
					none = true
				default:
					notNone = true
				}
			}
		}
		if none && notNone {
			none = false
		}
	}
	return
}

func (p *cPolicy) Empty() (empty bool) {
	empty = len(*p) == 0
	return
}

func (p *cPolicy) Unsafe(name string) (unsafe bool) {
	for _, d := range p.Find(name) {
		if sd, ok := d.(SourceDirective); ok {
			for _, src := range sd.Sources() {
				switch src.Value() {
				case UnsafeInline.Value(), UnsafeEval.Value(), UnsafeHashes.Value():
					unsafe = true
					return
				}
			}
		}
	}
	return
}

func (p *cPolicy) Directives() (directives []Directive) {
	directives = append(directives, *p...)
	return
}

func (p *cPolicy) Collapse() Policy {
	collapsed := make(cPolicy, 0)
	data := make(map[string]Directive)
	for _, d := range *p {
		dType := d.DirectiveType()
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