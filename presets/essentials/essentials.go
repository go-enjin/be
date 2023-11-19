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

package essentials

import (
	"github.com/go-enjin/be/features/outputs/htmlify"
	"github.com/go-enjin/be/features/pages/formats"
	"github.com/go-enjin/be/features/pages/formats/html"
	"github.com/go-enjin/be/features/pages/formats/tmpl"
	"github.com/go-enjin/be/features/pages/funcmaps"
	"github.com/go-enjin/be/features/pages/partials"
	"github.com/go-enjin/be/features/pages/status"
	"github.com/go-enjin/be/features/requests/deny"
	"github.com/go-enjin/be/features/requests/headers/proxy"
	modifiers "github.com/go-enjin/be/features/requests/pages/context-modifiers"
	"github.com/go-enjin/be/features/requests/pages/i18n"
	"github.com/go-enjin/be/features/requests/pages/policies"
	"github.com/go-enjin/be/features/requests/pages/request"
	"github.com/go-enjin/be/features/requests/pages/restrictions"
	"github.com/go-enjin/be/features/requests/pages/sitemenus"
	"github.com/go-enjin/be/features/srv/factories/spinlockers"
	"github.com/go-enjin/be/features/srv/listeners/httpd"
	beLogHandler "github.com/go-enjin/be/features/srv/logging/handler"
	beLogger "github.com/go-enjin/be/features/srv/logging/logger"
	"github.com/go-enjin/be/features/srv/middleware/locales"
	"github.com/go-enjin/be/features/srv/middleware/panics"
	"github.com/go-enjin/be/features/srv/pages"
	"github.com/go-enjin/be/features/srv/theme/renderer"
	"github.com/go-enjin/be/pkg/feature"
)

var (
	_ Preset     = (*CPreset[MakePreset])(nil)
	_ MakePreset = (*CPreset[MakePreset])(nil)
)

const Name = "preset-essentials"

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
}

type CPreset[MakeTypedPreset interface{}] struct {
	feature.CPreset[MakeTypedPreset]

	formats  []feature.PageFormat
	funcmaps []feature.FuncMapProvider
	renderer feature.ThemeRenderer
	listener feature.ServiceListener
}

func New() MakePreset {
	p := new(CPreset[MakePreset])
	p.Name = Name
	p.Features = feature.Features{
		panics.New().Make(),
		spinlockers.New().Make(),
		locales.New().Make(),
		deny.New().Defaults().Make(),
		proxy.New().Enable().Make(),
		status.New().Make(),
		partials.New().Make(),
		request.New().Make(),
		i18n.New().Make(),
		policies.New().Make(),
		restrictions.New().Make(),
		sitemenus.New().Make(),
		modifiers.New().Make(),
		pages.New().Make(),
		htmlify.New().Make(),
		beLogHandler.New().Make(),
		beLogger.New().SetCombined(true).Make(),
	}
	p.Init(p)
	return p
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
		AddFormat(tmpl.New().Make()).
		AddFormat(html.New().Make()).
		AddFormat(p.formats...).
		Make())

	if err = p.CPreset.Preset(b); err != nil {
		return
	}

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