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

package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/github-com-go-pkgz-auth/avatar"
	"github.com/go-enjin/github-com-go-pkgz-auth/middleware"
	"github.com/go-enjin/github-com-go-pkgz-auth/provider"
	"github.com/go-enjin/github-com-go-pkgz-auth/token"

	beKvs "github.com/go-enjin/be/pkg/kvs"
	"github.com/go-enjin/be/pkg/userbase"

	"github.com/go-enjin/github-com-go-pkgz-auth"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var (
	DefaultUrl            = "http://localhost:" + strconv.Itoa(globals.DefaultPort)
	DefaultTokenDuration  = time.Minute * 5
	DefaultCookieDuration = time.Hour * 24
	DefaultUserGroups     = []string{"user"}

	DefaultEmailNewTokenTemplate = "email-new-token"
)

const Tag feature.Tag = "user-auth-api"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	userbase.AuthUserApi
	feature.UseMiddleware
	feature.ApplyMiddleware
}

type MakeFeature interface {
	Make() Feature

	SetUrl(url string) MakeFeature
	SetIssuer(issuer string) MakeFeature

	SetPublicSignups(allowed bool) MakeFeature

	SetUsersManager(tag feature.Tag) MakeFeature

	SetUseGravatar(enabled bool) MakeFeature

	SetAuthApiMountPath(prefix string) MakeFeature
	SetAvatarMountPath(prefix string) MakeFeature
	SetAvatarStore(store avatar.Store) MakeFeature

	SetRefreshCache(tag feature.Tag, name, bucket string) MakeFeature
	SetCustomRefreshCache(cache middleware.RefreshCache) MakeFeature

	SetTokenDuration(d time.Duration) MakeFeature
	SetCookieDuration(d time.Duration) MakeFeature

	SetSecureCookies(secure bool) MakeFeature
	SetSameSiteCookie(site http.SameSite) MakeFeature

	SetXSRFHeaderKey(key string) MakeFeature
	SetXSRFCookieName(name string) MakeFeature

	SetJWTQuery(name string) MakeFeature
	SetJWTHeaderKey(key string) MakeFeature
	SetJWTCookieName(name string) MakeFeature
	SetJWTCookieDomain(domain string) MakeFeature
	SetSendJWTHeader(enabled bool) MakeFeature

	AddAudSecretKeys(keys ...string) MakeFeature
	AddProvider(name, cid, csecret string) MakeFeature
	AddDirectProvider(name string, fn provider.CredCheckerFunc) MakeFeature

	SetVerifyEmailAccount(name string) MakeFeature
	SetVerifyEmailTemplate(name string) MakeFeature
	IncludeVerifyEmailProvider(providerName string) MakeFeature

	SetLogLevel(level log.Level) MakeFeature
	EnableDevService(enabled bool) MakeFeature
}

type CFeature struct {
	feature.CFeature

	publicSignups bool

	verifyEmailAccount  string
	verifyEmailTemplate string
	verifyEmailProvider string

	refreshCacheTag    feature.Tag
	refreshCacheName   string
	refreshCacheBucket string

	emailSender   feature.EmailSender
	emailProvider feature.EmailProvider

	mountAuthApiPath string
	mountAvatarPath  string

	audKeys    []string
	audSecrets map[string]string

	authProviders       map[string][]string
	authProvidersDirect map[string]provider.CredCheckerFunc

	authOpts    auth.Opts
	authService *auth.Service

	enableDevService bool

	protectPrefix map[string]*regexp.Regexp
	protectGroups map[string][]string

	ubmTag feature.Tag
	ubm    userbase.Manager
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
	f.audSecrets = make(map[string]string)
	f.authProviders = make(map[string][]string)
	f.authProvidersDirect = make(map[string]provider.CredCheckerFunc)
	f.mountAuthApiPath = "/auth"
	f.mountAvatarPath = "/avatar"
	f.verifyEmailTemplate = DefaultEmailNewTokenTemplate
	f.authOpts = auth.Opts{
		URL:               "",
		Issuer:            "",
		XSRFHeaderKey:     "",
		XSRFCookieName:    "",
		JWTCookieDomain:   "",
		JWTCookieName:     "",
		JWTHeaderKey:      "",
		JWTQuery:          "",
		SecureCookies:     false,
		SameSiteCookie:    http.SameSiteDefaultMode,
		TokenDuration:     DefaultTokenDuration,
		CookieDuration:    DefaultCookieDuration,
		SecretReader:      token.SecretFunc(f.authAudSecretsFunc),
		ClaimsUpd:         token.ClaimsUpdFunc(f.authClaimsUpdFunc),
		Validator:         token.ValidatorFunc(f.authValidatorFunc),
		RefreshCache:      nil,
		AudienceReader:    nil,
		Logger:            log.NewLogf(log.LevelTrace),
		AudSecrets:        false,
		AvatarResizeLimit: 256,
	}
}

