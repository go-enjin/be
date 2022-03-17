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

//go:build restricted || all

package basic

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/abbot/go-http-auth"
	"github.com/iancoleman/strcase"
	"github.com/tg123/go-htpasswd"
	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

var _basic *Feature

var (
	_ feature.Feature                = (*Feature)(nil)
	_ feature.PageRestrictionHandler = (*Feature)(nil)
	_ feature.DataRestrictionHandler = (*Feature)(nil)
)

const Tag feature.Tag = "restricted"

type Feature struct {
	feature.CFeature

	realm string

	restrictAllPages bool
	restrictAllData  bool
	restrictDataMime []string

	enableEnv bool
	envUsers  map[string]string
	envGroup  map[string][]string

	htpFiles map[string]auth.SecretProvider
	htGroups map[string]*htpasswd.HTGroup

	authenticator *auth.BasicAuth
}

type MakeFeature interface {
	feature.MakeFeature

	Realm(name string) MakeFeature

	RestrictAll(enabled bool) MakeFeature
	RestrictAllData(enabled bool) MakeFeature
	RestrictDataMime(mime string) MakeFeature

	Htpasswd(paths ...string) MakeFeature
	Htgroups(paths ...string) MakeFeature

	EnableEnv(enabled bool) MakeFeature
	AddEnvUser(name, password string) MakeFeature
	AddEnvGroup(name string, users ...string) MakeFeature
	AddEnvUserGroups(name string, groups ...string) MakeFeature
}

func New() MakeFeature {
	if _basic == nil {
		_basic = new(Feature)
		_basic.Init(_basic)
	}
	return _basic
}

func (f *Feature) Realm(name string) MakeFeature {
	f.realm = name
	return f
}

func (f *Feature) RestrictAll(enabled bool) MakeFeature {
	f.restrictAllPages = enabled
	return f
}

func (f *Feature) RestrictAllData(enabled bool) MakeFeature {
	f.restrictAllData = enabled
	return f
}

func (f *Feature) RestrictDataMime(mime string) MakeFeature {
	mime = beStrings.GetBasicMime(mime)
	if !beStrings.StringInStrings(mime, f.restrictDataMime...) {
		f.restrictDataMime = append(f.restrictDataMime, mime)
	}
	return f
}

func (f *Feature) Htpasswd(paths ...string) MakeFeature {
	for _, path := range paths {
		f.htpFiles[path] = nil
	}
	return f
}

func (f *Feature) Htgroups(paths ...string) MakeFeature {
	for _, path := range paths {
		f.htGroups[path] = nil
	}
	return f
}

func (f *Feature) EnableEnv(enabled bool) MakeFeature {
	f.enableEnv = enabled
	return f
}

func (f *Feature) AddEnvUser(name, password string) MakeFeature {
	name = strcase.ToKebab(name)
	if _, ok := f.envUsers[name]; ok {
		log.WarnF("overwriting env user: %v", name)
	}
	f.envUsers[name] = password
	log.DebugF("added env user: %v", name)
	return f
}

func (f *Feature) AddEnvGroup(name string, users ...string) MakeFeature {
	if len(users) > 0 {
		name = strcase.ToKebab(name)
		for _, user := range users {
			user = strcase.ToKebab(user)
			if !beStrings.StringInStrings(user, f.envGroup[name]...) {
				f.envGroup[name] = append(f.envGroup[name], user)
			}
		}
		log.DebugF("added env group %v: %v", name, f.envGroup[name])
	} else {
		log.WarnF("not adding empty env group: %v", name)
	}
	return f
}

func (f *Feature) AddEnvUserGroups(name string, groups ...string) MakeFeature {
	if !f.IsEnvUser(name) {
		log.FatalF("user \"%v\" not found", name)
		return nil
	}
	for _, group := range groups {
		if group != "" {
			group = strcase.ToKebab(group)
			if !beStrings.StringInStrings(name, f.envGroup[group]...) {
				f.envGroup[group] = append(f.envGroup[group], name)
			}
		}
	}
	return f
}

func (f *Feature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.htpFiles = make(map[string]auth.SecretProvider)
	f.htGroups = make(map[string]*htpasswd.HTGroup)
	f.authenticator = nil
	f.realm = "Restricted Content"
	f.envUsers = make(map[string]string)
	f.envGroup = make(map[string][]string)
}

func (f *Feature) IsEnvUser(name string) (ok bool) {
	_, ok = f.envUsers[name]
	return
}

func (f *Feature) Tag() (tag feature.Tag) {
	tag = Tag
	return
}

