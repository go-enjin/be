// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package csp

var _ Directive = (*sandboxDirective)(nil)

type SandboxValue string

const (
	AllowDownloads                      SandboxValue = "allow-downloads"
	AllowDownloadsWithoutUserActivation SandboxValue = "allow-downloads-without-user-activation"
	AllowForms                          SandboxValue = "allow-forms"
	AllowModals                         SandboxValue = "allow-modals"
	AllowOrientationLock                SandboxValue = "allow-orientation-lock"
	AllowPointerLock                    SandboxValue = "allow-pointer-lock"
	AllowPopups                         SandboxValue = "allow-popups"
	AllowPopupsToEscapeSandbox          SandboxValue = "allow-popups-to-escape-sandbox"
	AllowPresentation                   SandboxValue = "allow-presentation"
	AllowSameOrigin                     SandboxValue = "allow-same-origin"
	AllowScripts                        SandboxValue = "allow-scripts"
	AllowStorageAccessByUserActivation  SandboxValue = "allow-storage-access-by-user-activation"
	AllowTopNavigation                  SandboxValue = "allow-top-navigation"
	AllowTopNavigationByUserActivation  SandboxValue = "allow-top-navigation-by-user-activation"
	AllowTopNavigationToCustomProtocols SandboxValue = "allow-top-navigation-to-custom-protocols"
)

type sandboxDirective []SandboxValue

func NewSandbox(values ...SandboxValue) Directive {
	return sandboxDirective(values)
}

func (d sandboxDirective) DirectiveType() string {
	return "sandbox"
}

func (d sandboxDirective) Value() (value string) {
	for idx, v := range d {
		if idx > 0 {
			value += " "
		}
		value += string(v)
	}
	return
}
