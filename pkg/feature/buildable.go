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
	"github.com/go-enjin/be/pkg/feature/signaling"
	"github.com/go-enjin/be/pkg/fs"
)

type Buildable interface {
	Builder
	signaling.SignalSupport

	// MakeEnvKey returns name with EnvPrefix (SCREAMING_SNAKE formatted)
	MakeEnvKey(name string) (key string)

	// MakeEnvKeys returns name with EnvPrefix (SCREAMING_SNAKE formatted)
	MakeEnvKeys(names ...string) (key []string)

	// RegisterPublicFileSystem mounts the given static FileSystems
	RegisterPublicFileSystem(mount string, filesystems ...fs.FileSystem)

	// RegisterTemplatePartial sets the named go-template content for inclusion at the specified block and position
	// Notes:
	//    * "block" must be one of "head" or "body"
	//    * position must be one of "head" or "tail"
	//    * adds the given name tmpl to the first feature.TemplatePartialsProvider
	//    * auto-adds a feature.TemplatePartialsProvider if none are present
	RegisterTemplatePartial(block, position, name, tmpl string) (err error)

	// Features provides access to the cache of enjin features
	Features() (cache *FeaturesCache)
}