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

package auth

import (
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/golang-org-x-text/message"

	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	site_environ "github.com/go-enjin/be/pkg/feature/site-environ"
	site_including "github.com/go-enjin/be/pkg/feature/site-including"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/menu"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/pkg/slices"
	"github.com/go-enjin/be/pkg/userbase"
	"github.com/go-enjin/be/types/site"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "site-auth"

type Feature interface {
	feature.SiteFeature
	feature.SiteAuthFeature
	feature.SiteAuthRequestHandler
}

type MakeFeature interface {
	SetAllBackupsRequired(required bool) MakeFeature
	SetRequiredAccountBackups(tags ...feature.Tag) MakeFeature
	SetRequiredMultiFactorBackups(tags ...feature.Tag) MakeFeature
	SetNumRequiredFactors(count int) MakeFeature

	SetUserSignups(allowed bool) MakeFeature
	DenySignupsFrom(emails ...string) MakeFeature
	AllowSignupsFrom(emails ...string) MakeFeature

	SetSecretKey(aud string, key []byte) MakeFeature
	SetXsrfHeaderName(headerName string) MakeFeature
	SetXsrfCookieName(cookieName string) MakeFeature
	SetJwtCookieName(cookieName string) MakeFeature

	SetSecureCookies(secure bool) MakeFeature
	SetSameSiteCookies(sameSite http.SameSite) MakeFeature

	SetSessionDuration(duration time.Duration) MakeFeature
	SetVerifiedDuration(duration time.Duration) MakeFeature

	SetRoutePaths(signIn, signOut, challenge string) MakeFeature

	IncludeProviders(features ...feature.Feature) MakeFeature
	IncludingProviders(tags ...feature.Tag) MakeFeature

	IncludeMultiFactors(features ...feature.Feature) MakeFeature
	IncludingMultiFactors(tags ...feature.Tag) MakeFeature

	Make() Feature
}

type CFeature struct {
	site.CSiteFeature[MakeFeature]

	env *site_environ.CSiteEnviron[MakeFeature]

	allBackupsRequired bool
	mustBackupAccounts feature.Tags
	mustBackupFactors  feature.Tags

	allowSignups  bool
	deniedEmails  []string
	allowedEmails []string

	secretKeys   map[string][]byte
	audienceKeys map[string][]byte

	secureCookies   bool
	sameSiteCookies http.SameSite

	jwtCookieName  string
	xsrfCookieName string
	xsrfHeaderName string

	numFactorsRequired int

	sessionDuration  time.Duration
	verifiedDuration time.Duration

	signInPath    string
	signOutPath   string
	challengePath string

	sap *site_including.CSiteIncluding[feature.SiteAuthProvider, MakeFeature]
	sab *site_including.CSiteIncluding[feature.SiteUserSetupStage, MakeFeature]
	mfa *site_including.CSiteIncluding[feature.SiteMultiFactorProvider, MakeFeature]
	mfb *site_including.CSiteIncluding[feature.SiteMultiFactorProvider, MakeFeature]
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.SetSiteFeatureKey("site-auth")
	f.SetSiteFeatureIcon("fa-solid fa-shield")
	f.SetSiteFeatureLabel(func(printer *message.Printer) (label string) {
		label = printer.Sprintf("Authentication")
		return
	})
	f.CSiteFeature.Construct(f)
	return f
}

func (f *CFeature) UsageNotes() (notes []string) {
	notes = f.env.SiteEnvironUsageNotes()
	return
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.env = site_environ.New[MakeFeature](this,
		"secret-keys", "64 or more random characters",
		"audience-keys", "32 or more random characters",
	)
	f.sap = site_including.New[feature.SiteAuthProvider, MakeFeature](this)
	f.sab = site_including.New[feature.SiteUserSetupStage, MakeFeature](this)
	f.mfa = site_including.New[feature.SiteMultiFactorProvider, MakeFeature](this)
	f.mfb = site_including.New[feature.SiteMultiFactorProvider, MakeFeature](this)
	f.signInPath = DefaultSignInPath
	f.signOutPath = DefaultSignOutPath
	f.challengePath = DefaultChallengePath
	f.jwtCookieName = DefaultJwtCookieName
	f.xsrfCookieName = DefaultXsrfCookieName
	f.xsrfHeaderName = DefaultXsrfHeaderName
	f.numFactorsRequired = DefaultRequiredFactors
	f.sessionDuration = DefaultSessionDuration
	f.verifiedDuration = DefaultVerifiedDuration
	f.secretKeys = make(map[string][]byte)
	f.audienceKeys = make(map[string][]byte)
	f.secureCookies = false
	f.sameSiteCookies = http.SameSiteStrictMode
	return
}

func (f *CFeature) SetAllBackupsRequired(required bool) MakeFeature {
	f.allBackupsRequired = required
	return f
}

func (f *CFeature) SetRequiredAccountBackups(tags ...feature.Tag) MakeFeature {
	if f.allBackupsRequired {
		f.allBackupsRequired = false
	}
	f.mustBackupAccounts = tags
	return f
}

func (f *CFeature) SetRequiredMultiFactorBackups(tags ...feature.Tag) MakeFeature {
	if f.allBackupsRequired {
		f.allBackupsRequired = false
	}
	f.mustBackupFactors = tags
	return f
}

func (f *CFeature) SetNumRequiredFactors(count int) MakeFeature {
	f.numFactorsRequired = count
	return f
}

func (f *CFeature) SetUserSignups(allowed bool) MakeFeature {
	f.allowSignups = allowed
	return f
}

func (f *CFeature) DenySignupsFrom(emails ...string) MakeFeature {
	if list, err := mail.ParseAddressList(strings.Join(emails, ", ")); err != nil {
		log.FatalDF(1, "error parsing address list: %v", err)
	} else {
		for _, address := range list {
			f.deniedEmails = slices.Append(f.deniedEmails, address.Address)
		}
	}
	return f
}

func (f *CFeature) AllowSignupsFrom(emails ...string) MakeFeature {
	if list, err := mail.ParseAddressList(strings.Join(emails, ",")); err != nil {
		log.FatalDF(1, "error parsing address list: %v", err)
	} else {
		for _, address := range list {
			f.allowedEmails = slices.Append(f.allowedEmails, strings.ToLower(address.Address))
		}
	}
	return f
}

func (f *CFeature) SetSecretKey(aud string, value []byte) MakeFeature {
	if aud == "" {
		aud = DefaultAudience
	} else if aud != DefaultAudience {
		aud = strcase.ToKebab(aud)
	}
	f.audienceKeys[aud] = value
	return f
}

func (f *CFeature) SetXsrfHeaderName(headerName string) MakeFeature {
	f.xsrfHeaderName = headerName
	return f
}

func (f *CFeature) SetXsrfCookieName(cookieName string) MakeFeature {
	f.xsrfCookieName = cookieName
	return f
}

func (f *CFeature) SetJwtCookieName(cookieName string) MakeFeature {
	f.jwtCookieName = cookieName
	return f
}

func (f *CFeature) SetSecureCookies(secure bool) MakeFeature {
	f.secureCookies = secure
	return f
}

func (f *CFeature) SetSameSiteCookies(sameSite http.SameSite) MakeFeature {
	f.sameSiteCookies = sameSite
	return f
}

func (f *CFeature) SetSessionDuration(duration time.Duration) MakeFeature {
	f.sessionDuration = duration
	return f
}

func (f *CFeature) SetVerifiedDuration(duration time.Duration) MakeFeature {
	f.verifiedDuration = duration
	return f
}

func (f *CFeature) SetRoutePaths(signIn, signOut, challenge string) MakeFeature {
	if signIn != "" {
		f.signInPath = bePath.CleanWithSlash(signIn)
	}
	if signOut != "" {
		f.signOutPath = bePath.CleanWithSlash(signOut)
	}
	if challenge != "" {
		f.challengePath = bePath.CleanWithSlash(challenge)
	}
	return f
}

func (f *CFeature) IncludeProviders(features ...feature.Feature) MakeFeature {
	f.sap.Include(features...)
	return f
}

func (f *CFeature) IncludingProviders(tags ...feature.Tag) MakeFeature {
	f.sap.Including(tags...)
	return f
}

func (f *CFeature) IncludeMultiFactors(features ...feature.Feature) MakeFeature {
	f.mfa.Include(features...)
	return f
}

func (f *CFeature) IncludingMultiFactors(tags ...feature.Tag) MakeFeature {
	f.mfa.Including(tags...)
	return f
}

func (f *CFeature) Make() (feat Feature) {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CSiteFeature.Build(b); err != nil {
		return
	}
	f.sap.BuildSiteIncluding(b)
	f.sab.BuildSiteIncluding(b)
	f.mfa.BuildSiteIncluding(b)
	f.mfb.BuildSiteIncluding(b)

	if f.signInPath != "" {
		f.signInPath = "/" + bePath.TrimSlashes(f.signInPath)
	}
	if f.signOutPath != "" {
		f.signOutPath = "/" + bePath.TrimSlashes(f.signOutPath)
	}

	category := f.KebabTag

	fns := f.makeFlagNames()

	b.AddFlags(
		&cli.StringFlag{
			Name:     fns.secretKey,
			Usage:    "specify the default audience secret key",
			EnvVars:  b.MakeEnvKeys(fns.secretKey),
			Category: category,
		},
		&cli.Int64Flag{
			Name:     fns.sessionDuration,
			Usage:    "specify the session duration",
			Value:    int64(f.sessionDuration.Seconds()),
			EnvVars:  b.MakeEnvKeys(fns.sessionDuration),
			Category: category,
		},
		&cli.StringFlag{
			Name:     fns.xsrfHeaderName,
			Usage:    "specify the name of the XSRF header key (omit to disable XSRF)",
			EnvVars:  b.MakeEnvKeys(fns.xsrfHeaderName),
			Category: category,
		},
		&cli.StringFlag{
			Name:     fns.xsrfCookieName,
			Usage:    "specify the name of the XSRF cookie",
			EnvVars:  b.MakeEnvKeys(fns.xsrfCookieName),
			Category: category,
		},
		&cli.StringFlag{
			Name:     fns.jwtCookieName,
			Usage:    "specify the name of the JWT cookie",
			EnvVars:  b.MakeEnvKeys(fns.jwtCookieName),
			Category: category,
		},
		&cli.BoolFlag{
			Name:     fns.secureCookies,
			Usage:    "set the secure flag on all cookies",
			EnvVars:  b.MakeEnvKeys(fns.secureCookies),
			Category: category,
		},
		&cli.BoolFlag{
			Name:     fns.allowSignups,
			Usage:    "enable public signups",
			EnvVars:  b.MakeEnvKeys(fns.allowSignups),
			Category: category,
		},
		&cli.StringSliceFlag{
			Name:     fns.allowEmails,
			Usage:    "specify emails allowed to signup when public signups are disabled",
			EnvVars:  b.MakeEnvKeys(fns.allowEmails),
			Category: category,
		},
		&cli.StringSliceFlag{
			Name:     fns.denyEmails,
			Usage:    "specify emails denied from the site",
			EnvVars:  b.MakeEnvKeys(fns.denyEmails),
			Category: category,
		},
		&cli.StringFlag{
			Name:     fns.signInPath,
			Usage:    "specify the sign-in sub-path",
			EnvVars:  b.MakeEnvKeys(fns.signInPath),
			Category: category,
			Value:    f.signInPath,
		},
		&cli.StringFlag{
			Name:     fns.signOutPath,
			Usage:    "specify the sign-out sub-path",
			EnvVars:  b.MakeEnvKeys(fns.signOutPath),
			Category: category,
			Value:    f.signOutPath,
		},
	)
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CSiteFeature.Startup(ctx); err != nil {
		return
	}
	if err = f.env.StartupSiteEnviron(); err != nil {
		return
	} else if err = f.loadEnvironment(); err != nil {
		return
	}
	f.mfa.StartupSiteIncluding(f.Enjin)
	f.sap.StartupSiteIncluding(f.Enjin)
	for _, sap := range f.sap.Features {
		if sap.IsBackupProvider() {
			f.sab.Include(sap)
		}
	}
	f.sab.StartupSiteIncluding(f.Enjin)
	for _, mfp := range f.mfa.Features {
		if mfp.IsMultiFactorBackup() {
			f.mfb.Include(mfp)
		}
	}
	f.mfb.StartupSiteIncluding(f.Enjin)

	fns := f.makeFlagNames()

	if ctx.IsSet(fns.allowSignups) {
		f.SetUserSignups(ctx.Bool(fns.allowSignups))
	}
	if ctx.IsSet(fns.allowEmails) {
		if emails := ctx.StringSlice(fns.allowEmails); len(emails) > 0 {
			f.AllowSignupsFrom(emails...)
		}
	}
	if ctx.IsSet(fns.denyEmails) {
		if emails := ctx.StringSlice(fns.denyEmails); len(emails) > 0 {
			f.DenySignupsFrom(emails...)
		}
	}

	if ctx.IsSet(fns.secretKey) {
		if v := ctx.String(fns.secretKey); v != "" {
			f.audienceKeys[DefaultAudience] = []byte(v)
		} else {
			err = fmt.Errorf("--%s value cannot be empty", fns.secretKey)
			return
		}
	} else {
		err = fmt.Errorf("--%s is required", fns.secretKey)
		return
	}

	for _, aud := range maps.SortedKeys(f.audienceKeys) {
		if count := len(f.audienceKeys[aud]); count < 32 {
			err = fmt.Errorf("%s audience secret key needs to be at least 32 bytes long", aud)
			return
		}
	}

	if ctx.IsSet(fns.sessionDuration) {
		if v := ctx.Int64(fns.sessionDuration); v > 0 {
			f.sessionDuration = time.Second * time.Duration(v)
		}
	}
	if f.sessionDuration.Seconds() < MinSessionDuration.Seconds() {
		err = fmt.Errorf("session duration is less than: %v", MinSessionDuration.String())
		return
	}

	if ctx.IsSet(fns.xsrfHeaderName) {
		f.xsrfHeaderName = http.CanonicalHeaderKey(ctx.String(fns.xsrfHeaderName))
	}

	if ctx.IsSet(fns.xsrfCookieName) {
		f.xsrfCookieName = strcase.ToSnake(ctx.String(ctx.String(fns.xsrfCookieName)))
	}

	if ctx.IsSet(fns.jwtCookieName) {
		f.jwtCookieName = ctx.String(fns.jwtCookieName)
	}

	if ctx.IsSet(fns.secureCookies) {
		f.secureCookies = ctx.Bool(fns.secureCookies)
	}

	if ctx.IsSet(fns.signInPath) {
		f.signInPath = ctx.String(fns.signInPath)
	}

	if ctx.IsSet(fns.signOutPath) {
		f.signOutPath = ctx.String(fns.signOutPath)
	}

	return
}

func (f *CFeature) SetupSiteFeature(s feature.Site) (err error) {
	if err = f.CSiteFeature.SetupSiteFeature(s); err != nil {
		return
	}
	for _, sap := range f.sap.Features {
		if err = sap.SetupSiteFeature(s); err != nil {
			return
		}
	}
	for _, mfp := range f.mfa.Features {
		if err = mfp.SetupSiteFeature(s); err != nil {
			return
		}
	}
	return
}

func (f *CFeature) PostStartup(_ *cli.Context) (err error) {
	if f.Site().SiteUsers() == nil {
		err = fmt.Errorf("site users not specified, please use the .SetSiteUsers builder method on the %q feature instance", f.Site().Tag())
		return
	}
	for _, mfp := range f.mfa.Features {
		mfp.SetupSiteAuthProvider(f)
	}

	if f.numFactorsRequired > f.NumFactorsPresent() {
		err = fmt.Errorf("%d factors required, site only supports %d", f.numFactorsRequired, f.NumFactorsPresent())
		return
	}

	if f.allBackupsRequired {
		f.mustBackupAccounts = f.sab.Features.Tags()
		f.mustBackupFactors = f.mfb.Features.Tags()
	} else {
		for _, tag := range f.mustBackupAccounts {
			if !f.sab.Features.Has(tag) {
				err = fmt.Errorf("inconsistent configuration, must backup account %q is not present", tag)
				return
			}
		}
		for _, tag := range f.mustBackupFactors {
			if !f.mfb.Features.Has(tag) {
				err = fmt.Errorf("inconsistent configuration, must backup multi-factor %q is not present", tag)
				return
			}
		}
		if f.numFactorsRequired > 0 {
			if f.mfb.Features.Len() == 0 {
				err = fmt.Errorf("at least one mfa backup feature is required when there is a minium number of factors required")
				return
			}
			if f.mustBackupFactors.Len() == 0 {
				// if no specific backups are specified as mandatory, all backups are required
				f.mustBackupFactors = f.mfb.Features.Tags()
			}
		}
	}

	log.InfoF("%v feature settings: %v", f.Tag(), maps.PrettyMap(map[string]interface{}{
		"sign-in-path":     f.signInPath,
		"sign-out-path":    f.signOutPath,
		"secure-cookies":   f.secureCookies,
		"jwt-cookie-name":  f.jwtCookieName,
		"xsrf-cookie-name": f.xsrfCookieName,
		"xsrf-header-name": f.xsrfHeaderName,
		"site-users":       f.Site().SiteUsers().Tag(),
		"sap-features":     f.sap.Features.Tags(),
		"sab-features":     f.sab.Features.Tags(),
		"mfa-features":     f.mfa.Features.Tags(),
		"mfb-features":     f.mfb.Features.Tags(),
		"secret-keys":      maps.SortedKeys(f.secretKeys),
		"audience-names":   maps.SortedKeys(f.audienceKeys),
	}))
	return
}

func (f *CFeature) Shutdown() {
	f.CFeature.Shutdown()
}

func (f *CFeature) UserActions() (actions feature.Actions) {
	actions = feature.Actions{
		f.Action("access", "feature"),
		f.Action("reset-own", "multi-factors"),
		f.Action("reset-other", "multi-factors"),
	}
	return
}

func (f *CFeature) UpdateSiteRoutes(r chi.Router) {

	r.Route(f.signInPath, func(r chi.Router) {
		if f.mfa.Features.Len() > 0 {
			r.Post(f.challengePath, f.ProcessChallengeRequest)
			r.Get(f.challengePath, f.ServeChallengeRequest)
		}
		r.Post("/*", f.HandleSignInPage)
		r.Get("/*", f.ServeSignInPage)
	})

	r.Route(f.signOutPath, func(r chi.Router) {
		r.Post("/*", f.HandleSignOutPage)
		r.Get("/*", f.ServeSignOutPage)
	})
	return
}

func (f *CFeature) SiteFeatureMenu(r *http.Request) (m menu.Menu) {
	printer := lang.GetPrinterFromRequest(r)
	if userbase.IsVisitor(r) {
		m = append(m, &menu.Item{
			Text: printer.Sprintf("Sign-in"),
			Href: f.Site().SitePath() + f.signInPath,
			Icon: "fa-solid fa-right-to-bracket",
		})
		return
	}
	m = append(m, &menu.Item{
		Text: printer.Sprintf("Sign-out"),
		Href: f.Site().SitePath() + f.signOutPath,
		Icon: "fa-solid fa-right-from-bracket",
	})
	return
}

func (f *CFeature) SiteAuthSignInPath() (path string) {
	path = f.Site().SitePath() + f.signInPath
	return
}

func (f *CFeature) SiteAuthSignOutPath() (path string) {
	path = f.Site().SitePath() + f.signOutPath
	return
}

func (f *CFeature) SiteAuthChallengePath() (path string) {
	path = f.Site().SitePath() + f.signInPath + f.challengePath
	return
}

func (f *CFeature) NumFactorsPresent() (count int) {
	count = f.mfa.Features.Len() - f.mfb.Features.Len()
	return
}

func (f *CFeature) NumFactorsRequired() (count int) {
	count = f.numFactorsRequired
	return
}

func (f *CFeature) GetSignUpsAllowed() (allowed bool) {
	allowed = f.allowSignups
	return
}

func (f *CFeature) GetSessionDuration() (duration time.Duration) {
	duration = f.sessionDuration
	return
}

func (f *CFeature) GetVerifiedDuration() (duration time.Duration) {
	duration = f.verifiedDuration
	return
}

func (f *CFeature) IsUserAllowed(email string) (allowed bool) {
	su := f.Site().SiteUsers()
	rid := su.MakeRealID(email)
	eid := su.MakeEnjinID(rid)

	if slices.Within(email, f.deniedEmails) {
		// banned email address
		return
	} else if !su.UserPresent(eid) {
		// user does not exist
		if !f.allowSignups {
			// block public signups
			if len(f.allowedEmails) > 0 {
				// unless explicitly allowed
				if !slices.Within(email, f.allowedEmails) {
					return
				}
			}
		}
	}

	allowed = true
	return
}

func (f *CFeature) ResetUserFactors(r *http.Request, eid string) (err error) {
	if userbase.GetCurrentEID(r) == eid && !userbase.CurrentUserCan(r, f.Action("reset-own", "multi-factors")) {
		err = berrs.ErrPermissionDenied
		return
	} else if !userbase.CurrentUserCan(r, f.Action("reset-other", "multi-factors")) {
		err = berrs.ErrPermissionDenied
		return
	}
	for _, mfp := range f.mfa.Features {
		if err = mfp.ResetUserFactors(r, eid); err != nil {
			return
		}
	}
	return
}

type flagNames struct {
	category        string
	secretKey       string
	sessionDuration string
	xsrfHeaderName  string
	xsrfCookieName  string
	jwtCookieName   string
	secureCookies   string
	allowSignups    string
	allowEmails     string
	denyEmails      string
	signInPath      string
	signOutPath     string
}

func (f *CFeature) makeFlagNames() (fn flagNames) {
	category := f.KebabTag
	return flagNames{
		category:        category,
		secretKey:       category + "-secret-key",
		sessionDuration: category + "-session-duration",
		xsrfHeaderName:  category + "-xsrf-header-name",
		xsrfCookieName:  category + "-xsrf-cookie-name",
		jwtCookieName:   category + "-jwt-cookie-name",
		secureCookies:   category + "-secure-cookies",
		allowSignups:    category + "-allow-signups",
		allowEmails:     category + "-allow-emails",
		denyEmails:      category + "-deny-emails",
		signInPath:      category + "-sign-in-path",
		signOutPath:     category + "-sign-out-path",
	}
	return
}