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

import "html/template"

type EnjinBlock interface {
	NjnFeature

	NjnBlockType() (name string)

	ProcessBlock(re EnjinRenderer, blockType string, data map[string]interface{}) (html template.HTML, redirect string, err error)

	PrepareBlock(re EnjinRenderer, blockType string, data map[string]interface{}) (block map[string]interface{}, redirect string, err error)
	RenderPreparedBlock(re EnjinRenderer, block map[string]interface{}) (html template.HTML, err error)
}

type CEnjinBlock struct {
	CNjnFeature
}