//go:build page_pql || pages || all

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

package selector

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/puzpuzpuz/xsync/v2"

	"github.com/go-enjin/be/pkg/cmp"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/pageql"
	"github.com/go-enjin/be/pkg/regexps"
	"github.com/go-enjin/be/pkg/slices"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/types/page"
)

type cSelector struct {
	input  string
	feat   feature.PageContextProvider
	theme  feature.Theme
	sel    *pageql.Selection
	cache  *xsync.MapOf[string, map[string]interface{}]
	lookup *xsync.MapOf[string, map[interface{}][]string]
	count  *xsync.MapOf[string, uint64]

	err error
}

func NewProcessWith(input string, t feature.Theme, f feature.PageContextProvider) (selected map[string]interface{}, err error) {
	matcher := &cSelector{
		feat:   f,
		theme:  t,
		input:  input,
		cache:  xsync.NewMapOf[map[string]interface{}](),
		lookup: xsync.NewMapOf[map[interface{}][]string](),
		count:  xsync.NewMapOf[uint64](),
	}
	if matcher.feat == nil {
		err = fmt.Errorf("pageql matcher process requires a pagecache.PageContentProvider feature to be present")
		return
	} else if matcher.theme == nil {
		err = fmt.Errorf("pageql matcher process requires an enjin theme to be present")
		return
	}
	selected, err = matcher.process()
	return
}

func (m *cSelector) process() (selected map[string]interface{}, err error) {
	var pErr *pageql.ParseError
	if m.sel, pErr = pageql.CompileSelect(m.input); pErr != nil {
		err = error(pErr)
		return
	}

	for _, sel := range m.sel.Selecting {

		if _, exists := m.lookup.Load(sel.ContextKey); !exists {
			m.lookup.Store(sel.ContextKey, make(map[interface{}][]string))
		}

		if sel.Count {
			m.count.Store(sel.ContextKey, m.feat.PageContextValuesCount(sel.ContextKey))
		}

		for pair := range m.feat.YieldPageContextValueStubs(sel.ContextKey) {
			kvp, _ := m.lookup.Load(sel.ContextKey)
			shasums, _ := kvp[pair.Value]
			shasums = append(shasums, pair.Stub.Shasum)
			kvp[pair.Value] = shasums
			m.lookup.Store(sel.ContextKey, kvp)
		}

	}

	if m.sel.Statement != nil {

		for _, key := range m.sel.Statement.ContextKeys {
			if _, exists := m.lookup.Load(key); !exists {
				m.lookup.Store(key, make(map[interface{}][]string))
				for pair := range m.feat.YieldPageContextValueStubs(key) {
					kvp, _ := m.lookup.Load(key)
					shasums, _ := kvp[pair.Value]
					shasums = append(shasums, pair.Stub.Shasum)
					kvp[pair.Value] = shasums
					m.lookup.Store(key, kvp)
				}
			}
		}

		selected, err = m.processWithStatement()
		return
	}

	selected, err = m.processWithoutStatement()
	return
}

func (m *cSelector) processWithoutStatement() (selected map[string]interface{}, err error) {

	// handle distinct and counted values
	distinct := make(map[string]interface{})
	for _, sel := range m.sel.Selecting {
		if !sel.Distinct && !sel.Count {
			continue
		}
		kvp, _ := m.lookup.Load(sel.ContextKey)
		values := maps.AnyKeys(kvp)
		count, _ := m.count.Load(sel.ContextKey)
		switch {
		case !sel.Count && sel.Random:
			idx := rand.Intn(len(values))
			distinct[sel.ContextKey] = values[idx]
		case sel.Count && sel.Random:
			if count >= 1 {
				distinct[sel.ContextKey] = 1
			} else {
				distinct[sel.ContextKey] = 0
			}
		case sel.Count && !sel.Random:
			distinct[sel.ContextKey] = count
		case sel.Distinct:
			if count == 1 {
				// collapse
				distinct[sel.ContextKey] = values[0]
			} else {
				distinct[sel.ContextKey] = values
			}
		}
	}

	// handle standard (indistinct and uncounted) and random values
	simples := make(map[string][]interface{})
	randoms := make(map[string]interface{})
	for _, sel := range m.sel.Selecting {
		if sel.Random {
			kvp, _ := m.lookup.Load(sel.ContextKey)
			values := maps.AnyKeys(kvp)
			if lv := len(values); lv > 1 {
				idx := rand.Intn(len(values))
				randoms[sel.ContextKey] = values[idx]
			} else if lv == 1 {
				randoms[sel.ContextKey] = values[0]
			} else {
				randoms[sel.ContextKey] = nil
			}
		} else if !sel.Distinct && !sel.Count {
			kvp, _ := m.lookup.Load(sel.ContextKey)
			for _, value := range maps.AnyKeys(kvp) {
				simples[sel.ContextKey] = append(simples[sel.ContextKey], value)
			}
		}
	}

	// prepare return value
	selected = make(map[string]interface{})

	// add selected
	for k, v := range simples {
		selected[k] = v
	}

	// add random
	for k, v := range randoms {
		switch t := v.(type) {
		default:
			selected[k] = t
		}
	}

	// add distinct
	for k, v := range distinct {
		switch t := v.(type) {
		default:
			selected[k] = t
		}
	}
	return
}