func (f *CFeature) SetLogLevel(level log.Level) MakeFeature {
	f.authOpts.Logger = log.NewLogf(level)
	return f
}

func (f *CFeature) SetUrl(url string) MakeFeature {
	f.authOpts.URL = url
	return f
}

func (f *CFeature) SetIssuer(issuer string) MakeFeature {
	f.authOpts.Issuer = issuer
	return f
}

func (f *CFeature) SetUseGravatar(enabled bool) MakeFeature {
	f.authOpts.UseGravatar = enabled
	return f
}

func (f *CFeature) SetAuthApiMountPath(prefix string) MakeFeature {
	f.mountAuthApiPath = prefix
	return f
}

func (f *CFeature) SetAvatarMountPath(prefix string) MakeFeature {
	f.mountAvatarPath = prefix
	return f
}

func (f *CFeature) SetAvatarStore(store avatar.Store) MakeFeature {
	f.authOpts.AvatarStore = store
	return f
}

func (f *CFeature) SetRefreshCache(tag feature.Tag, name, bucket string) MakeFeature {
	if f.authOpts.RefreshCache != nil {
		log.FatalDF(1, "only one refresh cache allowed")
	} else if tag == "" {
		log.FatalDF(1, "tag value is required")
	}
	if name == "" {
		name = "default"
	}
	if bucket == "" {
		name = "default"
	}
	f.refreshCacheTag = tag
	f.refreshCacheName = name
	f.refreshCacheBucket = bucket
	return f
}

func (f *CFeature) SetCustomRefreshCache(cache middleware.RefreshCache) MakeFeature {
	if f.refreshCacheTag != "" && f.refreshCacheName != "" && f.refreshCacheBucket != "" {
		log.FatalDF(1, "only one refresh cache allowed")
	}
	f.authOpts.RefreshCache = cache
	return f
}

func (f *CFeature) SetTokenDuration(d time.Duration) MakeFeature {
	f.authOpts.TokenDuration = d
	return f
}

func (f *CFeature) SetCookieDuration(d time.Duration) MakeFeature {
	f.authOpts.CookieDuration = d
	return f
}

func (f *CFeature) SetSecureCookies(secure bool) MakeFeature {
	f.authOpts.SecureCookies = secure
	return f
}

func (f *CFeature) SetSameSiteCookie(site http.SameSite) MakeFeature {
	f.authOpts.SameSiteCookie = site
	return f
}

func (f *CFeature) SetXSRFHeaderKey(key string) MakeFeature {
	f.authOpts.XSRFHeaderKey = key
	return f
}

func (f *CFeature) SetXSRFCookieName(name string) MakeFeature {
	f.authOpts.XSRFCookieName = name
	return f
}

func (f *CFeature) SetJWTQuery(name string) MakeFeature {
	f.authOpts.JWTQuery = name
	return f
}

func (f *CFeature) SetJWTHeaderKey(key string) MakeFeature {
	f.authOpts.JWTHeaderKey = key
	return f
}

func (f *CFeature) SetJWTCookieName(name string) MakeFeature {
	f.authOpts.JWTCookieName = name
	return f
}

func (f *CFeature) SetJWTCookieDomain(domain string) MakeFeature {
	f.authOpts.JWTCookieDomain = domain
	return f
}

func (f *CFeature) SetSendJWTHeader(enabled bool) MakeFeature {
	f.authOpts.SendJWTHeader = enabled
	return f
}

func (f *CFeature) AddAudSecretKeys(keys ...string) MakeFeature {
	for _, key := range keys {
		if !beStrings.StringInSlices(key, f.audKeys) {
			f.audKeys = append(f.audKeys, key)
		}
	}
	return f
}

func (f *CFeature) AddProvider(name, cid, csecret string) MakeFeature {
	f.authProviders[name] = []string{cid, csecret}
	return f
}

func (f *CFeature) AddDirectProvider(name string, fn provider.CredCheckerFunc) MakeFeature {
	f.authProvidersDirect[name] = fn
	return f
}

func (f *CFeature) SetVerifyEmailAccount(name string) MakeFeature {
	f.verifyEmailAccount = name
	return f
}

func (f *CFeature) SetVerifyEmailTemplate(name string) MakeFeature {
	f.verifyEmailTemplate = name
	return f
}

func (f *CFeature) EnableDevService(enabled bool) MakeFeature {
	f.enableDevService = enabled
	return f
}

