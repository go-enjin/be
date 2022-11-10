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
)

func Validate(query string) (err error) {
	var expr *Expression
	if expr, err = parser.ParseString("pageql", query); err != nil {
		return
	}
	if _, err = validateGrouping(expr); err == nil {
		err = validateProcess(expr)
	}
	return
}

func validateGrouping(expr *Expression) (chained bool, err error) {

	for _, op := range expr.Tail {
		if chained = op.Op == "AND" || op.Op == "OR"; chained {
			break
		}
	}

	if chained {
		if expr.Lhs.SubExpression == nil {
			err = fmt.Errorf("expressions must be grouped when using AND or OR operators")
			return
		}
		switch len(expr.Tail) {
		case 1:
			if expr.Tail[0].Rhs.SubExpression == nil {
				err = fmt.Errorf("expressions must be grouped when using AND or OR operators")
				return
			}
		default:
			err = fmt.Errorf("expressions must be grouped when using AND or OR operators")
			return
		}
	}

	return
}

func validateProcess(expr *Expression) (err error) {
	var chained bool
	if chained, err = validateGrouping(expr); err != nil {
		return
	}

	if chained {
		if err = validateProcess(expr.Lhs.SubExpression); err != nil {
			return
		}
		if err = validateProcess(expr.Tail[0].Rhs.SubExpression); err != nil {
			return
		}
		return
	}

	key := ""
	switch {

	case expr.Lhs.SubExpression != nil:
		return validateProcess(expr.Lhs.SubExpression)

	case expr.Lhs.ContextKey != nil:
		key = *expr.Lhs.ContextKey
		for _, op := range expr.Tail {
			if err = validateOp(key, op); err != nil {
				return
			}
		}

	default:
		err = fmt.Errorf("only context keys can be on the left-hand side of an expression")
	}

	return
}

func validateOp(key string, op *Operation) (err error) {
	switch op.Op {

	case "==", "!=":
		err = validateOpRhs(key, op.Rhs)

	default:
		err = fmt.Errorf("invalid operation: %v", op.Op)

	}
	return
}

func validateOpRhs(key string, op *Value) (err error) {
	switch {

	case op.Regexp != nil:
		if _, e := regexp.Compile(*op.Regexp); e != nil {
			err = fmt.Errorf("error compiling regular expression")
		}

	case op.SubExpression != nil:
		err = fmt.Errorf("sub-expressions are not permitted as comparison values")

	}
	return
}