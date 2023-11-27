// Copyright (c) 2023  The Go-Enjin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package editor

type FormMethod string

const (
	NilFormMethod  FormMethod = ""
	GetFormMethod  FormMethod = "get"
	PostFormMethod FormMethod = "post"
)

type Action struct {
	// Key is the kebab-cased verb used for machine purposes
	Key string `json:"key"`
	// Name is the label displayed for this action
	Name string `json:"name"`
	// Icon is the icon class to display
	Icon string `json:"icon"`
	// Button is the button class to render with (ie: "primary", "secondary", "caution", "danger", etc)
	Class string `json:"class"`
	// Active indicates whether this action is clickable or simply an indicator
	Active bool `json:"active"`
	// Prompt is the text to use in a prompt for confirmation of this action, leave empty to act without confirmation
	Prompt string `json:"prompt,omitempty"`
	// Method is the form method for submitting this action
	Method FormMethod `json:"method,omitempty"`
	// Tilde is a WorkFile extension to include in the URL of action links
	Tilde string `json:"tilde,omitempty"`
	// Dialog specifies the editor dialog template to use (default is "yesno")
	Dialog string `json:"dialog,omitempty"`
	// Order specifies the weighted position of this action in the menu UX, actions with an Order less than 10 are to
	// be displayed beside the menu dropdown for quick-access
	Order int `json:"order"`
}
