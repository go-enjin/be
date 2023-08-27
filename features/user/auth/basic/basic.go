//go:build user_auth_basic || user_auths || all

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

package basic

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"

	auth "github.com/abbot/go-http-auth"
	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	beNet "github.com/go-enjin/be/pkg/net"
	"github.com/go-enjin/be/pkg/net/serve"
	"github.com/go-enjin/be/pkg/slices"
	"github.com/go-enjin/be/pkg/userbase"
)

const (
	Tag            feature.Tag = "user-auth-basic"
	UserContextKey             = beContext.RequestKey("user-auth-basic-user-key")
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	userbase.AuthProvider
	feature.PageRestrictionHandler
	feature.DataRestrictionHandler
}

type MakeFeature interface {
	Make() Feature

	SetRealm(realm string) MakeFeature
	SetLogoutPath(path string) MakeFeature
	SetLogoutRedirectPath(path string) MakeFeature
	SetAuthCacheControl(control string) MakeFeature

	AddUserbase(usersProvider, groupsProvider, secretsProvider string) MakeFeature

	Ignore(patterns ...string) MakeFeature
	Protect(pattern, group string) MakeFeature
	ProtectAll(group string) MakeFeature
	AddBypassIP(ip ...string) MakeFeature
	AddBypassCIDR(cidr ...string) MakeFeature
}

type CFeature struct {
	feature.CFeature

	cacheControl string

	realm          string
	logoutPath     string
	redirectPath   string
	protectAll     string
	ignoredPaths   []*regexp.Regexp
	protectedPaths []*protectedPath

	bypassingIPs   []net.IP
	bypassingCIDRs []*net.IPNet

	upNames          []string
	gpNames          []string
	spNames          []string
	usersProviders   []userbase.AuthUserProvider
	groupsProviders  []userbase.GroupsProvider
	secretsProviders []userbase.SecretsProvider
}

