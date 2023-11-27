// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package permissions

import "fmt"

const (
	AllowAll  keywordOrigin = "*"
	AllowNone keywordOrigin = "()"
	AllowSelf keywordOrigin = "self"
	AllowSrc  keywordOrigin = "src"
)

type Origin interface {
	OriginType() string
	Value() string
}

type keywordOrigin string

func (k keywordOrigin) OriginType() string {
	return "keyword-origin"
}

func (k keywordOrigin) Value() string {
	return string(k)
}

type specificOrigin string

func NewSpecificOrigin(uri string) (o Origin) {
	o = specificOrigin(uri)
	return
}

func (s specificOrigin) OriginType() string {
	return "specific-origin"
}

func (s specificOrigin) Value() (value string) {
	return fmt.Sprintf(`"%v"`, string(s))
}
