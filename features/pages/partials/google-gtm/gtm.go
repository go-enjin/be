//go:build page_partials_google_gtm || google_gtm || page_partials_google || google || pages_partials || pages || all

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

package gtm

import (
	_ "embed"
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
)

var (
	DefaultGaDomain    = "www.google-analytics.com"
	DefaultGtmDomain   = "www.googletagmanager.com"
	DefaultGtmNonceTag = "google-tag-manager"
)

//go:embed gtm-head-head.tmpl
var HeadHeadTmpl string

//go:embed gtm-body-head.tmpl
var BodyHeadTmpl string

const Tag feature.Tag = "google-tag-manager"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	feature.RequestRewriter
	feature.PageContextModifier
	feature.ContentSecurityPolicyModifier
}

type MakeFeature interface {
	SetGtmId(id string) MakeFeature
	SetUseGA4(enabled bool) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	googleGtmId string
	googleIsGA4 bool
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
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) SetGtmId(id string) MakeFeature {
	f.googleGtmId = id
	return f
}

func (f *CFeature) SetUseGA4(enabled bool) MakeFeature {
	f.googleIsGA4 = enabled
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	b.AddFlags(
		&cli.StringFlag{
			Name:     "google-gtm-id",
			Usage:    "specify the GTM ID (overrides theme)",
			EnvVars:  b.MakeEnvKeys("GOOGLE_GTM_ID"),
			Category: f.Tag().String(),
		},
		&cli.BoolFlag{
			Name:     "google-gtm-use-gtag-js",
			Usage:    "use the gtag.js script instead of the default script+noscript",
			EnvVars:  b.MakeEnvKeys("GOOGLE_GTM_USE_GTAG_JS"),
			Category: f.Tag().String(),
		},
	)
	_ = b.RegisterTemplatePartial("head", "head", "gtm-head-script", HeadHeadTmpl)
	_ = b.RegisterTemplatePartial("body", "head", "gtm-body-script", BodyHeadTmpl)
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	if v := ctx.String("google-gtm-id"); v != "" {
		f.googleGtmId = v
		log.DebugF("using google-gtm-id: %v", f.googleGtmId)
	}
	if ctx.IsSet("google-gtm-use-gtag-js") {
		f.googleIsGA4 = ctx.Bool("google-gtm-use-gtag-js")
		log.DebugF("using google-gtm-use-gtag-js: %v", f.googleIsGA4)
	}
	return
}

func (f *CFeature) Shutdown() {

}

func (f *CFeature) GetGoogleGtmId(ctx context.Context) (gtmCode string) {
	var ctxGtmCode string
	if ctx != nil {
		ctxGtmCode = ctx.String("GoogleTagManagerId", "")
	}
	if ctxGtmCode != "" {
		// front-matter override
		gtmCode = ctxGtmCode
	} else if f.googleGtmId != "" {
		// enjin cli env override
		gtmCode = f.googleGtmId
	} else if v := f.Enjin.MustGetTheme().GetConfig().Context.String(".GoogleAnalytics.GTM", ""); v != "" {
		// enjin theme setting
		gtmCode = v
	}
	return
}

func (f *CFeature) RewriteRequest(w http.ResponseWriter, r *http.Request) (modified *http.Request) {
	_, modified = f.Enjin.ContentSecurityPolicy().GetRequestNonce(DefaultGtmNonceTag, r)
	return
}

func (f *CFeature) FilterPageContext(themeCtx, pageCtx context.Context, r *http.Request) (themeOut context.Context) {
	themeOut = themeCtx
	if gtmCode := f.GetGoogleGtmId(pageCtx); gtmCode != "" {
		gtmNonce, _ := f.Enjin.ContentSecurityPolicy().GetRequestNonce(DefaultGtmNonceTag, r)
		themeOut.SetSpecific("GoogleTagManagerContainerId", gtmCode)
		themeOut.SetSpecific("GoogleTagManagerScriptNonce", gtmNonce)
		themeOut.SetSpecific("GoogleTagManagerUseGtagJs", f.googleIsGA4)
	}
	return
}

func (f *CFeature) ModifyContentSecurityPolicy(policy csp.Policy, r *http.Request) (modified csp.Policy) {
	gtmNonce, _ := f.Enjin.ContentSecurityPolicy().GetRequestNonce(DefaultGtmNonceTag, r)
	modified = policy.
		Add(csp.NewImgSrc(csp.NewHostSource(DefaultGtmDomain))).
		Add(csp.NewFrameSrc(csp.NewHostSource("https://" + DefaultGtmDomain))).
		Add(csp.NewScriptSrc(csp.NewNonceSource(gtmNonce), csp.NewHostSource(DefaultGtmDomain))).
		Add(csp.NewConnectSrc(csp.NewHostSource("https://" + DefaultGaDomain)))
	return
}
