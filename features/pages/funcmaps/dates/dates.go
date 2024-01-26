//go:build page_funcmaps || pages || all

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

package dates

import (
	"fmt"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/maths"
)

var _ Feature = (*CFeature)(nil)

const Tag feature.Tag = "pages-funcmaps-dates"

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
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.CFeature.Construct(f)
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
		"cmpDateFmt":  CompareDateFormats,
		"fmtDate":     DateFormat,
		"fmtTime":     TimeFormat,
		"fmtDateTime": DateTimeFormat,
		"now":         time.Now,
		"todayAt":     TodayAt,
	}
	return
}

func CompareDateFormats(format string, a, b time.Time) (same bool) {
	same = a.Format(format) == b.Format(format)
	return
}

func DateFormat(input interface{}) (formatted string, err error) {
	switch v := input.(type) {
	case string:
		var t time.Time
		if t, err = beContext.ParseTimeStructure(v); err == nil {
			formatted = t.Format(beContext.DateLayout)
		}
	case time.Time:
		formatted = v.Format(beContext.DateLayout)
	default:
		err = fmt.Errorf("unsupported input type: %T", input)
	}
	return
}

func TimeFormat(input interface{}) (formatted string, err error) {
	switch v := input.(type) {
	case string:
		var t time.Time
		if t, err = beContext.ParseTimeStructure(v); err == nil {
			formatted = t.Format(beContext.TimeLayout)
		}
	case time.Time:
		formatted = v.Format(beContext.TimeLayout)
	default:
		err = fmt.Errorf("unsupported input type: %T", input)
	}
	return
}

func DateTimeFormat(input interface{}) (formatted string, err error) {
	switch v := input.(type) {
	case string:
		var t time.Time
		if t, err = beContext.ParseTimeStructure(v); err == nil {
			formatted = t.Format(beContext.DateTimeLayout)
		}
	case time.Time:
		formatted = v.Format(beContext.DateTimeLayout)
	default:
		err = fmt.Errorf("unsupported input type: %T", input)
	}
	return
}

func TodayAt(hour, minute interface{}) (today time.Time, err error) {
	parse := func(input interface{}) (value int) {
		switch t := input.(type) {
		case string:
			value, _ = strconv.Atoi(t)
		case int:
			value = t
		case float64:
			value = int(t)
		}
		return
	}
	var h, m int
	if h = maths.Clamp(parse(hour), 0, 24); h == 24 {
		h = 0
	}
	if m = maths.Clamp(parse(minute), 0, 60); m == 60 {
		m = 0
	}
	now := time.Now()
	today = time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location())
	return
}
