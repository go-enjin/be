// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

type upgradeInsecureRequestsDirective string

func (d upgradeInsecureRequestsDirective) Type() string {
	return string(d)
}

func (d upgradeInsecureRequestsDirective) Value() string {
	return string(d)
}

func NewUpgradeInsecureRequests() Directive {
	return upgradeInsecureRequestsDirective("upgrade-insecure-requests")
}