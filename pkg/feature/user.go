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

package feature

import (
	"github.com/go-enjin/be/pkg/context"
)

type User interface {
	GetRID() (rid string)
	GetEID() (eid string)
	GetName() (name string)
	GetEmail() (email string)
	GetImage() (image string)
	GetOrigin() (origin string)
	GetGroups() (groups Groups)
	GetActions() (groups Actions)

	GetActive() (active bool)
	GetAdminLocked() (locked bool)

	GetSetting(key string) (value interface{})
	GetSettings(includeKeys ...string) (settings context.Context)

	UnsafeContext() (ctx context.Context)
	SafeContext(includeKeys ...string) (ctx context.Context)

	Can(action ...Action) (allowed bool)
	CanAll(action ...Action) (allowed bool)

	IsVisitor() (visitor bool)

	Bytes() (data []byte)
}