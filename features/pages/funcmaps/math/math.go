// Copyright (c) 2023  The Go-Enjin Authors
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

package math

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Pramod-Devireddy/go-exprtk"
	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "pages-funcmaps-math"

type Feature interface {
	feature.Feature
	feature.FuncMapProvider
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	return
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	return
}

func (f *CFeature) Shutdown() {

}

func (f *CFeature) MakeFuncMap(ctx beContext.Context) (fm feature.FuncMap) {
	fm = feature.FuncMap{
		"numberAsInt": NumberAsInt,
		"add":         Add,
		"sub":         Sub,
		"mul":         Mul,
		"div":         Div,
		"mod":         Mod,
		"eval":        Evaluate,
		"addFloat":    AddFloat,
		"subFloat":    SubFloat,
		"mulFloat":    MulFloat,
		"divFloat":    DivFloat,
		"evalFloat":   EvalFloat,
	}
	return
}

func Int(v interface{}) (i int64) {
	switch t := v.(type) {
	case int:
		i = int64(t)
	case int8:
		i = int64(t)
	case int16:
		i = int64(t)
	case int32:
		i = int64(t)
	case int64:
		i = t
	case uint:
		i = int64(t)
	case uint8:
		i = int64(t)
	case uint16:
		i = int64(t)
	case uint32:
		i = int64(t)
	case uint64:
		i = int64(t)
	case float32:
		i = int64(t)
	case float64:
		i = int64(t)
	case string:
		if ii, err := strconv.Atoi(strings.TrimSpace(t)); err == nil {
			i = int64(ii)
		} else {
			log.ErrorF("error parsing string integer: %v", err)
		}
	}
	return
}

func Float(v interface{}) (i float64) {
	switch t := v.(type) {
	case int:
		i = float64(t)
	case int8:
		i = float64(t)
	case int16:
		i = float64(t)
	case int32:
		i = float64(t)
	case int64:
		i = float64(t)
	case uint:
		i = float64(t)
	case uint8:
		i = float64(t)
	case uint16:
		i = float64(t)
	case uint32:
		i = float64(t)
	case uint64:
		i = float64(t)
	case float32:
		i = float64(t)
	case float64:
		i = t
	case string:
		var err error
		if i, err = strconv.ParseFloat(strings.TrimSpace(t), 64); err != nil {
			log.ErrorF("error parsing string float: %v - %v", t, err)
		}
	}
	return
}

func Add(values ...interface{}) (result int64) {
	for _, v := range values {
		value := Int(v)
		result += value
	}
	return
}

func Sub(values ...interface{}) (result int64) {
	if len(values) == 0 {
		return
	}
	for idx, v := range values {
		value := Int(v)
		if idx == 0 {
			result = value
		} else {
			result -= value
		}
	}
	return
}

func Mul(a, b interface{}) (result int64) {
	result = Int(a) * Int(b)
	return
}

func Div(a, b interface{}) (result int64, err error) {
	ia, ib := Int(a), Int(b)
	if ib == 0 {
		err = fmt.Errorf("divide by zero")
		return
	}
	result = ia / ib
	return
}

func Mod(a, b interface{}) (result int64) {
	result = Int(a) % Int(b)
	return
}

func AddFloat(values ...interface{}) (result float64) {
	for _, v := range values {
		value := Float(v)
		result += value
	}
	return
}

func SubFloat(values ...interface{}) (result float64) {
	if len(values) == 0 {
		return
	}
	for idx, v := range values {
		value := Float(v)
		if idx == 0 {
			result = value
		} else {
			result -= value
		}
	}
	return
}

func MulFloat(a, b interface{}) (result float64) {
	result = Float(a) * Float(b)
	return
}

func DivFloat(a, b interface{}) (result float64, err error) {
	ia, ib := Float(a), Float(b)
	if ib == 0 {
		err = fmt.Errorf("divide by zero")
		return
	}
	result = ia / ib
	return
}

func NumberAsInt(v interface{}) (value int) {
	s := fmt.Sprintf("%v", v)
	value, _ = strconv.Atoi(s)
	return
}

func Evaluate(format string, argv ...interface{}) (output int, err error) {
	if v, e := EvalFloat(format, argv...); e != nil {
		err = e
	} else {
		output = int(v)
	}
	return
}

func EvalFloat(format string, argv ...interface{}) (output float64, err error) {
	expr := fmt.Sprintf(format, argv...)
	exp := exprtk.NewExprtk()
	defer exp.Delete()
	exp.SetExpression(expr)
	if err = exp.CompileExpression(); err != nil {
		err = fmt.Errorf(`%v: "%s"`, err, expr)
		return
	}
	output = exp.GetEvaluatedValue()
	return
}