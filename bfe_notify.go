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

	"github.com/go-enjin/be/pkg/globals"
)

func (fe *FastcgiEnjin) Notify(tag string) {
	fe.NotifyF(tag, globals.BuildInfoString())
}

func (fe *FastcgiEnjin) NotifyF(tag, format string, argv ...interface{}) {
	msgFormat := fmt.Sprintf("%s *%s*: %s", globals.BinName, tag, format)
	for _, hook := range _notifyHooks {
		hook(msgFormat, argv...)
	}
}