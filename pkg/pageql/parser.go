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
// TODO: decide how to use `=~` and `!~`, ie: regexp op instead of regexp pattern string compare? or is it a form of fuzzy matching? list-vs-string?
// TODO: implement <, >, <=, >= and general numerical comparison supports
// TODO: eventually implement simple math operations: +, -, /, * and this must support correct operator precedence

const (
	Ident      = `\b([a-zA-Z][.a-zA-Z0-9]*)\b`
	Int        = `\b(\d+)\b`
	Float      = `\b(\d*\.\d+)\b`
	Number     = `[-+]?\d*\.?\d+([eE][-+]?\d+)?`
	String     = `'[^']*'|"[^"]*"`
	Regexp     = `/(.+?)/|\!(.+?)\!|\@(.+?)\@|\~(.+?)\~`
	Operators  = `==|=~|!=|!~|[.,()]`
	Whitespace = `\s+`
)

var (
	stmntParser = participle.MustBuild[Statement](
		participle.Lexer(lexer.MustSimple([]lexer.SimpleRule{
			{`Keyword`, `(?i)\b(BY|ORDER|LIMIT|OFFSET|TRUE|FALSE|NULL|IS|NOT|AND|OR|IN|ASC|DSC|DESC)\b`},
			{`Ident`, Ident},
			{`Int`, Int},
			{`Float`, Float},
			{`Number`, Number},
			{`String`, String},
			{`Regexp`, Regexp},
			{`Operators`, Operators},
			{`whitespace`, Whitespace},
		})),
		// UnquoteRegexp("Regexp"),
		// participle.Unquote("String"),
		participle.CaseInsensitive("Keyword"),
	)

	selParser = participle.MustBuild[Selection](
		participle.Lexer(lexer.MustSimple([]lexer.SimpleRule{
			{`Keyword`, `(?i)\b(SELECT|COUNT|DISTINCT|WITHIN|BY|ORDER|LIMIT|OFFSET|TRUE|FALSE|NULL|IS|NOT|AND|OR|IN|ASC|DSC|DESC)\b`},
			{`Ident`, Ident},
			{`Int`, Int},
			{`Float`, Float},
			{`Number`, Number},
			{`String`, String},
			{`Regexp`, Regexp},
			{`Operators`, Operators},
			{`whitespace`, Whitespace},
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

func parseQueryString(query string) (stmnt *Statement, err *ParseError) {
	query = SanitizeQuery(query)

	var participleError error
	if stmnt, participleError = stmntParser.ParseString("pageql", query); participleError != nil {
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

func parseSelectString(query string) (sel *Selection, err *ParseError) {
	query = SanitizeQuery(query)

	var participleError error
	if sel, participleError = selParser.ParseString("pageql", query); participleError != nil {
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

	if sel.Statement != nil {
		contextKeys := extract(sel.Statement.Expression)
		if sel.Statement.OrderBy != nil {
			if !beStrings.StringInSlices(*sel.Statement.OrderBy, contextKeys) {
				contextKeys = append(contextKeys, *sel.Statement.OrderBy)
			}
		} else {
			if !beStrings.StringInSlices("Url", contextKeys) {
				contextKeys = append(contextKeys, "Url")
			}
		}
		sort.Sort(natural.StringSlice(contextKeys))
		sel.Statement.ContextKeys = contextKeys
	}
	return
}