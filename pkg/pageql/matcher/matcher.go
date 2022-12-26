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

package matcher

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/go-enjin/be/pkg/cmp"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/pagecache"
	"github.com/go-enjin/be/pkg/pageql"
	"github.com/go-enjin/be/pkg/regexps"
	"github.com/go-enjin/be/pkg/theme"
)

type cMatcher struct {
	input string
	feat  pagecache.PageContextProvider
	theme *theme.Theme
	stmnt *pageql.Statement
	cache map[string]map[string]interface{}

	limit   int
	offset  int
	orderBy string
	sortDir string

	err error
}

func NewProcess(input string, enjin feature.Internals) (matched []*pagecache.Stub, err error) {
	var ok bool
	var t *theme.Theme
	var f pagecache.PageContextProvider
	if t, err = enjin.GetTheme(); err != nil {
		return
	}
	for _, feat := range enjin.Features() {
		if f, ok = feat.(pagecache.PageContextProvider); ok {
			break
		}
	}
	matched, err = NewProcessWith(input, t, f)
	return
}

func NewProcessWith(input string, t *theme.Theme, f pagecache.PageContextProvider) (matched []*pagecache.Stub, err error) {
	matcher := &cMatcher{
		feat:    f,
		theme:   t,
		input:   input,
		cache:   make(map[string]map[string]interface{}),
		limit:   -1,
		offset:  0,
		orderBy: "Url",
		sortDir: "ASC",
	}
	if matcher.feat == nil {
		err = fmt.Errorf("pageql matcher process requires a pagecache.PageContentProvider feature to be present")
		return
	} else if matcher.theme == nil {
		err = fmt.Errorf("pageql matcher process requires an enjin theme to be present")
		return
	}
	matched, err = matcher.process()
	return
}

func (m *cMatcher) process() (matched []*pagecache.Stub, err error) {
	var pErr *pageql.ParseError
	if m.stmnt, pErr = pageql.CompileQuery(m.input); pErr != nil {
		err = error(pErr)
		return
	}
	if m.stmnt.Limit != nil {
		m.limit = *m.stmnt.Limit
	}
	if m.limit == 0 {
		return
	}
	if m.stmnt.Offset != nil {
		if m.offset = *m.stmnt.Offset; m.offset < 0 {
			m.offset = 0
		}
	}
	if m.stmnt.OrderBy != nil {
		m.orderBy = *m.stmnt.OrderBy
	}
	if m.stmnt.SortDir != nil {
		m.sortDir = strings.ToUpper(*m.stmnt.SortDir)
	}

	if matched, err = m.processQueryStatement(m.stmnt.Render()); err != nil {
		log.ErrorF("pqs error: %v", err)
		return
	}

	sort.Slice(matched, func(i, j int) (less bool) {
		if err != nil {
			// nop, skip out of sort
			return
		}

		a, b := matched[i], matched[j]
		ac, bc := m.cache[a.Shasum], m.cache[b.Shasum]
		av, aok := ac[m.orderBy]
		bv, bok := bc[m.orderBy]

		// log.WarnF("sorting (%v %v): %v < %v", m.orderBy, m.sortDir, av, bv)

		if !aok || !bok {
			less = aok
			return
		}
		if less, err = cmp.Less(av, bv); err == nil && m.sortDir != "ASC" {
			less = !less
		}
		// log.WarnF("sorted (%v %v): %v < %v", m.orderBy, m.sortDir, av, bv, less)
		return
	})
	if err != nil {
		matched = nil
		return
	}

	if m.limit > 0 {
		matched = matched[m.offset : m.offset+m.limit]
	}
	return
}

func (m *cMatcher) processQueryStatement(stmnt *pageql.Statement) (matched []*pagecache.Stub, err error) {
	if stmnt.Expression != nil {
		if matched, err = m.processQueryExpression(stmnt.Expression); err != nil {
			return
		}
	}
	return
}

func (m *cMatcher) processQueryExpression(expr *pageql.Expression) (matched []*pagecache.Stub, err error) {
	switch {
	case expr.Condition != nil:
		matched, err = m.processQueryCondition(expr.Condition)

	case expr.Operation != nil:
		matched, err = m.processQueryOperation(expr.Operation)
	}
	return
}

