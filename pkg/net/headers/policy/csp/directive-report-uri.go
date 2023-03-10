// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

import "fmt"

type reportUriDirective string

func (d reportUriDirective) Type() string {
	return "report-uri"
}

func (d reportUriDirective) Value() string {
	return fmt.Sprintf("report-uri %v", d)
}

func NewReportUri(uri string) Directive {
	return reportUriDirective(uri)
}