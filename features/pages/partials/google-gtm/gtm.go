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
	"html/template"
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
)

var (
	DefaultGtmDomain   = "www.googletagmanager.com"
	DefaultGtmNonceTag = "google-tag-manager"
)

//go:embed gtm-head-tail.tmpl
var HeadTailTmpl string

//go:embed gtm-body-head.tmpl
var BodyHeadTmpl string

const Tag feature.Tag = "google-tag-manager"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	feature.FuncMapProvider
	feature.RequestRewriter
	feature.PageContextModifier
	feature.ContentSecurityPolicyModifier
}

type MakeFeature interface {
	Make() Feature
}

type CFeature struct {
	feature.CFeature

	googleGtmId string
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
}

func (f *CFeature) Make() Feature {
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
	)
	_ = b.RegisterTemplatePartial("head", "tail", "gtm-script", HeadTailTmpl)
	_ = b.RegisterTemplatePartial("head", "tail", "gtm-noscript", BodyHeadTmpl)
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
	return
}

func (f *CFeature) Shutdown() {

}

func (f *CFeature) MakeFuncMap(ctx context.Context) (fm feature.FuncMap) {
	fm = feature.FuncMap{
		"gtmNoScriptTag":   f.GtmNoScriptTagFn,
		"gtmHeadScriptTag": f.GtmHeadScriptTagFn,
	}
	return
}

func (f *CFeature) GetGoogleGtmId(ctx context.Context) (gtmCode string) {
	var ctxGtmCode string
	if ctx != nil {
		ctxGtmCode = ctx.String("GoogleTagManagerId", "")
	}
	if gtmCode != "" {
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
	gtmCode := f.GetGoogleGtmId(pageCtx)
	if gtmCode != "" {
		gtmNonce, _ := f.Enjin.ContentSecurityPolicy().GetRequestNonce(DefaultGtmNonceTag, r)
		themeOut.SetSpecific("GoogleTagManagerContainerId", gtmCode)
		themeOut.SetSpecific("GoogleTagManagerScriptNonce", gtmNonce)
	}
	return
}

func (f *CFeature) ModifyContentSecurityPolicy(policy csp.Policy, r *http.Request) (modified csp.Policy) {
	gtmNonce, _ := f.Enjin.ContentSecurityPolicy().GetRequestNonce(DefaultGtmNonceTag, r)
	modified = policy.
		Add(csp.NewImgSrc(csp.NewHostSource(DefaultGtmDomain))).
		Add(csp.NewScriptSrc(csp.NewNonceSource(gtmNonce), csp.NewHostSource(DefaultGtmDomain))).
		Add(csp.NewConnectSrc(csp.NewHostSource("https://www.google-analytics.com")))
	return
}

func (f *CFeature) GtmHeadScriptTagFn(gtmCode, gtmNonce string) (embed template.HTML) {
	if gtmCode != "" {
		embed = template.HTML(`<!-- Google Tag Manager -->
<script nonce="` + gtmNonce + `">(function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({'gtm.start':
new Date().getTime(),event:'gtm.js'});var f=d.getElementsByTagName(s)[0],
j=d.createElement(s),dl=l!='dataLayer'?'&l='+l:'';j.async=true;j.src=
'https://www.googletagmanager.com/gtm.js?id='+i+dl;f.parentNode.insertBefore(j,f);
})(window,document,'script','dataLayer','` + gtmCode + `');</script>
<!-- End Google Tag Manager -->`)
		return
	}
	return
}

func (f *CFeature) GtmNoScriptTagFn(gtmCode string) (embed template.HTML) {
	embed = template.HTML(`<!-- Google Tag Manager (noscript) -->
<noscript><iframe src="https://www.googletagmanager.com/ns.html?id=` + gtmCode + `"
height="0" width="0" style="display:none;visibility:hidden"></iframe></noscript>
<!-- End Google Tag Manager (noscript) -->`)
	return
}
