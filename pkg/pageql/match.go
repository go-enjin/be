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
	"fmt"
	"regexp"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/log"
)

func Match(query string, ctx context.Context) (matched bool, err error) {
	var expr *Expression
	if expr, err = parser.ParseString("pageql", query); err != nil {
		return
	}
	matched, err = process(expr, ctx)
	return
}

func process(expr *Expression, ctx context.Context) (matched bool, err error) {

	var chained bool
	if chained, err = validateGrouping(expr); err != nil {
		return
	}

	if chained {
		var lhsMatched, rhsMatched bool
		if lhsMatched, err = process(expr.Lhs.SubExpression, ctx); err != nil {
			return
		}
		if rhsMatched, err = process(expr.Tail[0].Rhs.SubExpression, ctx); err != nil {
			return
		}
		if expr.Tail[0].Op == "AND" {
			matched = lhsMatched && rhsMatched
		} else {
			matched = lhsMatched || rhsMatched
		}
		return
	}

	key := ""
	switch {

	case expr.Lhs.SubExpression != nil:
		return process(expr.Lhs.SubExpression, ctx)

	case expr.Lhs.ContextKey != nil:
		key = *expr.Lhs.ContextKey
		for _, op := range expr.Tail {
			if matched, err = processOp(key, op, ctx); err != nil || !matched {
				return
			}
		}

	default:
		err = fmt.Errorf("only context keys can be on the left-hand side of an expression")
		return
	}

	return
}

func processOp(key string, op *Operation, ctx context.Context) (matched bool, err error) {
	switch op.Op {

	case "==":
		matched, err = processOpRhs(key, op.Rhs, ctx)

	case "!=":
		if matched, err = processOpRhs(key, op.Rhs, ctx); err == nil {
			matched = !matched
		}

	default:
		log.WarnF("uncaught op.Op: %v", op.Op)

	}
	return
}

func processOpRhs(key string, opValue *Value, ctx context.Context) (matched bool, err error) {
	switch {

	case opValue.Regexp != nil:
		if value, ok := ctx.Get(key).(string); ok {
			if rx, e := regexp.Compile(*opValue.Regexp); e != nil {
				err = fmt.Errorf("error compiling regular expression")
				return
			} else if matched = rx.MatchString(value); matched {
				// log.DebugF("page.%v is a string and does match m!%v!", key, *opValue.Regexp)
				return
			}
		} else {
			err = fmt.Errorf("page.%v is %T, expected string", key, ctx.Get(key))
			return
		}

	case opValue.String != nil:
		if value, ok := ctx.Get(key).(string); ok {
			if matched = value == *opValue.String; matched {
				// log.DebugF("page.%v is a string and is equal to %v", key, *opValue.String)
				return
			}
		} else {
			err = fmt.Errorf("page.%v is %T, expected string", key, ctx.Get(key))
			return
		}

	case opValue.SubExpression != nil:
		err = fmt.Errorf("sub-expressions are not permitted as comparison values")

	}
	return
}