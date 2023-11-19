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

package uses_enjin_salt

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
)

type CUsesEnjinSalt struct {
	this      interface{}
	_category string
	_flagName string
	_salt     string
}

func New(this interface{}) (c *CUsesEnjinSalt) {
	if f, ok := this.(feature.Feature); ok {
		kebab := f.Tag().Kebab()
		c = &CUsesEnjinSalt{
			this:      this,
			_category: kebab,
			_flagName: kebab + "-enjin-salt",
		}
		return
	}
	panic(fmt.Sprintf("%T does not implement feature.Feature", this))
}

func (c *CUsesEnjinSalt) SetEnjinSalt(value string) {
	c._salt = value
	return
}

func (c *CUsesEnjinSalt) GetEnjinSalt() (salt string) {
	salt = c._salt
	return
}

func (c *CUsesEnjinSalt) BuildEnjinSalt(b feature.Buildable) (err error) {
	b.AddFlags(
		&cli.StringFlag{
			Name:     c._flagName,
			Usage:    "specify this feature's enjin salt value",
			EnvVars:  b.MakeEnvKeys(c._flagName),
			Category: c._category,
		},
	)
	return
}

func (c *CUsesEnjinSalt) StartupEnjinSalt(ctx *cli.Context) (err error) {
	if ctx.IsSet(c._flagName) {
		if v := ctx.String(c._flagName); v != "" {
			c._salt = v
		}
	}
	if count := len(c._salt); count == 0 {
		err = fmt.Errorf("--%s is required", c._flagName)
		return
	} else if count < 32 {
		err = fmt.Errorf("--%s needs to be at least 32 characters long", c._flagName)
		return
	}
	return
}