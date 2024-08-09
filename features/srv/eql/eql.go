//go:build srv_eql || eql || all

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

package eql

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/context"
	"github.com/go-corelibs/enjinql"
	"github.com/go-corelibs/env"
	"github.com/go-corelibs/go-sqlbuilder"
	"github.com/go-corelibs/go-sqlbuilder/dialects"
	clPath "github.com/go-corelibs/path"
	"github.com/go-corelibs/slices"
	"github.com/go-corelibs/x-text/language"
	"github.com/go-enjin/be/pkg/dbi"
	"github.com/go-enjin/be/pkg/feature"
	site_including "github.com/go-enjin/be/pkg/feature/site-including"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/types/page"
)

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

const Tag feature.Tag = "srv-eql"

type Feature interface {
	feature.Feature
	feature.PageProvider
	feature.PageIndexFeature
	feature.PageContextProvider
	feature.QueryIndexFeature
}

type MakeFeature interface {
	site_including.MakeFeature[MakeFeature]

	// SetDatabase overrides the default internal Sqlite database with the
	// db feature connection given
	SetDatabase(connection feature.Tag) MakeFeature

	// IncludeContextKeys appends to the list of page context keys to index,
	// which by default has only the BaseIncludeContextKeys
	IncludeContextKeys(keys ...string) MakeFeature

	// SetIncludedContextKeys overwrites the list of context keys to index, use
	// this instead of IncludeContextKeys when needing to remove one or more of
	// the default list of inclusions
	SetIncludedContextKeys(keys ...string) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature
	site_including.CSiteIncluding[feature.QueryIndexSourceFeature, MakeFeature]

	// dsn is the connection string for the internal Sqlite database used when
	// no db feature has been configured
	dsn string

	dbTag   feature.Tag
	db      *sql.DB
	dialect sqlbuilder.Dialect

	eql enjinql.EnjinQL

	excludeContextKeys []string
	includeContextKeys []string
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

func (f *CFeature) UsageNotes() (lines []string) {
	if f.dbTag.IsNil() {
		category := f.Tag().String()
		patternKey := globals.MakeFlagEnvKey(category, "INTERNAL_DSN")
		lines = []string{
			f.Tag().String() + " is using an internal Sqlite database, to configure",
			"this connection, use the following environment variable:",
			patternKey + "=\"<dsn>\"",
		}
	}
	return
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.CSiteIncluding.InitSiteIncluding(f)
	f.includeContextKeys = BaseIncludeContextKeys()
	f.dsn = ":memory:"
}

func (f *CFeature) SetDatabase(connection feature.Tag) MakeFeature {
	f.dbTag = connection
	return f
}

func (f *CFeature) IncludeContextKeys(keys ...string) MakeFeature {
	for _, key := range keys {
		kebab := strcase.ToKebab(key)
		if !slices.Within(kebab, f.includeContextKeys) {
			f.includeContextKeys = append(f.includeContextKeys, kebab)
		}
	}
	return f
}

func (f *CFeature) SetIncludedContextKeys(keys ...string) MakeFeature {
	f.includeContextKeys = nil
	f.IncludeContextKeys(keys...)
	return f
}

func (f *CFeature) Make() Feature {
	// prepare a unique list of include keys
	f.includeContextKeys = slices.Unique(append(
		AlwaysIncludeContextKeys,
		f.includeContextKeys...,
	))
	// prune the must-exclude and required keys
	f.includeContextKeys = slices.Prune(
		f.includeContextKeys,
		append(
			RequiredContextKeys(),
			append(
				f.excludeContextKeys,
				MustExcludeContextKeys()...,
			)...,
		)...,
	)
	// convert them all to CamelCase
	f.includeContextKeys = slices.ToCamels(f.includeContextKeys)
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}
	f.CSiteIncluding.BuildSiteIncluding(b)
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	f.CSiteIncluding.StartupSiteIncluding(f.Enjin)

	var writeConfig string

	if f.dbTag.IsNil() {

		category := f.Tag().String()
		internalDSN := globals.MakeFlagEnvKey(category, "INTERNAL_DSN")
		dsn := env.String(internalDSN, "")

		dsnConn, dsnParams, _ := strings.Cut(dsn, "?")
		dsnFile := strings.TrimPrefix(dsnConn, "file:")

		if dsnFile == "" || dsnFile == ":memory:" || strings.Contains(dsnParams, "mode=memory") {
			// default in-memory requires a distinct db name
			// if devs use the default eql.Tag, no longer distinct, fix is to
			// include the enjin SiteTag as well, which in a multi-enjin context
			// must be distinct which makes SiteTag+f.Tag convenient for devs
			dsn = "file:" + strings.ToLower(f.Enjin.SiteTag()) + "-" + f.Tag().Kebab() + ".db?mode=memory&cache=shared"
		} else {
			// determine a suitable filename for the enjinql config
			// prune .db and append .eql
			writeConfig = strings.TrimSuffix(dsnFile, ".db") + ".eql"
		}

		f.dialect = dialects.Sqlite{}
		if dsn, err = dbi.UpdateURI(dsn, map[string]string{
			"charset":   "utf8",
			"parseTime": "true",
			"loc":       "UTC",
		}); err != nil {
			err = fmt.Errorf("internal dsn update error: %q - %w", dsn, err)
			return
		}
		if f.db, err = sql.Open("sqlite3", dsn); err != nil {
			err = fmt.Errorf("internal database connection error: %q - %w", dsn, err)
			return
		}

	} else {
		var db feature.DataBase
		if db, err = f.Enjin.DB(f.dbTag.String()); err != nil {
			err = fmt.Errorf("%q database connection not found", f.dbTag)
			return
		}
		f.db = db.SqlDB()
		f.dialect = db.Build().Dialect()
	}

	config := enjinql.NewConfig(strcase.ToSnake(f.Enjin.SiteTag()), f.FeatureTag.Snake())
	config.AddSource(enjinql.PageSourceConfig())
	config.AddSource(enjinql.PageRedirectSourceConfig())

	// TODO: refactor context.Fields to support concrete typing of indexed context keys
	//       - or - refactor eql to be configured with enjinql.SourceConfig instances
	for _, key := range f.includeContextKeys {
		snake := strcase.ToSnake(key)
		config.AddSource(enjinql.MakeSourceConfig(
			enjinql.PageSource, snake,
			enjinql.NewStringValue("text", -1),
		).AddIndex("text"))
	}

	// add all other sources
	for _, other := range f.CSiteIncluding.Features {
		for _, source := range other.AddSources() {
			config.AddSource(source)
		}
	}

	if f.eql, err = enjinql.New(config, f.db, f.dialect, enjinql.SkipCreateIndex); err != nil {
		err = fmt.Errorf("error constructing new enjinql instance: %w", err)
		return
	}

	data, _ := f.eql.Marshal()
	log.DebugF("%v feature enjinql config: %v", f.Tag(), string(data))

	if writeConfig != "" {
		if err = os.WriteFile(writeConfig, data, 0640); err != nil {
			err = fmt.Errorf("error writing enjinql config: %q - %w", writeConfig, err)
			return
		}
		log.DebugF("%v feature enjinql config written to: %q", writeConfig)
	}
	return
}

