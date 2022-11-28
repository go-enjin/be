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
	"strings"

	"github.com/go-enjin/be/pkg/cmp"
	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/regexps"
)

func Match(query string, ctx context.Context) (matched bool, err error) {
	var stmnt *Statement
	if stmnt, err = Compile(query); err != nil {
		return
	}
	matched, err = processQuery(stmnt.Expression, ctx)
	return
}

func processQuery(expr *Expression, ctx context.Context) (matched bool, err error) {
	switch {

	case expr.Condition != nil:
		matched, err = processQueryCondition(expr.Condition, ctx)

	case expr.Operation != nil:
		matched, err = processQueryOperation(expr.Operation, ctx)

	}
	return
}

func processQueryCondition(cond *Condition, ctx context.Context) (matched bool, err error) {
	if cond.Left != nil && cond.Right != nil {
		var leftMatch, rightMatch bool
		if leftMatch, err = processQuery(cond.Left, ctx); err != nil {
			return
		}
		if rightMatch, err = processQuery(cond.Right, ctx); err != nil {
			return
		}
		switch strings.ToUpper(cond.Type) {
		case "OR":
			matched = leftMatch || rightMatch
		case "AND":
			matched = leftMatch && rightMatch
		}
	}
	return
}

func processQueryOperation(op *Operation, ctx context.Context) (matched bool, err error) {
	switch op.Type {

	case "==":
		matched, err = processQueryOperationEquals(*op.Left, op.Right, ctx)

	case "!=":
		if matched, err = processQueryOperationEquals(*op.Left, op.Right, ctx); err == nil {
			matched = !matched
		}

	default:
		err = fmt.Errorf(`%v not implemented`, op.Type)

	}
	return
}

func processQueryOperationEquals(key string, opValue *Value, ctx context.Context) (matched bool, err error) {
	switch {

	case opValue.ContextKey != nil:
		lValue := ctx.Get(key)
		rValue := ctx.Get(*opValue.ContextKey)
		matched, err = cmp.Compare(lValue, rValue)

	case opValue.Regexp != nil:
		if value, ok := ctx.Get(key).(string); ok {
			if rx, e := regexps.Compile(*opValue.Regexp); e != nil {
				err = fmt.Errorf("error compiling regular expression")
			} else {
				matched = rx.MatchString(value)
			}
		} else {
			err = fmt.Errorf("page.%v is of type %T, expected string", key, ctx.Get(key))
		}

	case opValue.String != nil:
		if value, ok := ctx.Get(key).(string); ok {
			matched = value == *opValue.String
		} else {
			err = fmt.Errorf("page.%v is of type %T, expected string", key, ctx.Get(key))
		}

	}
	return
}