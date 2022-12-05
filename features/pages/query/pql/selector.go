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

package pql

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-enjin/be/pkg/cmp"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/pagecache"
	"github.com/go-enjin/be/pkg/pageql"
	"github.com/go-enjin/be/pkg/regexps"
	"github.com/go-enjin/be/pkg/theme"
)

// TODO: see pageql.Match for new process

type CSelector struct {
	input string
	feat  *CFeature
	theme *theme.Theme
	sel   *pageql.Selection
	cache map[string]map[string]interface{}

	err error
}

func newSelectorProcess(input string, f *CFeature) (selected map[string]interface{}, err error) {
	matcher := &CSelector{
		feat:  f,
		input: input,
		cache: make(map[string]map[string]interface{}),
	}
	if matcher.theme, err = f.enjin.GetTheme(); err != nil {
		return
	}
	selected, err = matcher.process()
	return
}

func (m *CSelector) process() (selected map[string]interface{}, err error) {
	var pErr *pageql.ParseError
	if m.sel, pErr = pageql.CompileSelect(m.input); pErr != nil {
		err = error(pErr)
		return
	}

	if m.sel.Statement != nil {
		selected, err = m.processWithStatement()
		return
	}

	selected, err = m.processWithoutStatement()
	return
}

func (m *CSelector) processWithoutStatement() (selected map[string]interface{}, err error) {

	distinct := make(map[string]interface{})
	for _, sel := range m.sel.Selecting {
		if !sel.Distinct && !sel.Count {
			continue
		}
		var values []interface{}
		for value, _ := range m.feat.index[sel.ContextKey] {
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
		if sel.Count {
			distinct[sel.ContextKey] = len(values)
		} else if len(values) == 1 {
			distinct[sel.ContextKey] = values[0]
		} else {
			distinct[sel.ContextKey] = values
		}
	}

	shasumLookup := make(map[string]map[string]interface{})
	for _, sel := range m.sel.Selecting {
		if sel.Distinct || sel.Count {
			continue
		}
		for value, stubs := range m.feat.index[sel.ContextKey] {
			for _, stub := range stubs {
				if _, exists := shasumLookup[stub.Shasum]; !exists {
					shasumLookup[stub.Shasum] = make(map[string]interface{})
				}
				shasumLookup[stub.Shasum][sel.ContextKey] = value
			}
		}
	}

	selectLookup := make(map[string][]interface{})
	for _, values := range shasumLookup {
		for k, v := range values {
			selectLookup[k] = append(selectLookup[k], v)
		}
	}

	selected = make(map[string]interface{})
	for k, v := range selectLookup {
		selected[k] = v
	}
	for k, v := range distinct {
		selected[k] = v
	}
	return
}

func (m *CSelector) processWithStatement() (selected map[string]interface{}, err error) {
	if m.sel.Statement != nil {
		var matched pagecache.Stubs
		if matched, err = m.processQueryStatement(m.sel.Statement.Render()); err != nil {
			log.ErrorF("pqs error: %v", err)
			return
		}

		// build up lists of unique values

		distinct := make(map[string]interface{})
		for _, sel := range m.sel.Selecting {
			if !sel.Distinct && !sel.Count {
				continue
			}
			var values []interface{}
			for value, stubs := range m.feat.index[sel.ContextKey] {
				if pagecache.AnyStubsInStubs(matched, stubs) {
					switch vt := value.(type) {
					case []interface{}:
						for _, vtv := range vt {
							values = append(values, vtv)
						}
					default:
						values = append(values, vt)
					}
				}
			}
			if sel.Count {
				distinct[sel.ContextKey] = len(values)
			} else if len(values) == 1 {
				distinct[sel.ContextKey] = values[0]
			} else {
				distinct[sel.ContextKey] = values
			}
		}

		// build a map where there is one value per match, for each context-key

		lookup := make(map[string][]interface{})
		for _, stub := range matched {
			for _, sel := range m.sel.Selecting {
				if sel.Distinct || sel.Count {
					continue
				}
				found := false
				for value, stubs := range m.feat.index[sel.ContextKey] {
					if found = stubs.HasShasum(stub.Shasum); found {
						lookup[sel.ContextKey] = append(lookup[sel.ContextKey], value)
						break
					}
				}
				if !found {
					lookup[sel.ContextKey] = append(lookup[sel.ContextKey], nil)
				}
			}
		}

		selected = make(map[string]interface{})
		for k, v := range lookup {
			selected[k] = v
		}
		for k, v := range distinct {
			selected[k] = v
		}
	}

	return
}

func (m *CSelector) processQueryStatement(stmnt *pageql.Statement) (matched []*pagecache.Stub, err error) {
	if stmnt.Expression != nil {
		if matched, err = m.processQueryExpression(stmnt.Expression); err != nil {
			return
		}
	}
	return
}

func (m *CSelector) processQueryExpression(expr *pageql.Expression) (matched []*pagecache.Stub, err error) {
	switch {
	case expr.Condition != nil:
		matched, err = m.processQueryCondition(expr.Condition)

	case expr.Operation != nil:
		matched, err = m.processQueryOperation(expr.Operation)
	}
	return
}

func (m *CSelector) processQueryCondition(cond *pageql.Condition) (matched []*pagecache.Stub, err error) {

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

func (m *CSelector) processQueryOperation(op *pageql.Operation) (matched []*pagecache.Stub, err error) {
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

func (m *CSelector) processOperationEquals(key string, opValue *pageql.Value, inclusive bool) (matched []*pagecache.Stub, err error) {
	results := make(map[string]*pagecache.Stub)

	// TODO: implement more than string and regexp comparisons

	switch {
	case opValue.Regexp != nil:
		var rx *regexp.Regexp
		if rx, err = regexps.Compile(*opValue.Regexp); err != nil {
			err = fmt.Errorf("error compiling regular expression: %v", err)
			return
		}
		if values, ok := m.feat.index[key]; ok {
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
							m.cache[stub.Shasum] = ctx.Select(m.sel.Statement.ContextKeys...)
						}
					}
				default:
					err = fmt.Errorf("page.%v is a %T, regular expressions expect strings", key, value)
					return
				}
			}
		}

	case opValue.String != nil:
		if values, ok := m.feat.index[key]; ok {
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
							m.cache[stub.Shasum] = ctx.Select(m.sel.Statement.ContextKeys...)
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