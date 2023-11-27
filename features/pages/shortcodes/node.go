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
	"fmt"

	beContext "github.com/go-enjin/be/pkg/context"
)

type Node struct {
	Raw        string
	Name       string
	Content    string
	Children   Nodes
	Attributes *Attributes
	Shortcode  *Shortcode
	Feature    ShortcodeFeature
	isClosing  bool
	isClosed   bool
}

func newNode(f Feature, name, content string) (node *Node) {
	var sc *Shortcode
	if v, ok := f.LookupShortcode(name); ok {
		sc = &v
	}
	node = &Node{
		Name:       name,
		Raw:        content,
		Content:    content,
		Attributes: newAttributes(),
		Shortcode:  sc,
		Feature:    f,
	}
	return
}

func (node *Node) CloneRawText() (cloned *Node) {
	cloned = newNode(node.Feature, "", node.Raw)
	return
}

func (node *Node) CloneOpening() (cloned *Node) {
	cloned = newNode(node.Feature, node.Name, "")
	cloned.Attributes = node.Attributes.Clone()
	return
}

func (node *Node) CloneClosing() (cloned *Node) {
	cloned = newNode(node.Feature, node.Name, "")
	cloned.isClosing = true
	return
}

func (node *Node) String() (output string) {
	if node.Name == "" {
		output += fmt.Sprintf(`(%s)`, node.Content)
		return
	}
	output += fmt.Sprintf(`(%s%s`, node.Name, node.Attributes.String())
	output += "["
	for idx, child := range node.Children {
		if idx > 0 {
			output += ","
		}
		output += child.String()
	}
	output += "]"
	if node.isClosed || node.isClosing {
		output += "/"
	}
	output += ")"
	return
}

func (node *Node) IsTextNode() (is bool) {
	is = node.Name == ""
	return
}

func (node *Node) IsShortcode() (is bool) {
	is = node.Name != "" && node.Shortcode != nil
	return
}

func (node *Node) InlineOnly() (only bool) {
	only = node.CanInlineRender() && !node.CanBlockRender()
	return
}

func (node *Node) BlockOnly() (only bool) {
	only = !node.CanInlineRender() && node.CanBlockRender()
	return
}

func (node *Node) CanInlineRender() (can bool) {
	can = node.IsShortcode() && node.Shortcode.InlineFn != nil
	return
}

func (node *Node) CanBlockRender() (can bool) {
	can = node.IsShortcode() && node.Shortcode.RenderFn != nil
	return
}

func (node *Node) Append(children ...*Node) {
	node.Children = append(node.Children, children...)
	return
}

func (node *Node) Render(ctx beContext.Context) (output string) {

	if node.Name == "" {
		output += node.Content
		return
	}

	if node.IsShortcode() {
		// this is a shortcode, not just text

		if node.Children.Len() > 0 {
			// wants block render

			if node.CanBlockRender() {
				// does block render
				output += node.Shortcode.RenderFn(node, ctx)

			} else if node.CanInlineRender() {
				// does inline render
				inline := newNode(node.Feature, node.Name, "")
				inline.Attributes = node.Attributes
				output += node.Shortcode.InlineFn(inline, ctx)
				// with any children rendered after
				output += node.Children.Render(ctx)

			} else {
				// shortcode does not have inline or render funcs
				output += node.Raw
			}

		} else if node.CanInlineRender() {
			// wants inline, does inline render
			output += node.Shortcode.InlineFn(node, ctx)

		} else if node.CanBlockRender() {
			// wants inline, does block render
			output += node.Shortcode.RenderFn(node, ctx)

		} else {
			// shortcode does not have inline or render funcs
			output += node.Raw
		}

	} else {
		// unknown shortcode
		output += node.Raw
	}

	return
}
