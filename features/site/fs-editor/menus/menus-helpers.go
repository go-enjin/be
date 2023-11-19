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

package menus

import (
	"encoding/json"
	"net/http"

	"github.com/go-enjin/be/pkg/editor"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/menu"
	"github.com/go-enjin/be/pkg/userbase"
)

func (f *CFeature) ParseFormToDraft(list []interface{}, info *editor.File, r *http.Request) (parsed menu.EditMenu, redirect string) {
	eid := userbase.GetCurrentEID(r)
	printer := lang.GetPrinterFromRequest(r)
	if data, ee := json.Marshal(list); ee != nil {
		log.ErrorRF(r, "error encoding form context: %v", ee)
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error encoding form context: "%[1]s"`, ee.Error()))
		redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditFilePath()
		return
	} else if ee = json.Unmarshal(data, &parsed); ee != nil {
		log.ErrorRF(r, "error decoding form context: %v", ee)
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error decoding form context: "%[1]s"`, ee.Error()))
		redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditFilePath()
		return
	}
	return
}

func (f *CFeature) ParseDraftToMenu(parsed menu.EditMenu, info *editor.File, r *http.Request) (cleaned menu.Menu, redirect string) {
	eid := userbase.GetCurrentEID(r)
	printer := lang.GetPrinterFromRequest(r)
	if data, ee := json.Marshal(parsed); ee != nil {
		log.ErrorRF(r, "error encoding cleaned menu: %v", ee)
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error encoding cleaned menu: "%v"`, ee))
		redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditFilePath()
	} else if ee = json.Unmarshal(data, &cleaned); ee != nil {
		log.ErrorRF(r, "error decoding cleaned menu: %v", ee)
		f.Editor.Site().PushErrorNotice(eid, true, printer.Sprintf(`error decoding cleaned menu: "%v"`, ee))
		redirect = f.SelfEditor().GetEditorPath() + "/" + info.EditFilePath()
	}
	return
}