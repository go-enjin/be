// Copyright (c) 2024  The Go-Enjin Authors
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

	"github.com/go-corelibs/htmlcss"
	"github.com/go-corelibs/maps"
	clStrings "github.com/go-corelibs/strings"
	beContext "github.com/go-enjin/be/pkg/context"
)

var (
	FontAwesomeIconShortcode = Shortcode{
		Name: "fa-icon",
		InlineFn: func(node *Node, ctx beContext.Context) (output string) {
			class, _ := node.Attributes.Lookup["class"]
			classes := htmlcss.ParseClass(class)
			styles := make(map[string]string)

			if v, ok := node.Attributes.Lookup["fa-icon"]; ok && v != "" {
				if _, ignore := node.Attributes.Lookup["name"]; !ignore {
					node.Attributes.Set("name", v)
				}
			}

			faParseFamilyStyle(node, classes)
			faParseIconName(node, classes)
			faParseIconSize(node, classes)
			faParseFlipRotate(node, classes, styles)
			faParseFixedWidth(node, classes)
			faParseBorder(node, classes, styles)
			faParsePull(node, classes, styles)

			output += `<i`
			if v := classes.String(); v != "" {
				output += fmt.Sprintf(` class=%q`, v)
			}
			if len(styles) > 0 {
				s := ""
				for _, key := range maps.SortedKeys(styles) {
					s += key + ":" + styles[key] + ";"
				}
				output += fmt.Sprintf(` style=%q`, s)
			}
			output += `></i>`
			return
		},
	}
)

func faParseFamilyStyle(node *Node, classes htmlcss.CssClass) {
	family := ""
	if v, ok := node.Attributes.Lookup["family"]; ok {
		v = strings.ToLower(v)
		switch v {
		case "sharp", "brands":
			family = v
		}
	}

	if vs, ok := node.Attributes.Lookup["style"]; ok {
		vs = strings.ToLower(vs)
		switch family {
		case "sharp":
			switch vs {
			case "solid", "regular", "light", "thin":
				classes.Add("fa-sharp fa-" + vs)
			default:
				classes.Add("fa-solid fa-" + vs)
			}
		case "brands":
			classes.Add("fa-brands fa-" + vs)
		default:
			switch vs {
			case "solid", "regular", "light", "thin", "duotone":
				classes.Add("fa-" + vs)
			default:
				classes.Add("fa-solid")
			}
		}
	}
}

func faParseIconName(node *Node, classes htmlcss.CssClass) {
	if vi, ok := node.Attributes.Lookup["name"]; ok {
		if vi = strings.ToLower(vi); vi != "" {
			classes.Add("fa-" + vi)
		}
	}
}

func faParseIconSize(node *Node, classes htmlcss.CssClass) {
	if v, ok := node.Attributes.Lookup["size"]; ok {
		v = strings.ToLower(v)
		switch v {
		case "2xs", "xs", "sm", "lg", "xl", "2xl",
			"1x", "2x", "3x", "4x", "5x", "6x", "7x", "8x", "9x", "10x":
			classes.Add("fa-" + v)
		}
	}
}

func faParseFixedWidth(node *Node, classes htmlcss.CssClass) {
	if v, ok := node.Attributes.Lookup["fixed"]; ok {
		if clStrings.IsTrue(v) {
			classes.Add("fa-fw")
		}
	}
}

func faParseFlipRotate(node *Node, classes htmlcss.CssClass, styles map[string]string) {
	if v, ok := node.Attributes.Lookup["flip"]; ok {
		v = strings.ToLower(v)
		switch v {
		case "horizontal", "vertical", "both":
			classes.Add("fa-flip-" + v)
		}
	} else if v, ok := node.Attributes.Lookup["rotate"]; ok {
		v = strings.ToLower(v)
		switch v {
		case "90", "180", "270":
			classes.Add("fa-rotate-" + v)
		default:
			classes.Add("fa-rotate-by")
			styles["--fa-rotate-angle"] = v
		}
	}
}

func faParseBorder(node *Node, classes htmlcss.CssClass, styles map[string]string) {
	var border bool
	if v, ok := node.Attributes.Lookup["border"]; ok {
		border = clStrings.IsTrue(v)
	}
	for _, key := range []string{"color", "padding", "radius", "style", "width"} {
		if v, ok := node.Attributes.Lookup["border-"+key]; ok {
			border = true
			styles["--fa-border-"+key] = v
		}
	}
	if border {
		classes.Add("fa-border")
	}
}

func faParsePull(node *Node, classes htmlcss.CssClass, styles map[string]string) {
	if v, ok := node.Attributes.Lookup["pull"]; ok {
		v = strings.ToLower(v)
		switch v {
		case "left", "right":
			classes.Add("fa-pull-" + v)
		}
		if margin, found := node.Attributes.Lookup["pull-margin"]; found {
			styles["--fa-pull-margin"] = margin
		}
	}
}
