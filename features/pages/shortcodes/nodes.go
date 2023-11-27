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
	"github.com/go-enjin/be/pkg/slices"
)

type Nodes []*Node

func (stack Nodes) Len() (count int) {
	count = len(stack)
	return
}

func (stack Nodes) String() (output string) {
	output += "["
	for idx, node := range stack {
		if idx > 0 {
			output += ","
		}
		output += "\n\t"
		output += node.String()
	}
	output += "\n]"
	return
}

func (stack Nodes) Pop() (node *Node, remainder Nodes) {
	if lastIdx := stack.Last(); lastIdx > -1 {
		node = stack[lastIdx]
		remainder = stack[:lastIdx]
	}
	return
}

func (stack Nodes) Last() (idx int) {
	idx = stack.Len() - 1
	return
}

func (stack Nodes) LastBlockOnly() (idx int) {
	idx = -1
	for i := stack.Last(); i > -1; i-- {
		if stack[i].BlockOnly() {
			idx = i
			return
		}
	}
	return
}

func (stack Nodes) LastNamed(name string) (idx int) {
	idx = -1
	for i := stack.Last(); i > -1; i-- {
		if stack[i].Name == name {
			idx = i
			return
		}
	}
	return
}

func (stack Nodes) LastOpen() (idx int) {
	idx = -1
	for i := stack.Last(); i > -1; i-- {
		if !stack[i].isClosed && !stack[i].isClosing {
			idx = i
			return
		}
	}
	return
}

func (stack Nodes) LastOpened(name string) (idx int) {
	idx = -1
	for i := stack.Last(); i > -1; i-- {
		if stack[i].Name == name && !stack[i].isClosed && !stack[i].isClosing {
			idx = i
			return
		}
	}
	return
}

func (stack Nodes) FindNamed(name string) (indexes map[int]interface{}) {
	indexes = make(map[int]interface{})
	for i := 0; i < stack.Len(); i++ {
		if stack[i].Name == name {
			indexes[i] = stack[i]
		} else if stack[i].Children.Len() > 0 {
			found := stack[i].Children.FindNamed(name)
			if len(found) > 0 {
				indexes[i] = found
			}
		}
	}
	return
}

func (stack Nodes) Raw() (raw string) {
	for _, child := range stack {
		raw += child.Raw
	}
	return
}

func (stack Nodes) Render(ctx beContext.Context) (output string) {
	for _, child := range stack {
		output += child.Render(ctx)
	}
	return
}

// Flatten returns the stack with all nodes and their children flattened out
// into a linear list - from tree to list
func (stack Nodes) Flatten() (flattened Nodes) {
	for _, node := range stack {
		if node.Name == "" {
			flattened = append(flattened, node)
		} else {
			flattened = append(flattened, newNode(node.Feature, node.Name, ""))
			if node.Children.Len() > 0 {
				children := node.Children.Flatten()
				flattened = append(flattened, children...)
				closing := newNode(node.Feature, node.Name, "")
				closing.isClosing = true
				flattened = append(flattened, closing)
			}
		}
	}
	return
}

func (stack Nodes) ListOpenings() (opened Nodes) {
	for _, node := range stack {
		if node.Name != "" && node.IsShortcode() && !node.isClosing {
			opened = append(opened, node)
		}
	}
	return
}

// AutoClose assumes the stack is a list of opened nodes, meant to be collapsed down to a single parent
func (stack Nodes) AutoClose() (resolved *Node) {

	last := stack.Last()
	for i := last; i > -1; i-- {

		stack[i].isClosed = true
		if i == 0 {
			stack[0].isClosed = true
			resolved = stack[0]
			return
		}
		stack[i-1].Append(stack[i])

	}

	return
}

func (stack Nodes) IsNextNamedClosing(from int, name string) (isClosing bool) {
	for i := from; i < stack.Len(); i++ {
		if stack[i].IsShortcode() {
			if stack[i].Name == name {
				isClosing = stack[i].isClosing
				return
			}
		}
	}
	return
}

func (stack Nodes) FindOpeningsNamed(from int, name string) (indexes []int) {
	for i := from; i < stack.Len(); i++ {
		if stack[i].IsShortcode() {
			if !stack[i].isClosing && stack[i].Name == name {
				indexes = append(indexes, i)
			}
		}
	}
	return
}

func (stack Nodes) FindClosingsNamed(from int, name string) (indexes []int) {
	for i := from; i < stack.Len(); i++ {
		if stack[i].IsShortcode() {
			if stack[i].isClosing && stack[i].Name == name {
				indexes = append(indexes, i)
			}
		}
	}
	return
}

