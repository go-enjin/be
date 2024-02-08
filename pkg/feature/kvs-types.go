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
	"context"

	"github.com/urfave/cli/v2"
)

type KeyValueCaches interface {
	Feature

	Get(name string) (kvs KeyValueCache, err error)
}

type KeyValueCache interface {
	// ListBuckets returns the names of all buckets
	ListBuckets() (names []string)
	// Bucket returns the named bucket or adds a new bucket and returns that
	Bucket(name string) (kvs KeyValueStore, err error)
	// MustBucket uses Bucket and on error will log.ErrorDF and panic
	MustBucket(name string) (kvs KeyValueStore)
	// AddBucket adds and returns a new bucket, errors if already exists
	AddBucket(name string) (kvs KeyValueStore, err error)
	// GetBucket returns a new bucket, errors if not found
	GetBucket(name string) (kvs KeyValueStore, err error)
	// GetBucketSource returns the underlying cache object, or nil if not found
	GetBucketSource(name string) (src interface{})
}

type KeyValueCacheFeature interface {
	Build(b Buildable) (err error)
	Startup(ctx *cli.Context) (err error)
	Shutdown()
}

type KeyValueStore interface {
	Get(key string) (value []byte, err error)
	Set(key string, value []byte) (err error)
	Delete(key string) (err error)
}

type KeyValueStoreRangeFn func(key string, value []byte) (stop bool)

type ExtendedKeyValueStore interface {
	KeyValueStore

	Size() (count int)
	Keys(prefix string) (keys []string)
	Range(prefix string, fn KeyValueStoreRangeFn)
	StreamKeys(prefix string, ctx context.Context) (keys chan string)
}
