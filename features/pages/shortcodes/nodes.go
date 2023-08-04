package shortcodes

import (
	beContext "github.com/go-enjin/be/pkg/context"
)

type Nodes []*Node

func (nodes Nodes) Len() (count int) {
	count = len(nodes)
	return
}

func (nodes Nodes) Raw() (raw string) {
	for _, child := range nodes {
		raw += child.Raw
	}
	return
}

func (nodes Nodes) Render(ctx beContext.Context) (output string) {
	for _, child := range nodes {
		output += child.Render(ctx)
	}
	return
}