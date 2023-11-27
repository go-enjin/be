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

package locales

//type LocaleMessages struct {
//	// Lookup is a mapping of shasum IDs to specific locale messages
//	Lookup map[string]*LocaleMessage `json:"lookup"`
//	// Order is the list of message Shasums as contained in the source file
//	Order []string `json:"order"`
//}
//
//func (l *LocaleMessages) Append(lm *LocaleMessage) {
//	if l.Lookup == nil {
//		l.Lookup = make(map[string]*LocaleMessage)
//	}
//	if _, present := l.Lookup[lm.Shasum]; !present {
//		l.Order = append(l.Order, lm.Shasum)
//		l.Lookup[lm.Shasum] = lm
//	}
//}