func (f *CFeature) IncludeVerifyEmailProvider(providerName string) MakeFeature {
	f.verifyEmailProvider = providerName
	return f
}

func (f *CFeature) SetUsersManager(tag feature.Tag) MakeFeature {
	f.ubmTag = tag
	return f
}

func (f *CFeature) SetPublicSignups(allowed bool) MakeFeature {
	f.publicSignups = allowed
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if f.ubmTag == "" {
		err = fmt.Errorf("%v requires a userbase.Manager feature tag to be set", f.Tag())
		return
	}
	tag := f.Tag().String()
	b.AddFlags(&cli.StringFlag{
		Name:     globals.MakeFlagName(tag, "base-url"),
		Usage:    "specify the auth site base url",
		Category: tag,
		Value:    DefaultUrl,
		EnvVars:  globals.MakeFlagEnvKeys(tag, "base-url"),
	})
	b.AddFlags(&cli.StringFlag{
		Name:     globals.MakeFlagName(tag, "default-aud-secret"),
		Usage:    "specify the default secret key",
		Category: tag,
		Required: len(f.audKeys) == 0,
		EnvVars:  globals.MakeFlagEnvKeys(tag, "default-aud-secret"),
	})
	for _, key := range f.audKeys {
		flagName := strcase.ToKebab(key) + "-aud-secret"
		b.AddFlags(&cli.StringFlag{
			Name:     globals.MakeFlagName(tag, flagName),
			Usage:    "specify the " + key + " audience secret",
			Category: tag,
			Required: true,
			EnvVars:  globals.MakeFlagEnvKeys(tag, flagName),
		})
	}
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.Enjin = enjin
	siteName := strcase.ToKebab(f.Enjin.SiteName())

	f.authOpts.AudSecrets = len(f.audKeys) > 0

	if f.authOpts.Issuer == "" {
		f.authOpts.Issuer = siteName
	}
	if f.authOpts.XSRFHeaderKey == "" {
		f.authOpts.XSRFHeaderKey = strcase.ToScreamingKebab(globals.MakeFlagName("x-"+siteName, "xsrf-token"))
	}
	if f.authOpts.XSRFCookieName == "" {
		f.authOpts.XSRFCookieName = strcase.ToSnake(globals.MakeFlagName(siteName, "xsrf-token"))
	}

	if f.authOpts.JWTHeaderKey == "" {
		f.authOpts.JWTHeaderKey = strcase.ToScreamingKebab(globals.MakeFlagName(siteName, "jwt-token"))
	}
	if f.authOpts.JWTCookieName == "" {
		f.authOpts.JWTCookieName = strcase.ToSnake(globals.MakeFlagName(siteName, "jwt-token"))
	}
	if f.authOpts.JWTQuery == "" {
		f.authOpts.JWTQuery = strcase.ToKebab(globals.MakeFlagName(siteName, "jwt-token"))
	}
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	for _, ef := range f.Enjin.Features() {
		if ubm, ok := ef.(userbase.Manager); ok && ef.Tag() == f.ubmTag {
			f.ubm = ubm
			break
		}
	}
	if f.ubm == nil {
		err = fmt.Errorf("%v error: userbase.Manager (tagged: %v) not found", f.Tag(), f.ubmTag)
		return
	}

	tag := f.Tag().String()

	processAudSecrets := func(ctx *cli.Context, key, name string, strict bool) (err error) {
		if flagName := globals.MakeFlagName(tag, name); ctx.IsSet(flagName) {
			if v := ctx.String(flagName); v != "" {
				f.audSecrets[key] = v
			} else if strict {
				err = fmt.Errorf("missing --" + flagName)
				return
			}
		} else if strict {
			err = fmt.Errorf("missing --" + flagName)
			return
		}
		return
	}

	if err = processAudSecrets(ctx, "_default_", "default-aud-secret", len(f.audKeys) == 0); err != nil {
		return
	}

	for _, key := range f.audKeys {
		if err = processAudSecrets(ctx, key, strcase.ToKebab(key)+"-aud-secret", true); err != nil {
			return
		}
	}

	if flagName := globals.MakeFlagName(tag, "base-url"); ctx.IsSet(flagName) {
		if v := ctx.String(flagName); v != "" {
			f.authOpts.URL = v
		}
	}

	defaultUrl := DefaultUrl
	if _, port := f.Enjin.ServiceInfo(); port != globals.DefaultPort {
		defaultUrl = "http://localhost:" + strconv.Itoa(port)
	}

	if f.authOpts.URL == "" || f.authOpts.URL == defaultUrl {
		log.WarnF("! using default user-auth-api url: %v", defaultUrl)
		f.authOpts.URL = defaultUrl
	}

	if f.authOpts.AvatarStore == nil {
		avatarStoragePath := filepath.Join(os.TempDir(), strcase.ToKebab(tag+"-avatar-storage"))
		if ee := os.MkdirAll(avatarStoragePath, 0770); ee != nil {
			err = fmt.Errorf("error making avatarStoragePath: %v - %v", avatarStoragePath, ee)
			return
		} else {
			f.authOpts.AvatarStore = avatar.NewLocalFS(avatarStoragePath)
			log.DebugF("using default avatar storage path: %v", avatarStoragePath)
		}
	}

	if f.verifyEmailAccount != "" {
		if es := f.Enjin.FindEmailAccount(f.verifyEmailAccount); es == nil {
			err = fmt.Errorf("error finding enjin feature.EmailSender account named: %v", f.verifyEmailAccount)
			return
		} else {
			f.emailSender = es

			var ep feature.EmailProvider
			for _, ef := range f.Enjin.Features() {
				if v, ok := ef.Self().(feature.EmailProvider); ok {
					ep = v
					break
				}
			}
			if ep == nil {
				err = fmt.Errorf("feature.EmailProvider not found")
				return
			} else {
				f.emailProvider = ep
			}

			log.DebugF("found feature.EmailSender and feature.EmailProvider")
		}
	}

	if f.authOpts.RefreshCache == nil && f.refreshCacheTag != "" && f.refreshCacheBucket != "" {
		for _, ef := range f.Enjin.Features() {
			if kvcf, ok := ef.Self().(beKvs.KeyValueCaches); ok && ef.Tag() == f.refreshCacheTag {
				if kvc, ee := kvcf.Get(f.refreshCacheName); ee == nil {
					if kvs, eee := kvc.GetBucket(f.refreshCacheBucket); eee == nil {
						f.authOpts.RefreshCache = beKvs.NewKVSA(kvs)
						log.DebugF("using refresh cache: %v/%s/%s", f.refreshCacheTag, f.refreshCacheName, f.refreshCacheBucket)
						break
					}
				}
			}
		}
	}

	f.authService = auth.NewService(f.authOpts)

	for name, argv := range f.authProviders {
		f.authService.AddProvider(name, argv[0], argv[1])
	}

	for name, fn := range f.authProvidersDirect {
		f.authService.AddDirectProvider(name, fn)
	}

	if f.verifyEmailProvider != "" {
		if f.verifyEmailAccount == "" {
			err = fmt.Errorf(".SetVerifyEmailAccount and have at least one enjin feature.EmailProvider are required for the verify-email provider")
			return
		}
		log.DebugF("adding user auth api email provider: %v", f.verifyEmailProvider)
		f.authAddVerifyEmailProviderFunc(f.verifyEmailProvider)
	}

	if f.enableDevService {
		if beStrings.StringInStrings(f.Enjin.Prefix(), "prd", "") {
			err = fmt.Errorf("cannot run user auth api dev service with production environments")
			return
		} else {
			log.WarnF("!!!!! INCLUDING USER AUTH API DEV SERVICE !!!!!")
		}
		go func() {
			if devAuthServer, devAuthServerErr := f.authService.DevAuth(); devAuthServerErr != nil {
				log.FatalF("error starting dev auth service: %v", devAuthServerErr)
			} else {
				devAuthServer.GetEmailFn = func(username string) string {
					return username + "@localhost.nope"
				}
				devAuthServer.Run(context.Background())
			}
		}()
	}

	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) Use(s feature.System) feature.MiddlewareFn {
	return func(next http.Handler) http.Handler {
		authenticator := f.authService.Middleware()
		return authenticator.Trace(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f.AuthApiServeHTTP(next, w, r)
		}))
	}
}

func (f *CFeature) Apply(s feature.System) (err error) {
	router := s.Router()
	authenticator := f.authService.Middleware()
	router.Use(authenticator.Trace)
	authRoutes, avatarRoutes := f.authService.Handlers()
	router.Mount(f.mountAuthApiPath, authRoutes)
	router.Mount(f.mountAvatarPath, avatarRoutes)
	return
}

func (f *CFeature) RequireApiUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if au := userbase.GetCurrentAuthUser(r); au != nil {
			log.WarnRF(r, "RequireApiUser not implemented, allowing all users")
			next.ServeHTTP(w, r)
		} else {
			f.Enjin.Serve404(w, r)
		}
	})
}

func (f *CFeature) RequireUserCan(action userbase.Action) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if au := userbase.GetCurrentAuthUser(r); au != nil {
				log.WarnRF(r, "RequireUserCan not implemented, allowing all users to: %v", action)
				next.ServeHTTP(w, r)
			} else {
				f.Enjin.Serve404(w, r)
			}
		})
	}
}