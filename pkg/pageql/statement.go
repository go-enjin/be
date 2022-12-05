// Copyright (c) 2022  The Go-Enjin Authors
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

package pageql

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Statement struct {
	Expression  *Expression `parser:"@@" json:"expressions,omitempty"`
	Limit       *int        `parser:"( 'LIMIT' @Int )?" json:"limit,omitempty"`
	Offset      *int        `parser:"( 'OFFSET' @Int )?" json:"offset,omitempty"`
	OrderBy     *string     `parser:"( 'ORDER' 'BY' '.' @Ident )?" json:"order-by,omitempty"`
	SortDir     *string     `parser:"( @'ASC' | @'DSC' | @'DESC' )?" json:"sort-dir,omitempty"`
	ContextKeys []string    `parser:"" json:"context-keys,omitempty"`
	rendered    bool        `parser:""`
}

func (s *Statement) Render() (out *Statement) {
	out = new(Statement)
	if s.Expression != nil {
		out.Expression = s.Expression.Render()
	}
	if s.Limit != nil {
		limit := *s.Limit
		out.Limit = &limit
	}
	if s.Offset != nil {
		offset := *s.Offset
		out.Offset = &offset
	}
	if s.OrderBy != nil {
		orderBy := *s.OrderBy
		out.OrderBy = &orderBy
	}
	if s.SortDir != nil {
		sortDir := *s.SortDir
		out.SortDir = &sortDir
	}
	out.ContextKeys = append(out.ContextKeys, s.ContextKeys...)
	out.rendered = true
	return
}

func (s *Statement) String() (query string) {
	if s.rendered {
		return
	}
	var compile func(expr *Expression)
	compile = func(expr *Expression) {
		switch {
		case expr.Operation != nil:
			var right string
			switch {
			case expr.Operation.Right.ContextKey != nil:
				right = "." + *expr.Operation.Right.ContextKey
			case expr.Operation.Right.String != nil:
				right = *expr.Operation.Right.String
			case expr.Operation.Right.Regexp != nil:
				right = "m" + *expr.Operation.Right.Regexp
			}
			query += fmt.Sprintf("(.%s %s %s)", *expr.Operation.Left, expr.Operation.Type, right)

		case expr.Condition != nil:
			query += "("
			compile(expr.Condition.Left)
			query += " " + strings.ToUpper(expr.Condition.Type) + " "
			compile(expr.Condition.Right)
			query += ")"
		}
	}
	compile(s.Expression)
	apply := func(v string) {
		if query != "" {
			query += " "
		}
		query += v
	}
	if s.Limit != nil {
		apply(fmt.Sprintf("LIMIT %d", *s.Limit))
	}
	if s.Offset != nil {
		apply(fmt.Sprintf("OFFSET %d", *s.Offset))
	}
	if s.OrderBy != nil {
		apply(fmt.Sprintf("ORDER BY .%s", *s.OrderBy))
	}
	if s.SortDir != nil {
		apply(strings.ToUpper(*s.SortDir))
	}
	return
}

func (s *Statement) Stringify() (out string) {
	b, _ := json.MarshalIndent(s.Render(), "", "  ")
	out = string(b)
	return
}