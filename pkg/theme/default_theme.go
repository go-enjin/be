//go:build all || !excludeDefaultTheme

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

package theme

import (
	"embed"

	beFsEmbed "github.com/go-enjin/be/pkg/fs/embed"
	"github.com/go-enjin/be/pkg/log"
)

//go:embed default/**
//go:embed default/layouts/_default/**
var defaultThemeEmbedFS embed.FS

var defaultThemeInstance *Theme = nil

func DefaultTheme() *Theme {
	if defaultThemeInstance != nil {
		return defaultThemeInstance
	}
	dt := new(Theme)
	dt.Path = "default"
	dt.Name = "default"
	var err error
	if dt.FileSystem, err = beFsEmbed.New(dt.Path, defaultThemeEmbedFS); err != nil {
		log.FatalF("error including defaultThemeFS: %v", err)
	}
	if err = dt.init(); err != nil {
		log.FatalF("error initializing defaultThemeInstance: %v", err)
	}
	log.DebugF("included default theme")
	defaultThemeInstance = dt
	return dt
}