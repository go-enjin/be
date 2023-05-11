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

package kvs

type KeyValueCaches interface {
	Get(name string) (kvs KeyValueCache, err error)
}

type KeyValueCache interface {
	// Bucket returns the named bucket or adds a new bucket and returns that
	Bucket(name string) (kvs KeyValueStore, err error)
	// MustBucket uses Bucket and log.FatalDF on error
	MustBucket(name string) (kvs KeyValueStore)
	// AddBucket adds and returns a new bucket, errors if already exists
	AddBucket(name string) (kvs KeyValueStore, err error)
	// GetBucket returns a new bucket, errors if not found
	GetBucket(name string) (kvs KeyValueStore, err error)
}

type KeyValueStore interface {
	Get(key interface{}) (value interface{}, err error)
	Set(key interface{}, value interface{}) (err error)
}

type KeyValueStoreAny interface {
	Get(key interface{}) (value interface{}, ok bool)
	Set(key interface{}, value interface{})
}