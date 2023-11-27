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

package spinlockers

import (
	"time"

	berrs "github.com/go-enjin/be/pkg/errors"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/kvs"
)

var (
	DefaultSpinlockerTimeout  = time.Second * 10
	DefaultSpinlockerInterval = time.Millisecond * 50
)

type spinlocker struct {
	tag         feature.Tag
	key         string
	readBucket  feature.KeyValueStore
	writeBucket feature.KeyValueStore
	timeout     time.Duration
	interval    time.Duration
}

func newSpinLocker(bucket feature.KeyValueStore, tag feature.Tag, key string, timeout, interval time.Duration) (tl feature.SyncLocker) {
	if timeout < 1 {
		timeout = DefaultSpinlockerTimeout
	}
	if interval < 1 {
		interval = DefaultSpinlockerInterval
	}
	tl = &spinlocker{
		tag:         tag,
		key:         key,
		writeBucket: bucket,
		timeout:     timeout,
		interval:    interval,
	}
	return
}

func newSpinRWLocker(readStore, writeStore feature.KeyValueStore, tag feature.Tag, key string, timeout, interval time.Duration) (tl feature.SyncRWLocker) {
	if timeout < 1 {
		timeout = DefaultSpinlockerTimeout
	}
	if interval < 1 {
		interval = DefaultSpinlockerInterval
	}
	tl = &spinlocker{
		tag:         tag,
		key:         key,
		readBucket:  readStore,
		writeBucket: writeStore,
		timeout:     timeout,
		interval:    interval,
	}
	return
}

func (l *spinlocker) makeKey() (key string) {
	if l.tag.IsNil() {
		key = l.key
	} else {
		key = l.tag.Kebab() + "_" + l.key
	}
	return
}

func (l *spinlocker) IsLocked() (locked bool) {
	locked = !kvs.FlatListEmpty(l.writeBucket, l.makeKey())
	return
}

func (l *spinlocker) spinWrite(rid string) (spent time.Duration) {
	for spent < l.timeout {
		if lockedBy, ok := kvs.FirstInFlatList[string](l.writeBucket, l.makeKey()); ok && lockedBy == rid {
			break
		}
		time.Sleep(l.interval)
		spent += l.interval
	}
	return
}

func (l *spinlocker) spinRead() (spent time.Duration) {
	for spent < l.timeout {
		if kvs.FlatListEmpty(l.writeBucket, l.makeKey()) {
			break
		}
		time.Sleep(l.interval)
		spent += l.interval
	}
	return
}

func (l *spinlocker) Lock(rid string) {
	key := l.makeKey()
	berrs.Must(kvs.AppendToFlatList(l.writeBucket, key, rid))
	if kvs.CountFlatList(l.writeBucket, key) == 1 {
		// early out because we're the one
	} else if l.spinWrite(rid) >= l.timeout {
		berrs.Must(kvs.RemoveFromFlatList(l.writeBucket, key, rid))
		panic(berrs.ErrFileSystemLockTimeout)
	}
	return
}

func (l *spinlocker) Unlock(rid string) {
	berrs.Must(kvs.RemoveFromFlatList(l.writeBucket, l.makeKey(), rid))
	return
}

func (l *spinlocker) RLock(rid string) {
	key := l.makeKey()
	berrs.Must(kvs.AppendToFlatList(l.readBucket, key, rid))
	if kvs.FlatListEmpty(l.writeBucket, key) {
		// early out
	} else if l.spinRead() >= l.timeout {
		berrs.Must(kvs.RemoveFromFlatList(l.readBucket, key, rid))
		panic(berrs.ErrFileSystemLockTimeout)
	}
	return
}

func (l *spinlocker) RUnlock(rid string) {
	berrs.Must(kvs.RemoveFromFlatList(l.readBucket, l.makeKey(), rid))
	return
}
