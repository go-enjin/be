// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package permissions

import "github.com/go-enjin/be/pkg/maps"

type Policy interface {
	Set(d Directive) Policy

	Get(name string) Directive

	Value() (value string)
}

type cPolicy struct {
	directives map[string]Directive
}

func NewPolicy(directives ...Directive) (policy Policy) {
	p := &cPolicy{
		directives: make(map[string]Directive),
	}
	for _, d := range directives {
		p.directives[d.DirectiveName()] = d
	}
	policy = p
	return
}

func (p *cPolicy) Set(d Directive) Policy {
	p.directives[d.DirectiveName()] = d
	return p
}

func (p *cPolicy) Get(name string) (d Directive) {
	d, _ = p.directives[name]
	return
}

func (p *cPolicy) Value() (value string) {
	for idx, name := range maps.SortedKeys(p.directives) {
		if idx > 0 {
			value += ", "
		}
		value += p.directives[name].Value()
	}
	return
}
