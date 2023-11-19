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
	"time"
)

type SyncLockerFactory interface {
	NewSyncLocker(tag Tag, key string, store KeyValueStore) (l SyncLocker)
	NewSyncLockerWith(tag Tag, key string, store KeyValueStore, timeout, interval time.Duration) (l SyncLocker)
	NewSyncRWLocker(tag Tag, key string, readStore, writeStore KeyValueStore) (l SyncRWLocker)
	NewSyncRWLockerWith(tag Tag, key string, readStore, writeStore KeyValueStore, timeout, interval time.Duration) (l SyncRWLocker)
}

type SyncLockerFactoryFeature interface {
	Feature
	SyncLockerFactory
}