func (f *Feature) Build(b feature.Buildable) (err error) {
	// add cli flags to the Buildable
	b.AddFlags(
		&cli.BoolFlag{
			Name:    "restrict-with-basic-env",
			Usage:   "enable adding users and groups via environment variables",
			EnvVars: []string{globals.EnvPrefix + "_RESTRICT_WITH_BASIC_ENV"},
		},
		&cli.BoolFlag{
			Name:    "restrict-basic-all-pages",
			Usage:   "all page requests are considered non-public",
			EnvVars: []string{globals.EnvPrefix + "_RESTRICT_BASIC_ALL_PAGES"},
		},
		&cli.BoolFlag{
			Name:    "restrict-basic-all-data",
			Usage:   "all data requests are considered non-public",
			EnvVars: []string{globals.EnvPrefix + "_RESTRICT_BASIC_ALL_DATA"},
		},
		&cli.StringFlag{
			Name:        "restrict-basic-realm",
			Usage:       "realm value to use for ht-digest authentication",
			DefaultText: "Restricted Content",
			EnvVars:     []string{globals.EnvPrefix + "_RESTRICT_BASIC_REALM"},
		},
		&cli.StringSliceFlag{
			Name:    "restrict-basic-htpasswd",
			Usage:   "provide one or more htpasswd files",
			EnvVars: []string{globals.EnvPrefix + "_RESTRICT_BASIC_HTPASSWD"},
		},
		&cli.StringSliceFlag{
			Name:    "restrict-basic-htgroups",
			Usage:   "provide one or more htgroups files",
			EnvVars: []string{globals.EnvPrefix + "_RESTRICT_BASIC_HTGROUPS"},
		},
	)
	// configure feature specific internals
	return
}

var rxSplitEquals = regexp.MustCompile(`^\s*([^=]+?)\s*=\s*(.+?)\s*$`)

func (f *Feature) Startup(ctx *cli.Context) (err error) {
	if ctx.Bool("restrict-with-basic-env") {
		f.enableEnv = true
	}

	if ctx.Bool("restrict-basic-all-pages") {
		f.restrictAllPages = true
	}

	if ctx.Bool("restrict-basic-all-data") {
		f.restrictAllData = true
	}

	realm := f.realm
	if ctx.IsSet("restrict-basic-realm") {
		realm = ctx.String("restrict-basic-realm")
	}
	f.authenticator = auth.NewBasicAuthenticator(realm, f.secretsProvider)

	htpFiles := ctx.StringSlice("restrict-basic-htpasswd")
	for path, _ := range f.htpFiles {
		if !beStrings.StringInStrings(path, htpFiles...) {
			htpFiles = append(htpFiles, path)
		}
	}

	for _, path := range htpFiles {
		f.htpFiles[path] = auth.HtpasswdFileProvider(path)
		log.DebugF("loaded htpasswd file: %v", path)
	}

	htGroups := ctx.StringSlice("restrict-basic-htgroups")
	for path, _ := range f.htGroups {
		if !beStrings.StringInStrings(path, htGroups...) {
			htGroups = append(htGroups, path)
		}
	}

	for _, path := range htGroups {
		if f.htGroups[path], err = htpasswd.NewGroups(path, nil); err != nil {
			return
		}
		log.DebugF("loaded htgroups file: %v", path)
	}

	if f.enableEnv {
		environ := os.Environ()
		var loadedUsers []string
		for _, env := range environ {
			if parts := rxSplitEquals.FindStringSubmatch(env); len(parts) == 3 {
				if name, password, ok := f.parseEnvUser(parts[1], parts[2]); ok {
					log.DebugF("parsed env user: '%v', '%v'", parts[1], parts[2])
					f.AddEnvUser(name, password)
					loadedUsers = append(loadedUsers, name)
				}
			}
		}
		var loadedGroups []string
		for _, env := range environ {
			if parts := rxSplitEquals.FindStringSubmatch(env); len(parts) == 3 {
				if name, users, ok := f.parseEnvGroup(parts[1], parts[2]); ok {
					for _, user := range users {
						f.AddEnvUserGroups(user, name)
					}
					loadedGroups = append(loadedGroups, name)
				}
			}
		}
		log.DebugF("found %d env users: %v", len(loadedUsers), loadedUsers)
		log.DebugF("found %d env groups: %v", len(loadedGroups), loadedGroups)
	}
	return
}

func (f *Feature) parseEnvUser(key, value string) (name, password string, ok bool) {
	check := fmt.Sprintf("%v_RESTRICT_BASIC_USER_", globals.EnvPrefix)
	check = strcase.ToKebab(check)
	lc := len(check)
	name = strcase.ToKebab(key)
	ln := len(name)
	if lc < ln {
		if name[0:lc] == check {
			name = name[lc:]
			password = value
			ok = true
			log.DebugF("user: %v, pass: %v", name, password)
		}
	}
	return
}

func (f *Feature) parseEnvGroup(key, value string) (name string, users []string, ok bool) {
	check := fmt.Sprintf("%v_RESTRICT_BASIC_GROUP_", globals.EnvPrefix)
	check = strcase.ToKebab(check)
	lc := len(check)
	name = strcase.ToKebab(key)
	ln := len(name)
	if lc < ln {
		if name[0:lc] == check {
			name = name[lc:]
			for _, user := range strings.Split(value, ",") {
				user = strcase.ToKebab(user)
				if !beStrings.StringInStrings(user, users...) {
					users = append(users, user)
				}
			}
			ok = len(users) > 0
			log.DebugF("group: %v, users: %v", name, users)
		}
	}
	return
}

