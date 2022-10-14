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
)

func (e *Enjin) String() string {
	return fmt.Sprintf(
		`{
	listen: "%v",
	port: %v,
	debug: %v,
	prefix: "%v",
	themes: %v
}`,
		e.listen,
		e.port,
		e.debug,
		e.prefix,
		e.ThemeNames(),
	)
}

func (e *Enjin) Run(argv []string) (err error) {
	e.setupFeatures()
	return e.cli.Run(argv)
}