//go:build all || curses

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
	"path/filepath"
	"sort"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-curses/ctk"

	"github.com/go-curses/cdk"
	cenums "github.com/go-curses/cdk/lib/enums"
	cstrings "github.com/go-curses/cdk/lib/strings"
	"github.com/go-curses/cdk/log"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/signals"
)

var (
	ctkApp        ctk.Application
	ctkConsoleTag feature.Tag
)

// Build Configuration Flags
// setting these will enable command line flags and their corresponding features
// use `go build -v -ldflags="-X 'github.com/go-enjin/be.CdkIncludeLogFullPaths=false'"`
var (
	CdkIncludeProfiling     = "false"
	CdkIncludeLogFile       = "false"
	CdkIncludeLogFormat     = "false"
	CdkIncludeLogFullPaths  = "false"
	CdkIncludeLogLevel      = "false"
	CdkIncludeLogLevels     = "false"
	CdkIncludeLogTimestamps = "false"
	CdkIncludeLogOutput     = "false"
)

var (
	DefaultConsoleTtyPath = "/dev/tty"
)

func init() {
	cdk.Build.Profiling = cstrings.IsTrue(CdkIncludeProfiling)
	cdk.Build.LogFile = cstrings.IsTrue(CdkIncludeLogFile)
	cdk.Build.LogFormat = cstrings.IsTrue(CdkIncludeLogFormat)
	cdk.Build.LogFullPaths = cstrings.IsTrue(CdkIncludeLogFullPaths)
	cdk.Build.LogLevel = cstrings.IsTrue(CdkIncludeLogLevel)
	cdk.Build.LogLevels = cstrings.IsTrue(CdkIncludeLogLevels)
	cdk.Build.LogTimestamps = cstrings.IsTrue(CdkIncludeLogTimestamps)
	cdk.Build.LogOutput = cstrings.IsTrue(CdkIncludeLogOutput)
	if cdk.Build.LogFile {
		log.DefaultLogPath = filepath.Join(os.TempDir(), globals.BinName+"--console.cdk.log")
	} else {
		log.DefaultLogPath = "/dev/null"
	}
}

func (e *Enjin) initConsoles() {
	e.Emit(signals.PreInitConsoleFeatures, feature.EnjinTag.String(), interface{}(e).(feature.Internals))

	numConsoles := len(e.eb.consoles)
	if numConsoles > 0 {
		var included []string
		for tag, _ := range e.eb.consoles {
			included = append(included, tag.String())
		}
		sort.Strings(included)
		var includedText string
		for _, tag := range included {
			includedText += "\n  - " + strcase.ToKebab(tag) + " (" + tag + ")"
		}
		e.cli.Commands = append(e.cli.Commands, &cli.Command{
			Name:        "console",
			Usage:       "go-curses console user interface",
			Description: fmt.Sprintf("%d console(s) are included within this build:%v", numConsoles, includedText),
			UsageText:   fmt.Sprintf("%v [global options] console <tag>", globals.BinName),
			Action:      e.consoleAction,
		})
	}

	e.Emit(signals.PostInitConsoleFeatures, feature.EnjinTag.String(), interface{}(e).(feature.Internals))
}

func (e *Enjin) consoleAction(ctx *cli.Context) (err error) {

	argv := ctx.Args().Slice()
	argc := len(argv)
	if argc == 0 {
		return cli.ShowCommandHelp(ctx, "console")
	}

	ctkConsoleTag = feature.Tag(strcase.ToKebab(argv[0]))
	if console, ok := e.eb.consoles[ctkConsoleTag]; !ok {
		fmt.Printf("\n# ERROR: console %v not found\n\n", ctkConsoleTag)
		return cli.ShowCommandHelp(ctx, "console")
	} else {
		console.Setup(ctx, e)
	}

	if err = e.startupFeatures(ctx); err != nil {
		return
	}

	if err = e.startupIntegrityChecks(ctx); err != nil {
		return
	}

	ctkApp = ctk.NewApplication(
		globals.BinName, globals.BinName, globals.Summary,
		globals.Version, globals.BinName, globals.BinName,
		DefaultConsoleTtyPath,
	)

	ctkApp.Connect(cdk.SignalPrepare, "enjin-consoles-prepare-handler", e.ctkPrepare)
	ctkApp.Connect(cdk.SignalStartup, "enjin-consoles-startup-handler", e.ctkStartup)
	ctkApp.Connect(cdk.SignalShutdown, "enjin-consoles-shutdown-handler", e.ctkShutdown)

	return ctkApp.Run([]string{globals.BinName})
}

func (e *Enjin) ctkPrepare(data []interface{}, argv ...interface{}) cenums.EventFlag {
	if len(argv) > 0 {
		if app, ok := argv[0].(ctk.Application); ok {
			if console, ok := e.eb.consoles[ctkConsoleTag]; ok && console != nil {
				console.Prepare(app)
			}
		}
	}
	return cenums.EVENT_PASS
}

func (e *Enjin) ctkStartup(data []interface{}, argv ...interface{}) cenums.EventFlag {
	if _, display, _, _, _, ok := ctk.ArgvApplicationSignalStartup(argv...); ok {

		display.CaptureCtrlC()
		display.Connect(cdk.SignalEventResize, "enjin-consoles-event-resize-handler", e.ctkDisplayResized)

		if console, ok := e.eb.consoles[ctkConsoleTag]; ok && console != nil {
			console.Startup(display)
		}

		return cenums.EVENT_PASS
	}
	log.ErrorF("internal error - unexpected startup argv: %+v", argv)
	return cenums.EVENT_STOP
}

func (e *Enjin) ctkShutdown(data []interface{}, argv ...interface{}) cenums.EventFlag {
	if console, ok := e.eb.consoles[ctkConsoleTag]; ok && console != nil {
		console.Shutdown()
	}
	return cenums.EVENT_PASS
}

func (e *Enjin) ctkDisplayResized(data []interface{}, argv ...interface{}) cenums.EventFlag {
	if len(argv) >= 2 {
		if evt, ok := argv[1].(*cdk.EventResize); ok {
			if console, ok := e.eb.consoles[ctkConsoleTag]; ok && console != nil {
				console.Resized(evt.Size())
			}
		}
	}
	return cenums.EVENT_PASS
}