func (f *Feature) RestrictServePage(ctx beContext.Context, w http.ResponseWriter, r *http.Request) (co beContext.Context, ro *http.Request, allow bool) {
	co = ctx.Copy()
	allow = true
	delete(co, "BasicUsername")
	var restricted []string

	if ctx.Has("RestrictBasic") {
		restricted = ctx.StringOrStrings("RestrictBasic")
		log.DebugF("page has restrictions set: %v", restricted)
	} else if f.restrictAllPages {
		restricted = append(restricted, "users")
		log.DebugF("default page restrictions: %v", restricted)
	}

	reqCtx := f.authenticator.NewContext(context.Background(), r)
	authInfo := auth.FromContext(reqCtx)
	authInfo.UpdateHeaders(w.Header())

	if len(restricted) > 0 {
		// restricted is a list of group names which the user must have at, least one of in their grouping
		restricted = beStrings.StringsToKebabs(restricted...)
		allow = false

		// all users have "public" (anonymous), no auth required
		if beStrings.StringInStrings("public", restricted...) {
			allow = true
		} else {
			// accept login or reject
			if authInfo == nil || !authInfo.Authenticated {
				allow = false
				co["BasicUsername"] = "anonymous"
				ro = r.WithContext(context.WithValue(r.Context(), "BasicUsername", "anonymous"))
				ro = ro.WithContext(context.WithValue(ro.Context(), "BasicDenied", true))
				return
			}
			// all logged in users also have "user"
			hasAtLeastOne := false
			userGroups := f.groupsProvider(authInfo.Username)
			for _, restrict := range restricted {
				if beStrings.StringInStrings(restrict, "user", "users") || beStrings.StringInStrings(restrict, userGroups...) {
					hasAtLeastOne = true
					break
				}
			}
			if hasAtLeastOne {
				allow = true
			}
		}
	}

	if authInfo != nil && authInfo.Authenticated {
		co["BasicUsername"] = authInfo.Username
		ro = r.WithContext(context.WithValue(r.Context(), "BasicUsername", authInfo.Username))
		ro = ro.WithContext(context.WithValue(ro.Context(), "BasicDenied", false))
		return
	}
	ro = r.WithContext(context.WithValue(r.Context(), "BasicUsername", "anonymous"))
	ro = ro.WithContext(context.WithValue(ro.Context(), "BasicDenied", true))
	return
}

func (f *Feature) RestrictServeData(data []byte, mime string, w http.ResponseWriter, r *http.Request) (out *http.Request, allow bool) {
	if v := r.Context().Value("BasicDenied"); v != nil {
		// ServeData is fundamental to all requests, including Serve403
		// if BasicDenied exists, regardless of value, this needs to pass
		out = r
		allow = true
		log.DebugF("basic denied present: %v", v)
		return
	}
	log.DebugF("basic denied not present")

	reqCtx := f.authenticator.NewContext(context.Background(), r)
	authInfo := auth.FromContext(reqCtx)
	authInfo.UpdateHeaders(w.Header())
	if f.restrictAllData {
		allow = authInfo != nil && authInfo.Authenticated
	} else if len(f.restrictDataMime) > 0 {
		mime = beStrings.GetBasicMime(mime)
		for _, rm := range f.restrictDataMime {
			if mime == rm {
				allow = authInfo != nil && authInfo.Authenticated
				break
			}
		}
	}

	if allow {
		out = r.WithContext(context.WithValue(r.Context(), "BasicUsername", authInfo.Username))
		out = out.WithContext(context.WithValue(out.Context(), "BasicDenied", false))
		return
	}
	out = r.WithContext(context.WithValue(r.Context(), "BasicUsername", "anonymous"))
	out = out.WithContext(context.WithValue(out.Context(), "BasicDenied", true))
	return
}

func (f *Feature) groupsProvider(user string) (groups []string) {
	if f.enableEnv {
		for group, users := range f.envGroup {
			if beStrings.StringInStrings(user, users...) {
				if !beStrings.StringInStrings(group, groups...) {
					groups = append(groups, group)
				}
			}
		}
	}
	for _, htg := range f.htGroups {
		for _, group := range htg.GetUserGroups(user) {
			group = strcase.ToKebab(group)
			if !beStrings.StringInStrings(group, groups...) {
				groups = append(groups, group)
			}
		}
	}
	return
}

func (f *Feature) secretsProvider(user, realm string) (secret string) {
	if f.enableEnv {
		if v, ok := f.envUsers[user]; ok && v != "" {
			secret = v
			log.DebugF("env provided user %v", user)
			return
		}
	}
	for file, htp := range f.htpFiles {
		if secret = htp(user, realm); secret != "" {
			log.DebugF("%v provided user %v", file, user)
			return
		}
	}
	log.DebugF("user %v not found", user)
	return
}