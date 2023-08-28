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

package defaults

import (
	"github.com/go-enjin/be/features/log/papertrail"
	"github.com/go-enjin/be/features/outputs/htmlify"
	"github.com/go-enjin/be/features/pages/formats"
	"github.com/go-enjin/be/features/pages/funcmaps"
	"github.com/go-enjin/be/features/pages/partials"
	"github.com/go-enjin/be/features/pages/permalink"
	"github.com/go-enjin/be/features/pages/query"
	"github.com/go-enjin/be/features/requests/deny"
	"github.com/go-enjin/be/features/requests/headers/proxy"
	"github.com/go-enjin/be/features/srv/listeners/httpd"
	beLogHandler "github.com/go-enjin/be/features/srv/logging/handler"
	beLogger "github.com/go-enjin/be/features/srv/logging/logger"
	"github.com/go-enjin/be/features/srv/pages"
	"github.com/go-enjin/be/features/srv/theme/renderer"
	"github.com/go-enjin/be/features/user/auth/basic"
	"github.com/go-enjin/be/features/user/base/htenv"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

var (
	_ Preset     = (*CPreset[MakePreset])(nil)
	_ MakePreset = (*CPreset[MakePreset])(nil)
)

const Name = "preset-defaults"

type Preset interface {
	feature.Preset
}

type MakePreset interface {
	feature.BaseMakePreset[MakePreset]

	Make() Preset

	SetRenderer(r feature.ThemeRenderer) MakePreset
	SetListener(l feature.ServiceListener) MakePreset

	AddFormats(formats ...feature.PageFormat) MakePreset
	AddFuncmaps(funcmaps ...feature.FuncMapProvider) MakePreset

	SetBasicAuthTag(tag feature.Tag) MakePreset

	SetHtenvTag(tag feature.Tag) MakePreset
	SetHtenvIgnored(patterns ...string) MakePreset
	AddHtenvIgnored(patterns ...string) MakePreset
}

type CPreset[MakeTypedPreset interface{}] struct {
	feature.CPreset[MakeTypedPreset]

	htenvTag     feature.Tag
	htenvIgnored []string
	basicAuthTag feature.Tag

	formats  []feature.PageFormat
	funcmaps []feature.FuncMapProvider
	renderer feature.ThemeRenderer
	listener feature.ServiceListener
}

func New() MakePreset {
	p := new(CPreset[MakePreset])
	p.Name = Name
	p.Features = feature.Features{
		deny.New().Defaults().Make(),
		proxy.New().Enable().Make(),
		partials.New().Make(),
		permalink.New().Make(),
		query.New().Make(),
		htmlify.New().Make(),
		papertrail.New().Make(),
		pages.New().Make(),
		beLogHandler.New().Make(),
		beLogger.New().SetCombined(true).Make(),
	}
	p.htenvTag = "htenv"
	p.htenvIgnored = []string{`^/favicon.ico$`}
	p.basicAuthTag = "basic-auth"
	p.Init(p)
	return p
}

func (p *CPreset[MakeTypedPreset]) SetBasicAuthTag(tag feature.Tag) MakeTypedPreset {
	if tag.IsNil() {
		log.FatalDF(1, "basic auth feature tag must not be empty")
	}
	p.basicAuthTag = tag
	return interface{}(p).(MakeTypedPreset)
}

func (p *CPreset[MakeTypedPreset]) SetHtenvTag(tag feature.Tag) MakeTypedPreset {
	if tag.IsNil() {
		log.FatalDF(1, "htenv feature tag must not be empty")
	}
	p.htenvTag = tag
	return interface{}(p).(MakeTypedPreset)
}

func (p *CPreset[MakeTypedPreset]) SetHtenvIgnored(patterns ...string) MakeTypedPreset {
	p.htenvIgnored = patterns
	return interface{}(p).(MakeTypedPreset)
}

func (p *CPreset[MakeTypedPreset]) AddHtenvIgnored(patterns ...string) MakeTypedPreset {
	p.htenvIgnored = append(p.htenvIgnored, patterns...)
	return interface{}(p).(MakeTypedPreset)
}

func (p *CPreset[MakeTypedPreset]) SetRenderer(r feature.ThemeRenderer) MakeTypedPreset {
	p.renderer = r
	return interface{}(p).(MakeTypedPreset)
}

func (p *CPreset[MakeTypedPreset]) SetListener(l feature.ServiceListener) MakeTypedPreset {
	p.listener = l
	return interface{}(p).(MakeTypedPreset)
}

func (p *CPreset[MakeTypedPreset]) AddFormats(formats ...feature.PageFormat) MakeTypedPreset {
	p.formats = append(p.formats, formats...)
	return interface{}(p).(MakeTypedPreset)
}

func (p *CPreset[MakeTypedPreset]) AddFuncmaps(funcmaps ...feature.FuncMapProvider) MakeTypedPreset {
	p.funcmaps = append(p.funcmaps, funcmaps...)
	return interface{}(p).(MakeTypedPreset)
}

func (p *CPreset[MakeTypedPreset]) Make() (feat Preset) {
	return p
}

func (p *CPreset[MakeTypedPreset]) Preset(b feature.Builder) (err error) {

	p.IncludeFeature(b, funcmaps.New().
		Defaults().
		Include(p.funcmaps...).
		Make())

	p.IncludeFeature(b, formats.New().
		Defaults().
		AddFormat(p.formats...).
		Make())

	if err = p.CPreset.Preset(b); err != nil {
		return
	}

	p.IncludeFeature(b, htenv.NewTagged(p.htenvTag).Make())
	p.IncludeFeature(b, basic.NewTagged(p.basicAuthTag).
		AddUserbase(p.htenvTag.String(), p.htenvTag.String(), p.htenvTag.String()).
		Ignore(p.htenvIgnored...).
		Make())

	if p.renderer != nil {
		p.IncludeFeature(b, p.renderer)
	} else {
		p.IncludeFeature(b, renderer.New().Make())
	}

	if p.listener != nil {
		p.IncludeFeature(b, p.listener)
	} else {
		p.IncludeFeature(b, httpd.New().Make())
	}

	return
}