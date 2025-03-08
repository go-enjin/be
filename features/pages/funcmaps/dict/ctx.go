//go:build page_funcmaps || pages || all

// Copyright (c) 2025  The Go-Enjin Authors
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

package dict

import (
	"fmt"

	clContext "github.com/go-corelibs/context"
	"github.com/samber/lo"
)

func NewContext(argv ...interface{}) (c clContext.Context, err error) {
	if argc := len(argv); argc > 0 && argc%2 != 0 {
		if ctx, ok := argv[0].(Dictionary); ok {
			c, err = WrapContext(ctx, argv[1:]...)
			return
		} else if ctx, ok := argv[0].(clContext.Context); ok {
			c, err = WrapContext(ctx, argv[1:]...)
			return
		} else if ctx, ok := argv[0].(map[string]interface{}); ok {
			c, err = WrapContext(ctx, argv[1:]...)
			return
		}
		err = fmt.Errorf("expected Dictionary, clContext.Context{} or map[string]interface{}, received: %T", argv[0])
		return
	}
	// argv is balanced k=v pairs
	c = clContext.Context{}
	for i := 0; i < len(argv); i += 2 {
		k, v := argv[i], argv[i+1]
		if key, ok := k.(string); ok {
			c[key] = v
		} else {
			err = fmt.Errorf("context key is not a string (idx %d): %#+v", i, k)
			return
		}
	}
	return
}

func WrapContext(ctx map[string]interface{}, argv ...interface{}) (c clContext.Context, err error) {
	c = lo.Assign(clContext.Context{}, ctx)
	// _, err = d.SetSpecific(argv...)
	return
}
