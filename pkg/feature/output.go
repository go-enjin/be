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

type TranslateOutputFn = func(input []byte) (output []byte, mime string)

type OutputTranslator interface {
	CanTranslate(mime string) (ok bool)
	TranslateOutput(s Service, input []byte, inputMime string) (output []byte, mime string, err error)
}

type TransformOutputFn = func(input []byte) (output []byte)

type OutputTransformer interface {
	CanTransform(mime string) (ok bool)
	TransformOutput(mime string, input []byte) (output []byte)
}