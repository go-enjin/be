// Copyright (c) 2024  The Go-Enjin Authors
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

package permalink

import (
	"fmt"

	"github.com/go-corelibs/enjinql"
	"github.com/go-enjin/be/pkg/feature"
)

func (f *CFeature) AddSources() (sources enjinql.ConfigSources) {
	return enjinql.ConfigSources{enjinql.PagePermalinkSourceConfig()}
}

func (f *CFeature) AddToSource(tx enjinql.SqlTX, sid int64, stub *feature.PageStub, p feature.Page) (err error) {
	if id := p.Permalink(); !id.IsNil() {
		if _, err = tx.Insert(enjinql.PagePermalinkSource, sid, p.PermalinkSha(), id.String()); err != nil {
			err = fmt.Errorf("error adding to permalink source: %q - %w", p.Url(), err)
			return
		}
	}
	return
}

func (f *CFeature) RemoveFromSource(tx enjinql.SqlTX, sid int64, stub *feature.PageStub, p feature.Page) (err error) {
	if _, err = tx.DeleteWhereEQ(enjinql.PagePermalinkSource, enjinql.PageSourceIdKey, sid); err != nil {
		err = fmt.Errorf("error removing from permalink source: %q - %w", p.Url(), err)
	}
	return
}
