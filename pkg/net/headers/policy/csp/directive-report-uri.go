// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import "fmt"

var _ Directive = (*reportUriDirective)(nil)

type reportUriDirective string

func (d reportUriDirective) DirectiveType() string {
	return "report-uri"
}

func (d reportUriDirective) Value() string {
	return fmt.Sprintf("report-uri %v", d)
}

func NewReportUri(uri string) Directive {
	return reportUriDirective(uri)
}
