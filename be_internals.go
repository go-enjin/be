package be

import (
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/net/headers"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/theme"
)

func (e *Enjin) Features() (features []feature.Feature) {
	for _, feat := range e.eb.features {
		features = append(features, feat)
	}
	return
}

func (e *Enjin) Pages() (pages map[string]*page.Page) {
	pages = e.eb.pages
	return
}

func (e *Enjin) Theme() (theme string) {
	theme = e.eb.theme
	return
}

func (e *Enjin) Theming() (theming map[string]*theme.Theme) {
	theming = e.eb.theming
	return
}

func (e *Enjin) Headers() (headers []headers.ModifyHeadersFn) {
	headers = e.eb.headers
	return
}

func (e *Enjin) Domains() (domains []string) {
	domains = e.eb.domains
	return
}

func (e *Enjin) Consoles() (consoles map[feature.Tag]feature.Console) {
	consoles = e.eb.consoles
	return
}

func (e *Enjin) Processors() (processors map[string]feature.ReqProcessFn) {
	processors = e.eb.processors
	return
}

func (e *Enjin) Translators() (translators map[string]feature.TranslateOutputFn) {
	translators = e.eb.translators
	return
}

func (e *Enjin) Transformers() (transformers map[string]feature.TransformOutputFn) {
	transformers = e.eb.transformers
	return
}

func (e *Enjin) Slugsums() (enabled bool) {
	enabled = e.eb.slugsums
	return
}