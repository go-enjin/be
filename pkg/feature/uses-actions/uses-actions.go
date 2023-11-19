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

package uses_actions

import (
	"fmt"

	"github.com/go-enjin/be/pkg/feature"
)

type CUsesActions struct {
	this        interface{}
	_actionsTag string
}

func (c *CUsesActions) ConstructUsesActions(this interface{}) {
	if f, ok := this.(feature.Feature); ok {
		c.this = this
		c._actionsTag = f.Tag().Kebab()
	} else {
		panic(fmt.Sprintf("%T does not implement feature.Feature", this))
	}
}

func (c *CUsesActions) UserActions() (actions feature.Actions) {
	return
}

func (c *CUsesActions) Action(verb string, details ...string) (action feature.Action) {
	action = feature.NewAction(c._actionsTag, verb, details...)
	return
}