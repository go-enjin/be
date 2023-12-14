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

package be

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/hostrouter"
	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/profiling"
	"github.com/go-enjin/golang-org-x-text/language"

	"github.com/go-enjin/be/pkg/factories/nonces"
	"github.com/go-enjin/be/pkg/factories/tokens"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/lang/catalog"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
	"github.com/go-enjin/be/pkg/net/headers/policy/permissions"
	"github.com/go-enjin/be/pkg/signals"
	"github.com/go-enjin/be/pkg/slices"
)

var (
	_ feature.Runner    = (*Enjin)(nil)
	_ feature.System    = (*Enjin)(nil)
	_ feature.Service   = (*Enjin)(nil)
	_ feature.Internals = (*Enjin)(nil)
)

func init() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s %s\n", globals.BinName, c.App.Version)
	}
}

type Enjin struct {
	prefix     string
	production bool

	debug bool

	catalog catalog.Catalog
	locales []language.Tag

	contentSecurityPolicy *csp.PolicyHandler
	permissionsPolicy     *permissions.PolicyHandler

	eb  *EnjinBuilder
	cli *cli.App

	router *chi.Mux
	enjins []*Enjin

	nonces  feature.NonceFactory
	tokens  feature.TokenFactory
	lockers feature.SyncLockerFactory

	mutex *sync.RWMutex
}

func newEnjin(eb *EnjinBuilder) *Enjin {
	var description string
	notes := make(map[feature.Tag][]string)
	for _, f := range eb.features.List() {
		if un := f.UsageNotes(); len(un) > 0 {
			notes[f.Tag()] = un
		}
	}
	for _, ie := range eb.enjins {
		for _, f := range ie.features.List() {
			if un := f.UsageNotes(); len(un) > 0 {
				notes[f.Tag()] = un
			}
		}
	}
	if len(notes) > 0 {
		description += "Feature usage notes:\n\n"
		for _, tag := range feature.SortedFeatureTags(notes) {
			description += " * " + tag.String() + ":\n"
			for _, note := range notes[tag] {
				description += "   * " + strings.TrimSpace(note) + "\n"
			}
		}
	}
	e := &Enjin{
		eb: eb,
		cli: &cli.App{
			Name:        globals.BinName,
			Usage:       globals.Summary,
			Version:     globals.BuildVersion(),
			Description: description,
			Flags:       eb.flags,
			Commands:    eb.commands,
		},
		// policies
		contentSecurityPolicy: csp.NewPolicyHandler(),
		permissionsPolicy:     permissions.NewPolicyHandler(),
		// make variables
		catalog: catalog.New(),
		mutex:   &sync.RWMutex{},
	}
	e.cli.Action = e.action

	e.Emit(signals.PreNewEnjin, feature.EnjinTag.String(), interface{}(e).(feature.Internals))

	e.initConsoles()
	e.setupFeatures()
	e.ReloadLocales()

	e.Emit(signals.PostNewEnjin, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
	return e
}

func newIncludedEnjin(eb *EnjinBuilder, parent *Enjin) *Enjin {
	e := &Enjin{
		eb:                    eb,
		contentSecurityPolicy: csp.NewPolicyHandler(),
		permissionsPolicy:     permissions.NewPolicyHandler(),
		catalog:               catalog.New(),
		mutex:                 &sync.RWMutex{},
	}
	e.Emit(signals.PreNewEnjinIncluded, feature.EnjinTag.String(),
		interface{}(e).(feature.Internals),
		interface{}(parent).(feature.Internals),
	)
	e.initConsoles()
	e.setupFeatures()
	e.ReloadLocales()
	e.Emit(signals.PostNewEnjinIncluded, feature.EnjinTag.String(),
		interface{}(e).(feature.Internals),
		interface{}(parent).(feature.Internals),
	)
	return e
}

func (e *Enjin) action(ctx *cli.Context) (err error) {

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint
		for _, enjin := range e.eb.enjins {
			for _, f := range enjin.Features().List() {
				f.Shutdown()
			}
		}
		e.Shutdown()
	}()

	if err = e.SetupRootEnjin(ctx); err != nil {
		return
	}

	if err = e.startupIntegrityChecks(ctx); err != nil {
		return
	}

	if err = e.startupFeatures(ctx); err != nil {
		return
	}

	if err = e.startupRootService(ctx); err != nil {
		if err.Error() == "http: Server closed" || strings.Contains(err.Error(), "Listener closed") {
			err = nil
		}
	}
	return
}

func (e *Enjin) Self() (self interface{}) {
	return e
}

func (e *Enjin) setupInternals(ctx *cli.Context) (err error) {

	if len(e.eb.theming) == 0 {
		err = fmt.Errorf("builder error: at least one theme is required")
		return
	} else if e.eb.fServiceListener == nil {
		err = fmt.Errorf("builder error: a feature.ServiceListener is required")
		return
	} else if e.eb.fPanicHandler == nil {
		err = fmt.Errorf("builder error: a feature.PanicHandler is required")
		return
	} else if e.eb.fLocaleHandler == nil {
		err = fmt.Errorf("builder error: a feature.LocaleHandler is required")
		return
	} else if e.eb.fSyncLockerFactory == nil {
		err = fmt.Errorf("builder error: a feature.SyncLockerFactoryFeature is required")
		return
	}

	if e.eb.fNonceFactory == nil {
		e.nonces = nonces.New(-1)
		log.DebugF("feature.NonceFactoryFeature not found, falling back to built-in factory (this enjin will not scale)")
	}
	if e.eb.fTokenFactory == nil {
		e.tokens = tokens.New(-1)
		log.DebugF("feature.TokenFactoryFeature not found, falling back to built-in factory (this enjin will not scale)")
	}

	if e.eb.fServiceLogHandler == nil {
		err = fmt.Errorf("builder error: a feature.ServiceLogHandler is required")
		return
	} else if list := feature.FilterTyped[feature.ServiceLogger](e.eb.features.List()); len(list) == 0 {
		err = fmt.Errorf("builder error: at least one feature.ServiceLogger is required")
		return
	} else {
		middleware.DefaultLogger = e.eb.fServiceLogHandler.LogHandler
	}

	return
}

