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

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/fs"
	"github.com/go-enjin/be/pkg/globals"
)

func (eb *EnjinBuilder) MakeEnvKey(name string) (key string) {
	key = globals.MakeEnvKey(name)
	return
}

func (eb *EnjinBuilder) MakeEnvKeys(names ...string) (keys []string) {
	keys = globals.MakeEnvKeys(names...)
	return
}

func (eb *EnjinBuilder) RegisterPublicFileSystem(mount string, filesystems ...fs.FileSystem) {
	eb.publicFileSystems.Register(mount, filesystems...)
	return
}

func (eb *EnjinBuilder) RegisterTemplatePartial(block, position, name, tmpl string) (err error) {
	if len(eb.fTemplatePartialsProvider) == 0 {
		err = fmt.Errorf("missing page partials feature")
		return
	}
	err = eb.fTemplatePartialsProvider[0].RegisterTemplatePartial(block, position, name, tmpl)
	return
}

func (eb *EnjinBuilder) Features() (cache *feature.FeaturesCache) {
	cache = eb.features
	return
}
