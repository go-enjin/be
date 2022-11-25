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

	"github.com/go-enjin/be/pkg/pagecache"
	"github.com/go-enjin/be/pkg/pageql"
	"github.com/go-enjin/be/pkg/regexps"
)

func (f *CFeature) processQuery(input string) (matched []*pagecache.Stub, err error) {
	var expr *pageql.Expression
	if expr, err = pageql.Compile(input); err != nil {
		return
	}
	matched, err = f.processQueryExpr(expr)
	return
}

func (f *CFeature) processQueryExpr(expr *pageql.Expression) (matched []*pagecache.Stub, err error) {
	var chained bool
	if chained, err = pageql.ValidateGrouping(expr); err != nil {
		return
	}

	if chained {
		var lhsMatched, rhsMatched []*pagecache.Stub
		if lhsMatched, err = f.processQueryExpr(expr.Lhs.SubExpression); err != nil {
			return
		}
		if rhsMatched, err = f.processQueryExpr(expr.Tail[0].Rhs.SubExpression); err != nil {
			return
		}
		if expr.Tail[0].Op == "AND" {
			for _, lhStub := range lhsMatched {
				for _, rhStub := range rhsMatched {
					if lhStub.Shasum == rhStub.Shasum {
						matched = append(matched, lhStub)
						break
					}
				}
			}
		} else {
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

	key := ""
	switch {

	case expr.Lhs.SubExpression != nil:
		return f.processQueryExpr(expr.Lhs.SubExpression)

	case expr.Lhs.ContextKey != nil:
		key = *expr.Lhs.ContextKey
		for _, op := range expr.Tail {
			if matched, err = f.processOp(key, op); err != nil {
				return
			}
		}

	default:
		err = fmt.Errorf("only context keys can be on the left-hand side of an expression")
		return
	}

	return
}

func (f *CFeature) processOp(key string, op *pageql.Operation) (matched []*pagecache.Stub, err error) {
	switch op.Op {
	case "==":
		matched, err = f.processOpRhs(key, op.Rhs, true)

	case "!=":
		matched, err = f.processOpRhs(key, op.Rhs, false)

	default:
		err = fmt.Errorf("unsupported operation: %v", op.Op)
	}
	return
}

func (f *CFeature) processOpRhs(key string, opValue *pageql.Value, inclusive bool) (matched []*pagecache.Stub, err error) {
	results := make(map[string]*pagecache.Stub)

	switch {
	case opValue.Regexp != nil:
		var rx *regexp.Regexp
		if rx, err = regexps.Compile(*opValue.Regexp); err != nil {
			err = fmt.Errorf("error compiling regular expression: %v", err)
			return
		}
		if values, ok := f.index[key]; ok {
			for value, stubs := range values {
				switch t := value.(type) {
				case string:
					match := rx.MatchString(t)
					if (inclusive && match) || (!inclusive && !match) {
						for _, stub := range stubs {
							results[stub.Shasum] = stub
						}
					}
				default:
					err = fmt.Errorf("%v is a %T, regular expressions expect strings", key, value)
					return
				}
			}
		}

	case opValue.String != nil:
		if values, ok := f.index[key]; ok {
			for value, stubs := range values {
				switch t := value.(type) {
				case string:
					match := t == *opValue.String
					if (inclusive && match) || (!inclusive && !match) {
						for _, stub := range stubs {
							results[stub.Shasum] = stub
						}
					}
				default:
					// TODO: implement more than string and regexp comparisons
				}
			}
		}

	case opValue.SubExpression != nil:
		err = fmt.Errorf("sub-expressions are not permitted as comparison values")

	default:
		err = fmt.Errorf("unsupported comparison value type for: %v", key)
	}

	for _, stub := range results {
		matched = append(matched, stub)
	}

	return
}