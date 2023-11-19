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

package site

import (
	"fmt"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/kvs"
)

func (f *CFeature) PushImportantNotice(eid string, dismiss bool, message string, argv ...interface{}) {
	f.PushNotices(eid, feature.MakeImportantNotice(dismiss, message, argv...))
}

func (f *CFeature) PushInfoNotice(eid string, dismiss bool, message string, argv ...interface{}) {
	f.PushNotices(eid, feature.MakeInfoNotice(dismiss, message, argv...))
}

func (f *CFeature) PushWarnNotice(eid string, dismiss bool, message string, argv ...interface{}) {
	f.PushNotices(eid, feature.MakeWarnNotice(dismiss, message, argv...))
}

func (f *CFeature) PushErrorNotice(eid string, dismiss bool, message string, argv ...interface{}) {
	f.PushNotices(eid, feature.MakeErrorNotice(dismiss, message, argv...))
}

func (f *CFeature) PushNotices(eid string, notices ...*feature.UserNotice) {
	f.userNoticeLocker.Lock(eid)
	defer f.userNoticeLocker.Unlock(eid)

	for _, notice := range notices {
		if err := kvs.AppendToFlatList(f.userNoticeBucket, eid, notice); err != nil {
			panic(fmt.Errorf("error appending to site notices flat list: %v - %#+v", err, notice))
		}
	}
	return
}

func (f *CFeature) PullNotices(eid string) (notices feature.UserNotices) {
	f.userNoticeLocker.Lock(eid)
	defer f.userNoticeLocker.Unlock(eid)

	for notice := range kvs.YieldFlatList[*feature.UserNotice](f.userNoticeBucket, eid) {
		notices = append(notices, notice)
	}

	kvs.ResetFlatList(f.userNoticeBucket, eid)
	return
}