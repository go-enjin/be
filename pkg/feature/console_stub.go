//go:build !all && !curses

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

	"github.com/go-enjin/be/pkg/log"
)

type Console interface {
	Tag() (tag Tag)
	Init(this interface{})
	This() (this interface{})
	Self() (self Console)
	Build(c Buildable) (err error)
	Depends() (deps Tags)
}

type MakeConsole interface {
	Make() Console
}

type CConsole struct {
	this interface{}

	sync.RWMutex
}

func (c *CConsole) Make() Console {
	log.ErrorF("enjin console support not included in this build")
	return nil
}

func (c *CConsole) Tag() (tag Tag) {
	return NotImplemented
}

func (c *CConsole) Init(this interface{}) {
	c.this = this
}

func (c *CConsole) This() (this interface{}) {
	return c.this
}

func (c *CConsole) Self() (self Console) {
	log.ErrorF("enjin console support not included in this build")
	return
}

func (c *CConsole) Build(b Buildable) (err error) {
	return fmt.Errorf("enjin console support not included in this build")
}

func (c *CConsole) Depends() (deps Tags) {
	return
}
