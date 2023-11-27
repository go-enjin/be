//go:build page_search || pages || all

// Copyright (c) 2022  The Go-Enjin Authors
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

package search

import (
	"fmt"

	"github.com/go-enjin/golang-org-x-text/language"
	"github.com/urfave/cli/v2"
)

func (f *CFeature) SearchAction(ctx *cli.Context) (err error) {
	var input string
	var size, pg int
	argv := ctx.Args().Slice()
	argc := len(argv)
	if argc == 0 {
		err = fmt.Errorf("search query is required")
		return
	}
	input = argv[0]
	if v := ctx.Int("size"); v > 0 {
		size = v
	} else {
		size = 10
	}
	if v := ctx.Int("pg"); v > 1 {
		pg = v - 1
	}
	query := f.search.PrepareSearch(language.Und, input)
	if results, e := f.search.PerformSearch(language.Und, query, size, pg); e != nil {
		err = e
	} else {
		fmt.Printf("%v\n", results)
	}
	return
}
