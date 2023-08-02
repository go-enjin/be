package shortcodes

import (
	beContext "github.com/go-enjin/be/pkg/context"
)

type Node struct {
	Raw        string
	Name       string
	Content    string
	Children   Nodes
	Attributes *Attributes
	parser     *parser
}

func newNode(p *parser, name, content string) (node *Node) {
	node = &Node{
		Name:       name,
		Raw:        content,
		Content:    content,
		Attributes: newNodeAttributes(),
		parser:     p,
	}
	return
}

func (node *Node) Append(children ...*Node) {
	node.Children = append(node.Children, children...)
	return
}

func (node *Node) Render(ctx beContext.Context) (output string) {

	if node.Name == "" {
		// text node has either children or content
		if len(node.Children) > 0 {
			for _, child := range node.Children {
				output += child.Render(ctx)
			}
		} else {
			output += node.Content
		}
		return
	}

	// this is a shortcode, not just text

	if sc, ok := node.parser.feature.LookupShortcode(node.Name); ok {

		output += sc.RenderFn(node, ctx)

	} else {
		// unknown shortcode
		output += node.Raw
	}

	return
}