func (e *Enjin) SetupRootEnjin(ctx *cli.Context) (err error) {

	e.debug = ctx.Bool("debug")
	e.prefix = ctx.String("prefix")
	e.prefix = strings.ToLower(e.prefix)
	e.production = e.prefix == "" || e.prefix == "prd"
	if e.production {
		e.prefix = ""
		e.debug = false
	}

	lvl := strings.ToLower(ctx.String("log-level"))
	if v, ok := log.Levels[lvl]; ok {
		log.Config.LogLevel = v
		log.Config.Apply()
	} else {
		if lvl != "" {
			log.FatalF("invalid log-level: %v", lvl)
		}
		if e.debug {
			log.Config.LogLevel = log.LevelDebug
			log.Config.Apply()
		} else if ctx.Bool("quiet") {
			log.Config.LogLevel = log.LevelWarn
			log.Config.Apply()
		}
	}

	if err = e.setupInternals(ctx); err != nil {
		return
	}

	if domains := ctx.StringSlice("domain"); domains != nil && len(domains) > 0 {
		for _, domain := range domains {
			if domain != "" && !slices.Present(domain, e.eb.domains...) {
				e.eb.domains = append(e.eb.domains, domain)
			}
		}
	}

	for _, enjin := range e.eb.enjins {
		tag := strcase.ToKebab(enjin.tag)
		if domains := ctx.StringSlice(tag + "-domain"); len(domains) > 0 {
			for _, domain := range domains {
				if !slices.Within(domain, enjin.domains) {
					log.DebugF("adding domain to %v enjin: %v", tag, domain)
					enjin.domains = append(enjin.domains, domain)
				}
			}
		}
	}

	e.Emit(signals.RootEnjinSetup, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
	return
}

func (e *Enjin) setupFeatures() {
	e.Emit(signals.PreSetupFeaturesPhase, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
	for _, f := range e.eb.features.List() {
		f.Setup(e)
	}
	e.Emit(signals.PostSetupFeaturesPhase, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
}

func (e *Enjin) startupFeatures(ctx *cli.Context) (err error) {
	e.Emit(signals.PreStartupFeaturesPhase, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
	for _, f := range e.eb.features.List() {
		if err = f.Startup(ctx); err != nil {
			err = fmt.Errorf("error starting up %q feature: %v", f.Tag(), err)
			return
		}
	}
	for _, f := range feature.FilterTyped[feature.PostStartupFeature](e.eb.features.List()) {
		if err = f.PostStartup(ctx); err != nil {
			err = fmt.Errorf("error running post-startup for %q feature: %v", f.Tag(), err)
			return
		}
	}
	e.Emit(signals.PostStartupFeaturesPhase, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
	return
}

func (e *Enjin) startupRootService(ctx *cli.Context) (err error) {

	e.router = chi.NewRouter()
	if err = e.setupRouter(e.router); err != nil {
		return
	}

	go func() {
		sighup := make(chan os.Signal)
		signal.Notify(sighup, syscall.SIGHUP)
		for {
			select {
			case <-sighup:
				e.Emit(signals.RootEnjinPreReload, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
				for _, enjin := range e.eb.enjins {
					for _, f := range enjin.Features().List() {
						if rf, ok := f.This().(feature.ReloadableFeature); ok {
							rf.Reload()
						}
					}
				}
				for _, f := range e.Features().List() {
					if rf, ok := f.This().(feature.ReloadableFeature); ok {
						rf.Reload()
					}
				}
				e.Emit(signals.RootEnjinPostReload, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
			}
		}
	}()

	if len(e.eb.enjins) == 0 {
		return e.eb.fServiceListener.StartListening(e.router, e)
	}

	hr := hostrouter.New()

	for _, eb := range e.eb.enjins {
		if len(eb.domains) == 0 {
			log.FatalDF(5, "%v enjin domains not found", eb.tag)
			continue
		}

		enjin := newIncludedEnjin(eb, e)
		if err = enjin.setupInternals(ctx); err != nil {
			log.FatalDF(5, "%v enjin setup error: %v", err)
			continue
		}

		e.enjins = append(e.enjins, enjin)

		if err = enjin.startupFeatures(ctx); err != nil {
			return
		}

		enjin.router = chi.NewRouter()
		if err = enjin.setupRouter(enjin.router); err != nil {
			return
		}

		for _, domain := range eb.domains {
			hr.Map(domain, enjin.router)
		}
	}

	hr.Map("*", e.router)

	root := chi.NewRouter()
	root.Mount("/", hr)
	e.Emit(signals.RootEnjinStarting, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
	return e.eb.fServiceListener.StartListening(root, e)
}

func (e *Enjin) Shutdown() {
	e.Emit(signals.PreShutdownFeaturesPhase, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
	for _, f := range e.eb.features.List() {
		f.Shutdown()
	}
	e.Emit(signals.PostShutdownFeaturesPhase, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
	if err := e.eb.fServiceListener.StopListening(); err != nil {
		log.ErrorF("error stopping http listener: %v - %v", e.eb.fServiceListener.Tag(), err)
	}
	e.Emit(signals.RootEnjinShutdown, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
	e.Notify("enjin shutdown complete")
	profiling.Stop()
	os.Exit(0)
}