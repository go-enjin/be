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
	"net/http"

	"github.com/go-enjin/be/pkg/context"
)

type RoutePagesHandler interface {
	Feature
	RoutePage(w http.ResponseWriter, r *http.Request)
}

type ServePagesHandler interface {
	Feature
	ServePage(p Page, t Theme, ctx context.Context, w http.ResponseWriter, r *http.Request) (err error)
}

type PrepareServePagesFeature interface {
	Feature
	PrepareServePage(ctx context.Context, t Theme, p Page, w http.ResponseWriter, r *http.Request) (out context.Context, modified *http.Request, stop bool)
}

type FinalizeServeRequestFeature interface {
	Feature
	FinalizeServeRequest(w http.ResponseWriter, r *http.Request) (modified *http.Request)
}
