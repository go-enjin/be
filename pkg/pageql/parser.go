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
	"sort"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/maruel/natural"

	"github.com/go-enjin/be/pkg/maps"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

// TODO: sort out errors with: `(.Title != m!(?i)thing!)` statements, use different regexp open/close markers (ie: ~ or / instead of !)
// TODO: implement ~= and !~ such that the `m` prefix to regexp is no longer necessary
// TODO: implement <, >, <=, >= and general numerical comparison supports
// TODO: eventually implement simple math operations: +, -, /, * and this must support correct operator precedence

var (
	parser = participle.MustBuild[Statement](
		participle.Lexer(lexer.MustSimple([]lexer.SimpleRule{
			{`Keyword`, `(?i)\b(BY|ORDER|LIMIT|OFFSET|TRUE|FALSE|NULL|IS|NOT|AND|OR|IN|ASC|DSC|DESC)\b`},
			{`Ident`, `\b([a-zA-Z][.a-zA-Z0-9]*)\b`},
			{`Int`, `\b(\d+)\b`},
			{`Float`, `\b(\d*\.\d+)\b`},
			{`Number`, `[-+]?\d*\.?\d+([eE][-+]?\d+)?`},
			{`String`, `'[^']*'|"[^"]*"`},
			{"Regexp", `/(.+?)/|\!(.+?)\!|\@(.+?)\@|\~(.+?)\~`},
			{`Operators`, `==|!=|[=?.()]`},
			{"whitespace", `\s+`},
		})),
		// UnquoteRegexp("Regexp"),
		// participle.Unquote("String"),
		participle.CaseInsensitive("Keyword"),
	)
)

func SanitizeQuery(query string) (sanitized string) {
	sanitized = strings.TrimSpace(query)
	return
}

func parseString(query string) (stmnt *Statement, err *ParseError) {
	query = SanitizeQuery(query)

	var participleError error
	if stmnt, participleError = parser.ParseString("pageql", query); participleError != nil {
		err = newParseError(query, participleError)
		return
	}

	var extract func(expr *Expression) (keys []string)
	extract = func(expr *Expression) (keys []string) {
		unique := make(map[string]bool)
		switch {
		case expr.Operation != nil:
			unique[*expr.Operation.Left] = true
			if expr.Operation.Right.ContextKey != nil {
				unique[*expr.Operation.Right.ContextKey] = true
			}
		case expr.Condition != nil:
			for _, key := range extract(expr.Condition.Left) {
				unique[key] = true
			}
			for _, key := range extract(expr.Condition.Right) {
				unique[key] = true
			}
		}
		keys = maps.Keys(unique)
		return
	}

	contextKeys := extract(stmnt.Expression)
	if stmnt.OrderBy != nil {
		if !beStrings.StringInSlices(*stmnt.OrderBy, contextKeys) {
			contextKeys = append(contextKeys, *stmnt.OrderBy)
		}
	} else {
		if !beStrings.StringInSlices("Url", contextKeys) {
			contextKeys = append(contextKeys, "Url")
		}
	}
	sort.Sort(natural.StringSlice(contextKeys))
	stmnt.ContextKeys = contextKeys
	return
}

// func ParticipleUnquoteRegexp(types ...string) participle.Option {
// 	if len(types) == 0 {
// 		types = []string{"Regexp"}
// 	}
// 	return participle.Map(func(t lexer.Token) (lexer.Token, error) {
// 		value, err := UnquoteRegexp(t.Value)
// 		if err != nil {
// 			return t, participle.Errorf(t.Pos, "invalid regexp string %q: %s", t.Value, err.Error())
// 		}
// 		t.Value = value
// 		return t, nil
// 	}, types...)
// }