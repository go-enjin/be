//go:build page_shortcodes || pages || all

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

package shortcodes

import (
	beContext "github.com/go-enjin/be/pkg/context"
)

type ShortcodeHandlerFn = func(node *Node, ctx beContext.Context) (output string)

type Shortcode struct {
	Name     string
	Aliases  []string
	Nesting  bool
	InlineFn ShortcodeHandlerFn
	RenderFn ShortcodeHandlerFn
}

func (sc Shortcode) Is(name string) (is bool) {
	if is = sc.Name == name; is {
		return
	}
	for _, alias := range sc.Aliases {
		if is = alias == name; is {
			return
		}
	}
	return
}