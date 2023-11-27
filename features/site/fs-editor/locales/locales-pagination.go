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

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/editor"
)

func (f *CFeature) makePaginationRedirect(info *editor.File, form context.Context) (redirect string) {
	var searchQuery string
	var numPerPage, pageIndex = -1, -1
	if v, ok := form["search.query"].(string); ok {
		searchQuery = strings.TrimSpace(v)
	}
	if v, ok := form["pagination"].(map[string]interface{}); ok {
		if vv, ok := v["num-per-page"].(string); ok {
			numPerPage, _ = strconv.Atoi(vv)
		}
		if vv, ok := v["page-index"].(string); ok {
			pageIndex, _ = strconv.Atoi(vv)
		}
	}
	if numPerPage > -1 && pageIndex > -1 {
		if searchQuery != "" {
			redirect = fmt.Sprintf(`%s/%s/:%s/%d/%d#search`, f.GetEditorPath(), info.EditCodeFilePath(), url.PathEscape(searchQuery), numPerPage, pageIndex)
		} else {
			redirect = fmt.Sprintf(`%s/%s/%d/%d`, f.GetEditorPath(), info.EditCodeFilePath(), numPerPage, pageIndex)
		}
	} else if searchQuery != "" {
		redirect = fmt.Sprintf(`%s/%s/:%s#search`, f.GetEditorPath(), info.EditCodeFilePath(), url.PathEscape(searchQuery))
	} else {
		redirect = fmt.Sprintf(`%s/%s`, f.GetEditorPath(), info.EditCodeFilePath())
	}
	return
}
