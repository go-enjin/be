//go:build all || semanticTheme

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

//go:embed semantic/**
//go:embed semantic/layouts/_default/**
var semanticThemeEmbedFS embed.FS

var semanticThemeInstance *Theme = nil

func SemanticTheme() *Theme {
	if semanticThemeInstance != nil {
		return semanticThemeInstance
	}
	dt := new(Theme)
	dt.Path = "semantic"
	dt.Name = "semantic"
	var err error
	if dt.FileSystem, err = beFsEmbed.New(dt.Path, semanticThemeEmbedFS); err != nil {
		log.FatalF("error including semanticThemeFS: %v", err)
	}
	if err = dt.init(); err != nil {
		log.FatalF("error initializing semanticThemeInstance: %v", err)
	}
	log.DebugF("included semantic theme")
	semanticThemeInstance = dt
	return dt
}