type protectedPath struct {
	group   string
	pattern *regexp.Regexp
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

func (f *CFeature) UsageNotes() (notes []string) {
	category := f.Tag().String()
	patternKey := globals.MakeFlagEnvKey(category, "PROTECT_PATH_REGEX")
	groupKey := globals.MakeFlagEnvKey(category, "PROTECT_PATH_GROUP")

	notes = []string{
		"this feature supports dynamically restricting content through environment variables",
		"two variables are required to protect any given path",
		"make as many pairs to protect all the paths required",
		"the pair is:",
		patternKey + "_<KEY>=<PATTERN>",
		groupKey + "_<KEY>=<GROUPS>",
		"<KEY>     is a unique identifier for the pair",
		"<PATTERN> is a regular expression",
		"<GROUPS>  is a space separated list of groups allowed",
	}
	return
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.realm = "-"
	f.logoutPath = "/logout"
	f.redirectPath = "/"
	f.cacheControl = "no-store"
}

func (f *CFeature) SetRealm(realm string) MakeFeature {
	f.realm = realm
	return f
}

func (f *CFeature) SetLogoutPath(path string) MakeFeature {
	f.logoutPath = path
	return f
}

func (f *CFeature) SetLogoutRedirectPath(path string) MakeFeature {
	f.redirectPath = path
	return f
}

func (f *CFeature) SetAuthCacheControl(control string) MakeFeature {
	f.cacheControl = control
	return f
}

func (f *CFeature) AddUserbase(usersProvider, groupsProvider, secretsProvider string) MakeFeature {
	if usersProvider != "" && !slices.Within(usersProvider, f.upNames) {
		f.upNames = append(f.upNames, usersProvider)
	}
	if groupsProvider != "" && !slices.Within(groupsProvider, f.gpNames) {
		f.gpNames = append(f.gpNames, groupsProvider)
	}
	if secretsProvider != "" && !slices.Within(secretsProvider, f.spNames) {
		f.spNames = append(f.spNames, secretsProvider)
	}
	return f
}

func (f *CFeature) Ignore(patterns ...string) MakeFeature {
	for _, pattern := range patterns {
		if compiled, err := regexp.Compile(pattern); err != nil {
			log.FatalDF(1, "error compiling regexp: %v", err)
		} else {
			f.ignoredPaths = append(f.ignoredPaths, compiled)
		}
	}
	return f
}

func (f *CFeature) Protect(pattern, group string) MakeFeature {
	if compiled, err := regexp.Compile(pattern); err != nil {
		log.FatalDF(1, "error compiling regexp: %v", err)
	} else {
		f.protectedPaths = append(f.protectedPaths, &protectedPath{
			group:   group,
			pattern: compiled,
		})
	}
	return f
}

func (f *CFeature) ProtectAll(group string) MakeFeature {
	f.protectAll = group
	return f
}

func (f *CFeature) AddBypassIP(addresses ...string) MakeFeature {
	for _, address := range addresses {
		if parsed := net.ParseIP(address); parsed == nil {
			log.FatalDF(1, "invalid address: %v", address)
		} else {
			f.bypassingIPs = append(f.bypassingIPs, parsed)
		}
	}
	return f
}

func (f *CFeature) AddBypassCIDR(ranges ...string) MakeFeature {
	for _, cidr := range ranges {
		if _, parsed, err := net.ParseCIDR(cidr); err != nil {
			log.FatalDF(1, "invalid CIDR: %v - err", cidr, err)
		} else {
			f.bypassingCIDRs = append(f.bypassingCIDRs, parsed)
		}
	}
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	category := f.Tag().String()

	b.AddFlags(
		&cli.StringFlag{
			Name:     globals.MakeFlagName(category, "realm"),
			Usage:    "specify the basic auth realm",
			EnvVars:  globals.MakeFlagEnvKeys(category, "REALM"),
			Category: category,
		},
		&cli.StringFlag{
			Name:     globals.MakeFlagName(category, "protect-all"),
			Usage:    "specify group required for all requests",
			EnvVars:  globals.MakeFlagEnvKeys(category, "PROTECT_ALL"),
			Category: category,
		},
		&cli.StringSliceFlag{
			Name:     globals.MakeFlagName(category, "bypass-addrs"),
			Usage:    "space separated list of IPs which bypass protections",
			EnvVars:  globals.MakeFlagEnvKeys(category, "BYPASS_ADDRS"),
			Category: category,
		},
		&cli.StringSliceFlag{
			Name:     globals.MakeFlagName(category, "bypass-cidrs"),
			Usage:    "space separated list of CIDR ranges which bypass protections",
			EnvVars:  globals.MakeFlagEnvKeys(category, "BYPASS_CIDRS"),
			Category: category,
		},
		&cli.StringFlag{
			Name:     globals.MakeFlagName(category, "auth-cache-control"),
			Usage:    "specify logged in session cache-control header",
			Value:    "no-store",
			EnvVars:  globals.MakeFlagEnvKeys(category, "AUTH_CACHE_CONTROL"),
			Category: category,
		},
	)
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	report := func(conditional bool, format string, argv ...interface{}) {
		if conditional {
			log.DebugDF(1, format, argv...)
		}
	}

	category := f.Tag().String()

	if flagName := globals.MakeFlagName(category, "realm"); ctx.IsSet(flagName) {
		if v, ok := ctx.Value(flagName).(string); ok {
			if v == "" && f.realm == "" {
				f.realm = "-"
			} else {
				f.realm = v
			}
		}
	}
	if f.realm == "-" {
		report(true, "realm set to: (auto)")
	} else {
		report(f.realm != "", "realm set to: %v", f.realm)
	}

	if flagName := globals.MakeFlagName(category, "protect-all"); ctx.IsSet(flagName) {
		if v, ok := ctx.Value(flagName).(string); ok {
			f.protectAll = v
		}
	}
	report(f.protectAll != "", `all requests require access group of: "%v"`, f.protectAll)

	if flagName := globals.MakeFlagName(category, "bypass-addrs"); ctx.IsSet(flagName) {
		if v := ctx.StringSlice(flagName); len(v) > 0 {
			for _, list := range v {
				for _, s := range strings.Split(list, " ") {
					if parsed := net.ParseIP(s); parsed != nil {
						f.bypassingIPs = append(f.bypassingIPs, parsed)
					} else {
						log.ErrorF("error parsing --%v [%v]: not a net.IP address", flagName, s)
					}
				}
			}
		}
	}
	report(len(f.bypassingIPs) > 0, "bypassing with %d IP addresses: %+v", len(f.bypassingIPs), f.bypassingIPs)

	if flagName := globals.MakeFlagName(category, "bypass-cidrs"); ctx.IsSet(flagName) {
		if v := ctx.StringSlice(flagName); len(v) > 0 {
			for _, list := range v {
				for _, s := range strings.Split(list, " ") {
					if _, ipNet, ee := net.ParseCIDR(s); ee == nil {
						f.bypassingCIDRs = append(f.bypassingCIDRs, ipNet)
					} else {
						log.ErrorF("error parsing --%v [%v]: not a valid CIDR notation", flagName, s)
					}
				}
			}
		}
	}
	report(len(f.bypassingCIDRs) > 0, "bypassing with %d IP ranges: %+v", len(f.bypassingCIDRs), f.bypassingCIDRs)

	if flagName := globals.MakeFlagName(category, "auth-cache-control"); ctx.IsSet(flagName) {
		if v, ok := ctx.Value(flagName).(string); ok {
			f.cacheControl = v
		}
	}

	// patternKey := globals.MakeFlagEnvKey(category, "PROTECT_PATH_REGEX")
	// groupKey := globals.MakeFlagEnvKey(category, "PROTECT_PATH_GROUP")

	foundPatterns := make(map[string]string)
	foundGroups := make(map[string]string)
	for _, setting := range os.Environ() {
		parts := strings.Split(setting, "=")
		if len(parts) < 2 {
			log.ErrorF("os.Environ returned with invalid format: \"%s\"", setting)
			continue
		}
		envKey, value := parts[0], strings.Join(parts[1:], "=")
		if rxProtectPath.MatchString(envKey) {
			m := rxProtectPath.FindAllStringSubmatch(envKey, 1)
			name := m[0][1]
			key := m[0][2]
			if name == "GROUP" {
				foundGroups[key] = value
			} else if name == "REGEX" {
				foundPatterns[key] = value
			}
		}
	}
	var brokenPairs []string
	for fpk, fpv := range foundPatterns {
		if fgv, ok := foundGroups[fpk]; ok {
			f.Protect(fpv, fgv)
			log.DebugF(`"%v" group required for access to: "%v"`, fgv, fpv)
		} else {
			brokenPairs = append(brokenPairs, fpk)
		}
	}
	for fgk, _ := range foundGroups {
		if _, ok := foundPatterns[fgk]; !ok {
			if !slices.Within(fgk, brokenPairs) {
				brokenPairs = append(brokenPairs, fgk)
			}
		}
	}
	if len(brokenPairs) > 0 {
		err = fmt.Errorf("the following protect-path keys are missing either the REGEX or GROUP in the expected pair: %v", brokenPairs)
		return
	}

	for _, ef := range f.Enjin.Features().List() {
		efTag := ef.Tag().String()
		for _, upName := range f.upNames {
			if efTag == upName {
				if upf, ok := feature.AsTyped[userbase.AuthUserProvider](ef); ok {
					f.usersProviders = append(f.usersProviders, upf)
				} else {
					err = fmt.Errorf("%v feature is not a userbase.AuthUserProvider", efTag)
					return
				}
			}
		}
		for _, gpName := range f.gpNames {
			if efTag == gpName {
				if gpf, ok := feature.AsTyped[userbase.GroupsProvider](ef); ok {
					f.groupsProviders = append(f.groupsProviders, gpf)
				} else {
					err = fmt.Errorf("%v feature is not a userbase.GroupsProvider", efTag)
					return
				}
			}
		}
		for _, spName := range f.spNames {
			if efTag == spName {
				if spf, ok := feature.AsTyped[userbase.SecretsProvider](ef); ok {
					f.secretsProviders = append(f.secretsProviders, spf)
				} else {
					err = fmt.Errorf("%v feature is not a userbase.SecretsProvider", efTag)
					return
				}
			}
		}
	}

	if len(f.usersProviders) == 0 {
		err = fmt.Errorf("at least one userbase.AuthUserProvider is required")
		return
	} else if len(f.groupsProviders) == 0 {
		err = fmt.Errorf("at least one userbase.GroupsProvider is required")
		return
	} else if len(f.secretsProviders) == 0 {
		err = fmt.Errorf("at least one userbase.SecretsProvider is required")
		return
	}
	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) makeAuthenticator(realm string) (authenticator *auth.BasicAuth) {
	authenticator = auth.NewBasicAuthenticator(realm, f.getUserSecret)
	if authenticator.Headers == nil {
		authenticator.Headers = auth.NormalHeaders
	}
	return
}

func (f *CFeature) AuthenticateRequest(w http.ResponseWriter, r *http.Request) (handled bool, modified *http.Request) {
	var realm string
	if realm = f.realm; realm == "-" {
		realm = r.Host
	}

	authenticator := f.makeAuthenticator(realm)

	validated := authenticator.CheckAuth(r)
	if v := f.getUser(validated, realm); v == nil {
		validated = ""
	}

	if r.URL.Path == f.logoutPath {
		redirectToUrl := beNet.MakeUrl(f.redirectPath, r)
		r.Header.Set("Cache-Control", "no-store")
		authenticator.Headers.UnauthContentType = LogoutRedirectContentType
		authenticator.Headers.UnauthResponse = fmt.Sprintf(LogoutRedirectResponseFormat, redirectToUrl)
		authenticator.RequireAuth(w, r)
		log.DebugRF(r, "logged out of basic auth: %v", validated)
		return true, nil
	}

	modifyRequest := func(validated string, w http.ResponseWriter, r *http.Request) *http.Request {
		if validated != "" {
			r = r.Clone(context.WithValue(r.Context(), UserContextKey, validated))
		}
		if f.cacheControl != "" {
			r = serve.SetCacheControl(f.cacheControl, w, r)
		}
		return r
	}

	if group, protected := f.isRequestProtected(r); protected {
		if validated != "" {
			if f.isUserInGroup(validated, group) {
				log.DebugRF(r, "user has access to protected content: user=%v need=%v realm=%v path=%v", validated, group, realm, r.URL.Path)
				return false, modifyRequest(validated, w, r)
			}
			log.DebugRF(r, "user requesting access to protected content: user=%v need=%v realm=%v path=%v", validated, group, realm, r.URL.Path)
		}
		r.Header.Set("Cache-Control", "no-store")
		authenticator.Headers.UnauthContentType = UnauthContentType
		authenticator.Headers.UnauthResponse = UnauthResponse
		authenticator.RequireAuth(w, r)
		log.DebugRF(r, "requiring basic auth: %v - %v", group, r.URL.Path)
		return true, nil
	}

	return false, modifyRequest(validated, w, r)
}

func (f *CFeature) RestrictServePage(pgCtx beContext.Context, w http.ResponseWriter, r *http.Request) (modCtx beContext.Context, modReq *http.Request, allow bool) {
	modReq = r
	modCtx = pgCtx.Copy()

	if group, protected := f.isRequestProtected(r); protected {
		var id string
		if id = f.getRequestUserID(r); id != "" {
			if allow = f.isUserInGroup(id, group); allow {
				tag := f.Tag().Camel()
				modCtx[tag+"UserID"] = id
				modCtx[tag+"User"] = userbase.NewAuthUser(id, id, "", "", beContext.Context{"ID": id, "GetName": id})
				if f.cacheControl != "" {
					if _, exists := modCtx["CacheControl"]; !exists {
						modCtx["CacheControl"] = f.cacheControl
					}
				}
				return
			}
		}
		f.Enjin.Serve403(w, r)
		return
	}

	allow = true
	return
}

func (f *CFeature) RestrictServeData(data []byte, mime string, w http.ResponseWriter, r *http.Request) (modReq *http.Request, allow bool) {
	modReq = r
	if group, protected := f.isRequestProtected(modReq); protected {
		if id := f.getRequestUserID(modReq); id != "" {
			if allow = f.isUserInGroup(id, group); allow {
				return
			}
		}
		f.Enjin.Serve403(w, modReq)
		return
	}
	allow = true
	return
}

func (f *CFeature) getRequestUserID(r *http.Request) (id string) {
	if v, ok := r.Context().Value(UserContextKey).(string); ok {
		id = v
	}
	return
}

func (f *CFeature) isRequestBypassed(r *http.Request) (bypass bool) {
	if len(f.bypassingIPs) > 0 || len(f.bypassingCIDRs) > 0 {
		if ip, err := beNet.ParseIpFromRequest(r); err == nil {

			for _, safe := range f.bypassingIPs {
				if bypass = ip.Equal(safe); bypass {
					return
				}
			}

			for _, safe := range f.bypassingCIDRs {
				if bypass = safe.Contains(ip); bypass {
					return
				}
			}

		} else {
			log.ErrorRF(r, "error parsing ip from request: %v", err)
		}
	}
	return
}

func (f *CFeature) isRequestProtected(r *http.Request) (group userbase.Group, protected bool) {

	if f.isRequestIgnored(r.URL.Path) {
		return
	}

	if f.isRequestBypassed(r) {
		return
	}

	for _, pp := range f.protectedPaths {
		if protected = pp.pattern.MatchString(r.URL.Path); protected {
			group = userbase.NewGroup(pp.group)
			return
		}
	}

	if f.protectAll != "" {
		group = userbase.NewGroup(f.protectAll)
		protected = true
		return
	}

	return
}

func (f *CFeature) isRequestIgnored(path string) (ignored bool) {
	if f.protectAll == "" {
		// not all are restricted
		if ignored = len(f.protectedPaths) == 0; ignored {
			// no protections specified, early out
			return
		}
	}

	for _, rx := range f.ignoredPaths {
		if ignored = rx.MatchString(path); ignored {
			return
		}
	}
	return
}

func (f *CFeature) getUser(id, _ string) (user *userbase.AuthUser) {
	for _, upf := range f.usersProviders {
		if found, err := upf.GetAuthUser(id); err == nil && found != nil {
			user = found
			return
		}
	}
	return
}

func (f *CFeature) getUserSecret(user, _ string) (secret string) {
	for _, spf := range f.secretsProviders {
		if secret = spf.GetUserSecret(user); secret != "" {
			return
		}
	}
	return
}

func (f *CFeature) isUserInGroup(user string, group userbase.Group) (present bool) {
	for _, gpf := range f.groupsProviders {
		if present = gpf.IsUserInGroup(user, group); present {
			return
		}
	}
	return
}

var rxProtectPath = regexp.MustCompile(`^\s*[_A-Z0-9]+?PROTECT_PATH_(REGEX|GROUP)_([A-Z0-9]+[_A-Z0-9]+?)\s*$`)

var UnauthContentType = "text/plain"
var UnauthResponse = "401 - " + http.StatusText(http.StatusUnauthorized)
var LogoutRedirectContentType = "text/html"
var LogoutRedirectResponseFormat = `<!DOCTYPE html>
<html lang="en">
<head><meta http-equiv="refresh" content="0; url=%[1]s" /></head>
<body><a href="%[1]s">%[1]s</a>
</html>`