//go:build buntdb_indexing || buntdb || all

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

package indexing

import (
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

func (f *CFeature) buntdbKwsCommandAction(ctx *cli.Context) (err error) {
	if v := ctx.Path("buntdb-path"); v == "" {
		cli.ShowCommandHelpAndExit(ctx, "buntdb-precache", 1)
	}
	if ri, ok := f.enjin.Self().(feature.RootInternals); ok {
		if err = ri.SetupRootEnjin(ctx); err != nil {
			return
		}
	}

	f.cliStartup = true

	if err = f.Startup(ctx); err != nil {
		return
	}

	// if err = f.buntdb.SetConfig(buntdb.Config{AutoShrinkDisabled: true, SyncPolicy: buntdb.Always}); err != nil {
	// 	return
	// }

	for _, feat := range f.enjin.Features() {
		if pp, ok := feat.(feature.PageProvider); ok {
			log.DebugF("starting up page provider: %v", feat.Tag())
			if err = pp.Startup(ctx); err != nil {
				return
			}
		}
	}
	for _, feat := range f.enjin.Features() {
		if pp, ok := feat.(feature.PageProvider); ok {
			log.DebugF("shutting down page provider: %v", feat.Tag())
			pp.Shutdown()
		}
	}
	f.Shutdown()
	return
}