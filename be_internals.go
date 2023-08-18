package be

import (
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/net/headers"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
	"github.com/go-enjin/be/pkg/net/headers/policy/permissions"
	"github.com/go-enjin/be/pkg/page"
)

func (e *Enjin) Features() (cache *feature.FeaturesCache) {
	cache = e.eb.features
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

func (e *Enjin) Theming() (theming map[string]feature.Theme) {
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

func (e *Enjin) ContentSecurityPolicy() (handler *csp.PolicyHandler) {
	handler = e.contentSecurityPolicy
	return
}

func (e *Enjin) PermissionsPolicy() (handler *permissions.PolicyHandler) {
	handler = e.permissionsPolicy
	return
}

func (e *Enjin) PublicFileSystems() (registry fs.Registry) {
	registry = e.eb.publicFileSystems
	return
}

func (e *Enjin) ListTemplatePartials(block, position string) (names []string) {
	found := make(map[string]struct{})
	for _, tpp := range e.eb.fTemplatePartialsProvider {
		for _, name := range tpp.ListTemplatePartials(block, position) {
			if _, present := found[name]; present {
				continue
			}
			names = append(names, name)
			found[name] = struct{}{}
		}
	}
	return
}

func (e *Enjin) GetTemplatePartial(block, position, name string) (tmpl string, ok bool) {
	for _, tpp := range e.eb.fTemplatePartialsProvider {
		if tmpl, ok = tpp.GetTemplatePartial(block, position, name); ok {
			return
		}
	}
	return
}

func (e *Enjin) GetThemeRenderer(ctx context.Context) (renderer feature.ThemeRenderer) {

	if namedRenderer := ctx.String("ThemeRenderer", ""); namedRenderer != "" {
		for _, tr := range e.eb.fThemeRenderers {
			if tr.Tag().Equal(feature.Tag(namedRenderer)) {
				renderer = tr
				break
			}
		}
	}
	if renderer == nil {
		renderer = e.eb.fThemeRenderers[0]
	}

	return
}