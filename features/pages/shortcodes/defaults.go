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
	"strings"

	beContext "github.com/go-enjin/be/pkg/context"
)

func BasicRenderFn(name, content string) (output string) {
	output += "<" + name + ">"
	output += content
	output += "</" + name + ">"
	return
}

// ParseRawContents returns the raw contents of a *Node with empty first and last lines removed
func ParseRawContents(node *Node) (contents string) {
	var raw string
	if len(node.Children) > 0 {
		raw = node.Children.Raw()
	} else {
		raw = node.Raw
	}
	lines := strings.Split(raw, "\n")
	last := len(lines) - 1
	for idx, line := range lines {
		if idx == 0 && strings.TrimSpace(line) == "" {
			// skip empty first line
			continue
		} else if idx == last && strings.TrimSpace(line) == "" {
			// skip empty last line
			break
		} else if contents += line; idx < last {
			// added line to contents and not last so add newline too
			contents += "\n"
		}
	}
	return
}

var (

	// BoldShortcode makes the content text <b>bold</b>
	BoldShortcode = Shortcode{
		Name: "b",
		RenderFn: func(node *Node, ctx beContext.Context) (output string) {
			output = BasicRenderFn("b", node.Children.Render(ctx))
			return
		},
	}

	// ItalicShortcode makes the content text <i>italicised</i>
	ItalicShortcode = Shortcode{
		Name: "i",
		RenderFn: func(node *Node, ctx beContext.Context) (output string) {
			output = BasicRenderFn("i", node.Children.Render(ctx))
			return
		},
	}

	// UnderlineShortcode makes the content text <u>underlined</u>
	UnderlineShortcode = Shortcode{
		Name: "u",
		RenderFn: func(node *Node, ctx beContext.Context) (output string) {
			output = BasicRenderFn("u", node.Children.Render(ctx))
			return
		},
	}

	// StrikethroughShortcode makes the content text <s>crossed out</s>
	StrikethroughShortcode = Shortcode{
		Name: "s",
		RenderFn: func(node *Node, ctx beContext.Context) (output string) {
			output = BasicRenderFn("s", node.Children.Render(ctx))
			return
		},
	}

	// SuperscriptShortcode makes the content text <sup>raised</sup>
	SuperscriptShortcode = Shortcode{
		Name: "sup",
		RenderFn: func(node *Node, ctx beContext.Context) (output string) {
			output = BasicRenderFn("sup", node.Children.Render(ctx))
			return
		},
	}

	// SubscriptShortcode makes the content text <sub>lowered</sub>
	SubscriptShortcode = Shortcode{
		Name: "sub",
		RenderFn: func(node *Node, ctx beContext.Context) (output string) {
			output = BasicRenderFn("sub", node.Children.Render(ctx))
			return
		},
	}

	// UrlShortcode makes an anchor tag from the attributes and contents
	// - [url]https://go-enjin.org[/url]
	// - [url=https://go-enjin.org]Go-Enjin.org[/url]
	// - [url href=https://go-enjin.org]Go-Enjin.org[/url]
	// - [url link=https://go-enjin.org target=_blank]Go-Enjin.org[/url]
	UrlShortcode = Shortcode{
		Name: "url",
		RenderFn: func(node *Node, ctx beContext.Context) (output string) {
			// if not attributes, the contents is the URL
			var url, label, target string
			if v, ok := node.Attributes.Lookup["url"]; ok {
				url = v
			} else if v, ok := node.Attributes.Lookup["link"]; ok {
				url = v
			} else if v, ok := node.Attributes.Lookup["href"]; ok {
				url = v
			} else {
				url = node.Content
			}
			if v, ok := node.Attributes.Lookup["target"]; ok {
				v = strings.ToLower(v)
				switch v {
				case "_self", "_blank", "_parent", "_top":
					target = v
				}
			}
			label = node.Children.Render(ctx)
			output = "<a "
			output += "href=\"" + url + "\""
			if target != "" {
				output += " target=\"" + target + "\""
			}
			output += ">" + label + "</a>"
			return
		},
	}

	// ColorShortcode changes the foreground and optionally the background colours of the content text
	// - [color=green]green text[/color]
	// - [color fg=yellow bg=green]yellow text on a green background[/color]
	ColorShortcode = Shortcode{
		Name:    "color",
		Aliases: []string{"colour"},
		RenderFn: func(node *Node, ctx beContext.Context) (output string) {
			var fg, bg string

			// foreground
			if v, ok := node.Attributes.Lookup["colour"]; ok {
				fg = v
			} else if v, ok := node.Attributes.Lookup["color"]; ok {
				fg = v
			} else if v, ok := node.Attributes.Lookup["fg"]; ok {
				fg = v
			} else {
				fg = "orange"
			}

			// background
			if v, ok := node.Attributes.Lookup["background-colour"]; ok {
				bg = v
			} else if v, ok := node.Attributes.Lookup["background-color"]; ok {
				bg = v
			} else if v, ok := node.Attributes.Lookup["background"]; ok {
				bg = v
			} else if v, ok := node.Attributes.Lookup["bg"]; ok {
				bg = v
			}

			if bg != "" {
				output += fmt.Sprintf(`<span style="color:%v;background-color:%v;">`, fg, bg)
			} else {
				output += fmt.Sprintf(`<span style="color:%v;">`, fg)
			}

			output += node.Children.Render(ctx)
			output += `</span>`
			return
		},
	}

	CodeShortcode = Shortcode{
		Name: "code",
		RenderFn: func(node *Node, ctx beContext.Context) (output string) {
			output += "<pre>"
			output += ParseRawContents(node)
			output += "</pre>"
			return
		},
	}

	QuoteShortcode = Shortcode{
		Name:    "quote",
		Aliases: []string{"blockquote"},
		RenderFn: func(node *Node, ctx beContext.Context) (output string) {
			output += "<blockquote>"
			output += ParseRawContents(node)
			output += "</blockquote>"

			var cite string

			if v, ok := node.Attributes.Lookup["cite"]; ok {
				cite = v
			} else if v, ok := node.Attributes.Lookup["quote"]; ok {
				cite = v
			} else if v, ok := node.Attributes.Lookup["author"]; ok {
				cite = v
			}

			if cite != "" {
				output += "<figcaption>"
				output += "<cite>" + cite + "</cite>"
				output += "</figcaption>"
			}
			return
		},
	}
)