// Collapse returns the stack with all nodes inlined or auto-closed
//   - from list to tree
func (stack Nodes) Collapse() (collapsed Nodes) {

	if stack.Len() == 0 {
		collapsed = Nodes{}
		return
	}

	// opened is a running lineage of container nodes (parent=0,child=1,grandchild=2...)
	var opened Nodes
	// stream is used to not manipulate the internal stack
	stream := slices.Copy(stack)

	appendOrCollapse := func(node *Node) {
		if node.isClosing {
			return
		}
		if last := opened.Last(); last > -1 {
			opened[last].Append(node)
		} else {
			node.isClosed = true
			collapsed = append(collapsed, node)
		}
	}

	closeOpened := func(lastOpenedIdx, streamIdx int, node *Node) {
		inner := slices.Copy(opened[lastOpenedIdx+1:])
		lastOpened := opened[lastOpenedIdx]
		opened = slices.Truncate(opened, lastOpenedIdx)

		// TODO: don't auto-close inline-able nodes?
		if innerCollapsed := inner.AutoClose(); innerCollapsed != nil {
			lastOpened.Append(innerCollapsed)
		}

		lastOpened.isClosed = true

		if lastOpenedIdx == 0 {
			collapsed = append(collapsed, lastOpened)
		} else {
			opened[lastOpenedIdx-1].Append(lastOpened)
		}
		if !node.isClosing {
			if node.BlockOnly() {
				opened = append(opened, node)
			} else { // node can inline
				if stream.IsNextNamedClosing(streamIdx+1, node.Name) {
					// next instance of this node is a closing tag
					opened = append(opened, node)
				} else if lastOpenedIdx > 0 {
					// next instance of this node is not a closing tag, or there is no closing tag at all
					opened[lastOpenedIdx-1].Append(node)
				} else {
					// opened.Len is 0, collapse
					collapsed = append(collapsed, node)
				}
			}
		}

		if innerOpens := inner.ListOpenings(); innerOpens.Len() > 0 {
			var reopens Nodes
			for _, child := range innerOpens {
				//if child.CanBlockRender() {
				if child.BlockOnly() {
					childClosingIndexes := stream.FindClosingsNamed(streamIdx+1, child.Name)
					if len(childClosingIndexes) > 0 {
						reopens = append(reopens, child.CloneOpening())
					}
				}
			}
			if reopens.Len() > 0 {
				stream = slices.Insert[*Node](stream, streamIdx+1, reopens...)
			}
		}

		// scan forward for remaining closing tags
		foundOpeningIndexes := stream.FindOpeningsNamed(streamIdx+1, node.Name)
		foundClosingIndexes := stream.FindClosingsNamed(streamIdx+1, node.Name)
		// insert this node reopening here because the lower-in-stack opening
		// wasn't intentionally closed?
		if count := len(foundClosingIndexes); count-1 > len(foundOpeningIndexes) {
			clonedOpen := lastOpened.CloneOpening()
			stream = slices.Insert(stream, foundClosingIndexes[0]+1, clonedOpen)
		}
	}

	for idx := 0; idx < stream.Len(); idx++ {
		node := stream[idx]

		if node.Name == "" {
			// plain text collapses to last opened
			appendOrCollapse(node)
			continue
		}

		if !node.IsShortcode() {
			// not a shortcode and not plain text, collapse to last opened
			appendOrCollapse(node.CloneRawText())
			continue
		}

		if node.InlineOnly() {
			// inline only nodes are closed and collapsed as-is
			node.isClosed = true
			appendOrCollapse(node)
			continue
		}

		if !node.isClosing {
			// is an opening node

			if node.Shortcode.Nesting {
				// self nesting allowed, stack as-is
				opened = append(opened, node)
				continue
			}

			// nesting not supported
			node.isClosed = false

			if lastOpenedIdx := opened.LastOpened(node.Name); lastOpenedIdx > -1 {
				// this node is nested, prior opened needs to auto-close and
				// auto-reopen after this node

				closeOpened(lastOpenedIdx, idx, node)

			} else {
				// nesting not detected, stack as-is
				if node.BlockOnly() {
					opened = append(opened, node)
				} else if stream.IsNextNamedClosing(idx+1, node.Name) {
					opened = append(opened, node)
				} else {
					if last := opened.Last(); last > -1 {
						node.isClosed = true
						opened[last].Append(node)
					} else {
						collapsed = append(collapsed, node)
					}
				}
			}

			continue
		}

		// is a closing node

		if lastOpenedIdx := opened.LastOpened(node.Name); lastOpenedIdx > -1 {

			closeOpened(lastOpenedIdx, idx, node)

			continue
		}

		// is a stray closing node for a valid tag
		// NOP
		continue

	} // end of stack for loop

	if closed := opened.AutoClose(); closed != nil {
		collapsed = append(collapsed, closed)
	}

	return
}
