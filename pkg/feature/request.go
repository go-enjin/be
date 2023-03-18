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
	"net/http"

	"github.com/go-enjin/be/pkg/net/headers/policy/csp"
	"github.com/go-enjin/be/pkg/net/headers/policy/permissions"
)

type RequestFilterFn = func(r *http.Request) (err error)

type RequestFilter interface {
	FilterRequest(r *http.Request) (err error)
}

type RequestRewriter interface {
	RewriteRequest(w http.ResponseWriter, r *http.Request) (modified *http.Request)
}

type RequestModifier interface {
	ModifyRequest(w http.ResponseWriter, r *http.Request)
}

type HeadersModifier interface {
	ModifyHeaders(w http.ResponseWriter, r *http.Request)
}

type ContentSecurityPolicyModifier interface {
	ModifyContentSecurityPolicy(policy csp.Policy, r *http.Request) (modified csp.Policy)
}

type PermissionsPolicyModifier interface {
	ModifyPermissionsPolicy(policy permissions.Policy, r *http.Request) (modified permissions.Policy)
}