func (f *CFeature) PostStartup(ctx *cli.Context) (err error) {
	start := time.Now()
	if err = f.eql.CreateIndexes(); err != nil {
		return
	}
	log.InfoF("%v created indexes in: %v", f.Tag(), time.Now().Sub(start))
	return
}

func (f *CFeature) EQL() enjinql.EnjinQL {
	return f.eql
}

func (f *CFeature) FindPageID(shasum string) (id int64, ok bool) {
	if _, results, err := f.eql.Perform(`LOOKUP .ID WITHIN .Shasum == %q`, shasum); err == nil && len(results) > 0 {
		id, ok = results[0]["id"].(int64)
	}
	return
}

func (f *CFeature) FindPageStub(shasum string) (stub *feature.PageStub) {
	if _, results, err := f.eql.Perform(`QUERY WITHIN .Shasum == %q`, shasum); err == nil && len(results) > 0 {
		if data, ok := results[0]["stub"].(string); ok {
			if stub, err = feature.MakePageStub([]byte(data)); err != nil {
				log.ErrorF("error making page stub: %q - %q", data)
			}
		}
	}
	return
}

func (f *CFeature) PerformQuery(format string, argv ...interface{}) (stubs feature.PageStubs, err error) {
	var results context.Contexts
	if _, results, err = f.eql.Perform(format, argv...); err == nil {
		for _, result := range results {
			if data, ok := result["stub"].(string); ok {
				var stub *feature.PageStub
				if stub, err = feature.MakePageStub([]byte(data)); err != nil {
					log.ErrorF("error making page stub: %q - %q", data, err)
					return
				}
				stubs = append(stubs, stub)
			}
		}
	}
	return
}

