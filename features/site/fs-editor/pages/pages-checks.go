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

package pages

import (
	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/types/page"
	"github.com/go-enjin/be/types/page/matter"
)

func (f *CFeature) PageRenderCheck(p feature.Page) (err error) {
	renderer := f.Enjin.GetThemeRenderer(p.Context())
	_, _, err = renderer.PrepareRenderPage(f.Enjin.MustGetTheme(), f.Enjin.Context(nil), p)
	return
}

func (f *CFeature) InfoRenderCheck(info *editor.File) (p feature.Page, pm *matter.PageMatter, err error) {
	if info.HasDraft {
		if pm, err = f.ReadDraftMatter(info); err != nil {
			return
		}
	} else {
		if pm, err = f.ReadPageMatter(info); err != nil {
			return
		}
	}
	if p, err = page.NewFromPageMatter(pm.Copy(), f.Enjin.MustGetTheme(), f.Enjin.Context(nil)); err != nil {
		return
	}
	err = f.PageRenderCheck(p)
	return
}