func (m *cMatcher) processQueryCondition(cond *pageql.Condition) (matched []*pagecache.Stub, err error) {

	var lhsMatched, rhsMatched []*pagecache.Stub
	if lhsMatched, err = m.processQueryExpression(cond.Left); err != nil {
		return
	}
	if rhsMatched, err = m.processQueryExpression(cond.Right); err != nil {
		return
	}

	switch strings.ToUpper(cond.Type) {
	case "AND":
		for _, lhStub := range lhsMatched {
			for _, rhStub := range rhsMatched {
				if lhStub.Shasum == rhStub.Shasum {
					matched = append(matched, lhStub)
					break
				}
			}
		}

	case "OR":
		add := make(map[string]*pagecache.Stub)
		for _, stub := range lhsMatched {
			add[stub.Shasum] = stub
		}
		for _, stub := range rhsMatched {
			add[stub.Shasum] = stub
		}
		for _, stub := range add {
			matched = append(matched, stub)
		}
	}
	return
}

func (m *cMatcher) processQueryOperation(op *pageql.Operation) (matched []*pagecache.Stub, err error) {
	switch op.Type {
	case "==", "=~":
		matched, err = m.processOperationEquals(*op.Left, op.Right, true)

	case "!=", "!~":
		matched, err = m.processOperationEquals(*op.Left, op.Right, false)

	// case "=~":
	// 	matched, err = m.processOperationContains(*op.Left, op.Right, true)

	// case "!~":
	// 	matched, err = m.processOperationContains(*op.Left, op.Right, false)

	default:
		err = fmt.Errorf("unsupported operation: %v", op.Type)
	}
	return
}

func (m *cMatcher) processOperationEquals(key string, opValue *pageql.Value, inclusive bool) (matched []*pagecache.Stub, err error) {
	results := make(map[string]*pagecache.Stub)

	// TODO: implement more than string and regexp comparisons

	switch {
	case opValue.Regexp != nil:
		var rx *regexp.Regexp
		if rx, err = regexps.Compile(*opValue.Regexp); err != nil {
			err = fmt.Errorf("error compiling regular expression: %v", err)
			return
		}
		var values map[interface{}]pagecache.Stubs
		if values, err = m.feat.GetPageContextValueStubs(key); err != nil {
			err = fmt.Errorf("error getting context key values: %v", err)
			return
		} else {
			for value, stubs := range values {
				switch t := value.(type) {
				case string:
					match := rx.MatchString(t)
					if (inclusive && match) || (!inclusive && !match) {
						for _, stub := range stubs {
							results[stub.Shasum] = stub
							p, _ := stub.Make(m.theme)
							ctx := p.Context.Copy()
							ctx.CamelizeKeys()
							m.cache[stub.Shasum] = ctx.Select(m.stmnt.ContextKeys...)
							// log.WarnF("m.cache[%v] = %v", stub.Shasum, m.cache[stub.Shasum])
						}
					}
				default:
					err = fmt.Errorf("page.%v is a %T, regular expressions expect strings", key, value)
					return
				}
			}
		}

	case opValue.String != nil:
		var values map[interface{}]pagecache.Stubs
		if values, err = m.feat.GetPageContextValueStubs(key); err != nil {
			err = fmt.Errorf("error getting context key values: %v", err)
			return
		} else {
			for value, stubs := range values {
				if match, ee := cmp.Compare(value, *opValue.String); ee != nil {
					err = ee
					return
				} else {
					if (inclusive && match) || (!inclusive && !match) {
						for _, stub := range stubs {
							results[stub.Shasum] = stub
							p, _ := stub.Make(m.theme)
							ctx := p.Context.Copy()
							ctx.CamelizeKeys()
							m.cache[stub.Shasum] = ctx.Select(m.stmnt.ContextKeys...)
						}
					}
				}
			}
		}

	}

	for _, stub := range results {
		matched = append(matched, stub)
	}

	return
}