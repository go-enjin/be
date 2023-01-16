//go:build fastcgi || all

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

package be

import (
	"fmt"

	"github.com/iancoleman/strcase"
)

func (fe *FastcgiEnjin) String() string {
	return strcase.ToKebab(fe.feb.tag)
}

func (fe *FastcgiEnjin) StartupString() string {
	return fmt.Sprintf(
		`{
	listen: "%v",
	port: %v,
	target: %v,
	source: %v,
	network: %v,
	debug: %v,
	prefix: "%v",
	domains: %v
}`,
		fe.listen,
		fe.port,
		fe.target,
		fe.source,
		fe.network,
		fe.debug,
		fe.prefix,
		fe.feb.domains,
	)
}

func (fe *FastcgiEnjin) Run(argv []string) (err error) {
	if fe.cli == nil {
		err = fmt.Errorf("calling .Run on included enjin")
		return
	}
	return fe.cli.Run(argv)
}