func (m *cSelector) processWithStatement() (selected map[string]interface{}, err error) {
	if m.sel.Statement != nil {
		var matched feature.PageStubs
		if matched, err = m.processQueryStatement(m.sel.Statement.Render()); err != nil {
			log.ErrorF("pqs error: %v", err)
			return
		}
		var matchedShasums []string
		for _, stub := range matched {
			matchedShasums = append(matchedShasums, stub.Shasum)
		}

		// build up lists of unique values

		distinct := make(map[string]interface{})
		for _, sel := range m.sel.Selecting {
			if !sel.Distinct && !sel.Count {
				continue
			}
			var values []interface{}
			kvp, _ := m.lookup.Load(sel.ContextKey)
			for value, shasums := range kvp {
				if beStrings.AnyStringsInStrings(matchedShasums, shasums) {
					switch vt := value.(type) {
					case []string:
						for _, vtv := range vt {
							values = append(values, vtv)
						}
					case []interface{}:
						for _, vtv := range vt {
							values = append(values, vtv)
						}
					default:
						values = append(values, vt)
					}
				}
			}
			count, _ := m.count.Load(sel.ContextKey)
			switch {
			case !sel.Count && sel.Random:
				idx := rand.Intn(len(values))
				distinct[sel.ContextKey] = values[idx]
			case sel.Count && sel.Random:
				if count >= 1 {
					distinct[sel.ContextKey] = 1
				} else {
					distinct[sel.ContextKey] = 0
				}
			case sel.Count && !sel.Random:
				distinct[sel.ContextKey] = count
			case len(values) == 1:
				// collapse
				distinct[sel.ContextKey] = values[0]
			default:
				distinct[sel.ContextKey] = values
			}
		}

		randoms := make(map[string]interface{})
		for _, sel := range m.sel.Selecting {
			if sel.Random && !sel.Distinct && !sel.Count {
				kvp, _ := m.lookup.Load(sel.ContextKey)
				values := maps.AnyKeys(kvp)
				idx := rand.Intn(len(values))
				randoms[sel.ContextKey] = values[idx]
			}
		}

		// build a map where there is one value per match, for each context-key
		simples := make(map[string][]interface{})
		for _, stub := range matched {
			for _, sel := range m.sel.Selecting {
				if sel.Distinct || sel.Random || sel.Count {
					continue
				}
				found := false
				kvp, _ := m.lookup.Load(sel.ContextKey)
				for value, shasums := range kvp {
					if found = slices.Within(stub.Shasum, shasums); found {
						simples[sel.ContextKey] = append(simples[sel.ContextKey], value)
						break
					}
				}
				if !found {
					simples[sel.ContextKey] = append(simples[sel.ContextKey], nil)
				}
			}
		}

		selected = make(map[string]interface{})
		for k, v := range randoms {
			selected[k] = v
		}
		for k, v := range simples {
			selected[k] = v
		}
		for k, v := range distinct {
			selected[k] = v
		}
	}

	return
}

func (m *cSelector) processQueryStatement(stmnt *pageql.Statement) (matched []*feature.PageStub, err error) {
	if stmnt.Expression != nil {
		if matched, err = m.processQueryExpression(stmnt.Expression); err != nil {
			return
		}
	}
	return
}

func (m *cSelector) processQueryExpression(expr *pageql.Expression) (matched []*feature.PageStub, err error) {
	switch {
	case expr.Condition != nil:
		matched, err = m.processQueryCondition(expr.Condition)

	case expr.Operation != nil:
		matched, err = m.processQueryOperation(expr.Operation)
	}
	return
}

func (m *cSelector) processQueryCondition(cond *pageql.Condition) (matched []*feature.PageStub, err error) {

	var lhsMatched, rhsMatched []*feature.PageStub
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
		add := make(map[string]*feature.PageStub)
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

func (m *cSelector) processQueryOperation(op *pageql.Operation) (matched []*feature.PageStub, err error) {
	switch op.Type {
	case "==":
		matched, err = m.processOperationEquals(*op.Left, op.Right, true)

	case "!=":
		matched, err = m.processOperationEquals(*op.Left, op.Right, false)

	default:
		err = fmt.Errorf("unsupported operation: %v", op.Type)
	}
	return
}

func (m *cSelector) processOperationEquals(key string, opValue *pageql.Value, inclusive bool) (matched []*feature.PageStub, err error) {
	results := make(map[string]*feature.PageStub)

	// TODO: implement more than string and regexp comparisons

	switch {
	case opValue.Regexp != nil:
		var rx *regexp.Regexp
		if rx, err = regexps.Compile(*opValue.Regexp); err != nil {
			err = fmt.Errorf("error compiling regular expression: %v", err)
			return
		}
		if values, ok := m.lookup.Load(key); ok {
			for value, shasums := range values {
				switch t := value.(type) {
				case string:
					match := rx.MatchString(t)
					if (inclusive && match) || (!inclusive && !match) {
						for _, shasum := range shasums {
							results[shasum] = m.feat.FindPageStub(shasum)
							p, _ := page.NewPageFromStub(results[shasum], m.theme)
							ctx := p.Context().Copy()
							ctx.CamelizeKeys()
							m.cache.Store(shasum, ctx.Select(m.sel.Statement.ContextKeys...))
						}
					}
				default:
					err = fmt.Errorf("page.%v is a %T, regular expressions expect strings", key, value)
					return
				}
			}
		}

	case opValue.String != nil:
		if values, ok := m.lookup.Load(key); ok {
			for value, shasums := range values {
				if match, ee := cmp.Compare(value, *opValue.String); ee != nil {
					err = ee
					return
				} else {
					if (inclusive && match) || (!inclusive && !match) {
						for _, shasum := range shasums {
							results[shasum] = m.feat.FindPageStub(shasum)
							p, _ := page.NewPageFromStub(results[shasum], m.theme)
							ctx := p.Context().Copy()
							ctx.CamelizeKeys()
							m.cache.Store(shasum, ctx.Select(m.sel.Statement.ContextKeys...))
						}
					}
				}
			}
		}

	default:
		err = fmt.Errorf("unsupported opValue type case: %#+v", opValue)
		return

	}

	for _, stub := range results {
		matched = append(matched, stub)
	}

	return
}