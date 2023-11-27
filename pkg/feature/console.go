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

package feature

import (
	"fmt"
	"sync"

	"github.com/go-curses/cdk"
	"github.com/go-curses/ctk"
	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/log"
)

var (
	_ Console     = (*CConsole)(nil)
	_ MakeConsole = (*CConsole)(nil)
)

type Console interface {
	Tag() (tag Tag)
	Init(this interface{})
	This() (this interface{})
	Self() (self Console)
	Title() (title string)
	Build(c Buildable) (err error)
	Depends() (deps Tags)
	Setup(ctx *cli.Context, ei Internals)
	Prepare(app ctk.Application)
	Startup(display cdk.Display)
	Shutdown()
	Resized(w, h int)
	App() (app ctk.Application)
	Display() (display cdk.Display)
	Window() (window ctk.Window)
}

type MakeConsole interface {
	Make() Console
}

type CConsole struct {
	PackageTag Tag
	ConsoleTag Tag
	Enjin      Internals

	this interface{}

	app     ctk.Application
	display cdk.Display
	window  ctk.Window

	sync.RWMutex
}

func (c *CConsole) Make() Console {
	if c.ConsoleTag == "" {
		c.ConsoleTag = NotImplemented
	}
	log.DebugF("making console: %v", c.Self().Tag())
	return c.Self()
}

func (c *CConsole) Tag() (tag Tag) {
	if c.ConsoleTag == NotImplemented {
		panic(fmt.Errorf("%T console feature tag not implemented", c))
	}
	return c.ConsoleTag
}

func (c *CConsole) BaseTag() (tag Tag) {
	if c.PackageTag == NotImplemented {
		panic(fmt.Errorf("%T console package tag not implemented", c))
	}
	return c.ConsoleTag
}

func (c *CConsole) Init(this interface{}) {
	c.this = this
}

func (c *CConsole) This() (this interface{}) {
	return c.this
}

func (c *CConsole) Self() (self Console) {
	var ok bool
	if self, ok = c.this.(Console); !ok {
		log.FatalF("internal error - feature not a console: %T %+v", c.this, c.this)
	}
	return
}

func (c *CConsole) Title() (title string) {
	return strcase.ToDelimited(c.Self().Tag().String(), ' ')
}

func (c *CConsole) Build(b Buildable) (err error) {
	return
}

func (c *CConsole) Depends() (deps Tags) {
	return
}

func (c *CConsole) Setup(ctx *cli.Context, ei Internals) {
	c.Enjin = ei
}

func (c *CConsole) Prepare(app ctk.Application) {
	c.app = app
}

func (c *CConsole) Startup(display cdk.Display) {
	c.display = display

	c.window = ctk.NewWindow()
	c.window.SetName(fmt.Sprintf("%v-window", c.Tag().Kebab()))
	c.window.SetTitle(c.Self().Title())
}

func (c *CConsole) Shutdown() {
}

func (c *CConsole) Resized(w, h int) {
}

func (c *CConsole) App() (app ctk.Application) {
	if c.app == nil {
		log.FatalDF(1, "calling %v.App() before Prepare() happened", c.Self().Tag())
	}
	return c.app
}

func (c *CConsole) Display() (display cdk.Display) {
	if c.display == nil {
		log.FatalDF(1, "calling %v.Display() before Startup() happened", c.Self().Tag())
	}
	return c.display
}

func (c *CConsole) Window() (window ctk.Window) {
	if c.window == nil {
		log.FatalDF(1, "calling %v.Window() before Startup() happened", c.Self().Tag())
	}
	return c.window
}