func (f *CFeature) PerformLookup(format string, argv ...interface{}) (columns []string, results context.Contexts, err error) {
	columns, results, err = f.eql.Perform(format, argv...)
	return
}

func (f *CFeature) FindRedirection(url string) (p feature.Page) {

	if stubs, err := f.PerformQuery(`QUERY WITHIN %[1]s.Url == %q`, enjinql.PageRedirectSource, url); err == nil && len(stubs) > 0 {
		t := f.Enjin.MustGetTheme()
		ctx := f.Enjin.Context(nil)
		if pg, ee := page.NewPageFromStub(stubs[0], t, ctx); ee == nil {
			p = pg
		}

	}

	return
}

func (f *CFeature) FindTranslations(url string) (pages []feature.Page) {

	url = clPath.CleanWithSlash(url)

	if stubs, err := f.PerformQuery(`QUERY WITHIN (.Url == %q)`, url); err == nil && len(stubs) > 0 {
		t := f.Enjin.MustGetTheme()
		ctx := f.Enjin.Context(nil)
		for _, stub := range stubs {
			if pg, ee := page.NewPageFromStub(stub, t, ctx); ee == nil {
				pages = append(pages, pg)
			}
		}
	}

	return
}

func (f *CFeature) FindTranslationUrls(url string) (pages map[language.Tag]string) {

	pages = make(map[language.Tag]string)

	for _, p := range f.FindTranslations(url) {
		pages[p.LanguageTag()] = p.Url()
	}

	return
}

func (f *CFeature) FindPageStubFrom(tag language.Tag, path string) (p *feature.PageStub) {
	if stubs, err := f.PerformQuery(`QUERY WITHIN (.Language == %q) AND (.Url == %q)`, tag.String(), path); err == nil && len(stubs) > 0 {
		p = stubs[0]
	}
	return
}

func (f *CFeature) FindPage(r *http.Request, tag language.Tag, path string) (p feature.Page) {

	var stub *feature.PageStub

	if stub = f.FindPageStubFrom(tag, path); stub == nil && tag != language.Und {
		stub = f.FindPageStubFrom(language.Und, path)
	}

	// check for exact match
	if stub != nil {
		if pg, err := page.NewPageFromStub(stub, f.Enjin.MustGetTheme(), f.Enjin.Context(r)); err != nil {
			log.ErrorF("error making new page from stub: [%v] %q - %v", tag, path, err)
		} else {
			p = pg
		}
	}

	return
}

func (f *CFeature) LookupPrefixed(prefix string) (pages []feature.Page) {

	if stubs, err := f.PerformQuery(`QUERY WITHIN .Url ^= %q`, prefix); err == nil && len(stubs) > 0 {
		t := f.Enjin.MustGetTheme()
		ctx := f.Enjin.Context(nil)
		for _, stub := range stubs {
			if pg, ee := page.NewPageFromStub(stub, t, ctx); ee == nil {
				pages = append(pages, pg)
			}
		}
	}

	return
}
