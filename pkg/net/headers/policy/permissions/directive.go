// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package permissions

type Directive interface {
	DirectiveName() string
	Value() string
}

func NewDirective(name string, allowed ...Origin) (d Directive) {
	d = cDirective{
		name: name,
		list: allowed,
	}
	return
}

type cDirective struct {
	name string
	list AllowList
}

func (d cDirective) DirectiveName() string {
	return d.name
}

func (d cDirective) Value() (value string) {
	value = d.name + "=" + d.list.Value()
	return
}