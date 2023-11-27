// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import "fmt"

var _ Directive = (*reportToDirective)(nil)

type reportToDirective string

func (d reportToDirective) DirectiveType() string {
	return "report-to"
}

func (d reportToDirective) Value() string {
	return fmt.Sprintf("report-to %v", d)
}

func NewReportTo(groupName string) Directive {
	return reportToDirective(groupName)
}
