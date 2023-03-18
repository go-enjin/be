//go:build page_external_google_gtm || page_external_google || pages || all

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
	"html/template"
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
	"github.com/go-enjin/be/pkg/theme"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const (
	Tag feature.Tag = "PagesExternalGoogleGTM"

	DefaultGtmDomain   = "www.googletagmanager.com"
	DefaultGtmNonceTag = "google-tag-manager"
)

type Feature interface {
	feature.Feature
	feature.RequestRewriter
	feature.PageContextModifier
	feature.ContentSecurityPolicyModifier
}

type CFeature struct {
	feature.CFeature

	googleGtmId string

	cli   *cli.Context
	enjin feature.Internals
	theme *theme.Theme
}

type MakeFeature interface {
	Make() Feature
}

func New() MakeFeature {
	f := new(CFeature)
	f.Init(f)
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
}

func (f *CFeature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	b.AddFlags(
		&cli.StringFlag{
			Name:    "google-gtm-id",
			Usage:   "specify the GTM ID (overrides theme)",
			EnvVars: b.MakeEnvKeys("GOOGLE_GTM_ID"),
		},
	)
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.enjin = enjin
	var err error
	if f.theme, err = f.enjin.GetTheme(); err != nil {
		log.FatalF("error getting enjin theme: %v - %v", f.enjin.SiteName())
	}

	theme.RegisterFuncMap("gtmHeadScriptTag", f.GtmHeadScriptTagFn)
	theme.RegisterFuncMap("gtmNoScriptTag", f.GtmNoScriptTagFn)
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	f.cli = ctx
	if v := f.cli.String("google-gtm-id"); v != "" {
		f.googleGtmId = v
		log.DebugF("using google-gtm-id: %v", f.googleGtmId)
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
	} else if f.theme.Config.GoogleAnalytics.GTM != "" {
		// enjin theme setting
		gtmCode = f.theme.Config.GoogleAnalytics.GTM
	}
	return
}

func (f *CFeature) RewriteRequest(w http.ResponseWriter, r *http.Request) (modified *http.Request) {
	_, modified = f.enjin.ContentSecurityPolicy().GetRequestNonce(DefaultGtmNonceTag, r)
	return
}

func (f *CFeature) FilterPageContext(themeCtx, pageCtx context.Context, r *http.Request) (themeOut context.Context) {
	themeOut = themeCtx
	gtmCode := f.GetGoogleGtmId(pageCtx)
	if gtmCode != "" {
		gtmNonce, _ := f.enjin.ContentSecurityPolicy().GetRequestNonce(DefaultGtmNonceTag, r)
		themeOut.SetSpecific("GoogleTagManagerContainerId", gtmCode)
		themeOut.SetSpecific("GoogleTagManagerScriptNonce", gtmNonce)
	}
	return
}

func (f *CFeature) ModifyContentSecurityPolicy(policy csp.Policy, r *http.Request) (modified csp.Policy) {
	gtmNonce, _ := f.enjin.ContentSecurityPolicy().GetRequestNonce(DefaultGtmNonceTag, r)
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