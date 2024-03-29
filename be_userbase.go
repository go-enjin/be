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
	"github.com/go-enjin/be/pkg/feature"
)

func (e *Enjin) GetPublicAccess() (actions feature.Actions) {
	actions = actions.Append(e.eb.publicUser...)
	return
}

func (e *Enjin) FindAllUserActions() (list feature.Actions) {

	list = append(list, e.eb.userActions...)

	for _, uap := range e.eb.fUserActionsProviders {
		list = append(list, uap.UserActions()...)
	}

	return
}
