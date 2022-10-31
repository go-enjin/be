// Code generated with _scripts/bg-pkg-list.sh DO NOT EDIT.

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

package be

func GoEnjinPackageList() (list []string) {
	list = append(list,
		"github.com/go-enjin/be",
		"github.com/go-enjin/be/pkg/autocert",
		"github.com/go-enjin/be/pkg/cli/env",
		"github.com/go-enjin/be/pkg/cli/git",
		"github.com/go-enjin/be/pkg/cli/run",
		"github.com/go-enjin/be/pkg/cli/tar",
		"github.com/go-enjin/be/pkg/cli/zip",
		"github.com/go-enjin/be/pkg/context",
		"github.com/go-enjin/be/pkg/database",
		"github.com/go-enjin/be/pkg/feature",
		"github.com/go-enjin/be/pkg/forms",
		"github.com/go-enjin/be/pkg/forms/nonce",
		"github.com/go-enjin/be/pkg/fs",
		"github.com/go-enjin/be/pkg/fs/embed",
		"github.com/go-enjin/be/pkg/fs/local",
		"github.com/go-enjin/be/pkg/globals",
		"github.com/go-enjin/be/pkg/hash/sha",
		"github.com/go-enjin/be/pkg/ints",
		"github.com/go-enjin/be/pkg/lang",
		"github.com/go-enjin/be/pkg/log",
		"github.com/go-enjin/be/pkg/maps",
		"github.com/go-enjin/be/pkg/menu",
		"github.com/go-enjin/be/pkg/net",
		"github.com/go-enjin/be/pkg/net/gorilla-handlers",
		"github.com/go-enjin/be/pkg/net/headers",
		"github.com/go-enjin/be/pkg/net/ip/deny",
		"github.com/go-enjin/be/pkg/net/ip/ranges/atlassian",
		"github.com/go-enjin/be/pkg/net/ip/ranges/cloudflare",
		"github.com/go-enjin/be/pkg/net/minify",
		"github.com/go-enjin/be/pkg/notify",
		"github.com/go-enjin/be/pkg/page",
		"github.com/go-enjin/be/pkg/path",
		"github.com/go-enjin/be/pkg/path/embed",
		"github.com/go-enjin/be/pkg/search",
		"github.com/go-enjin/be/pkg/search/indexes",
		"github.com/go-enjin/be/pkg/slug",
		"github.com/go-enjin/be/pkg/strings",
		"github.com/go-enjin/be/pkg/theme",
		"github.com/go-enjin/be/pkg/theme/funcs",
		"github.com/go-enjin/be/pkg/types/site",
		"github.com/go-enjin/be/pkg/types/theme-types",
		"github.com/go-enjin/be/features/database",
		"github.com/go-enjin/be/features/fs/databases/content",
		"github.com/go-enjin/be/features/fs/embeds/content",
		"github.com/go-enjin/be/features/fs/embeds/locales",
		"github.com/go-enjin/be/features/fs/embeds/menu",
		"github.com/go-enjin/be/features/fs/embeds/public",
		"github.com/go-enjin/be/features/fs/locals/content",
		"github.com/go-enjin/be/features/fs/locals/locales",
		"github.com/go-enjin/be/features/fs/locals/menu",
		"github.com/go-enjin/be/features/fs/locals/public",
		"github.com/go-enjin/be/features/log/papertrail",
		"github.com/go-enjin/be/features/notify/slack",
		"github.com/go-enjin/be/features/outputs/htmlify",
		"github.com/go-enjin/be/features/outputs/minify",
		"github.com/go-enjin/be/features/pages/context",
		"github.com/go-enjin/be/features/pages/formats",
		"github.com/go-enjin/be/features/pages/formats/html",
		"github.com/go-enjin/be/features/pages/formats/md",
		"github.com/go-enjin/be/features/pages/formats/njn",
		"github.com/go-enjin/be/features/pages/formats/njn/blocks/card",
		"github.com/go-enjin/be/features/pages/formats/njn/blocks/carousel",
		"github.com/go-enjin/be/features/pages/formats/njn/blocks/content",
		"github.com/go-enjin/be/features/pages/formats/njn/blocks/header",
		"github.com/go-enjin/be/features/pages/formats/njn/blocks/icon",
		"github.com/go-enjin/be/features/pages/formats/njn/blocks/image",
		"github.com/go-enjin/be/features/pages/formats/njn/blocks/linkList",
		"github.com/go-enjin/be/features/pages/formats/njn/blocks/notice",
		"github.com/go-enjin/be/features/pages/formats/njn/blocks/pair",
		"github.com/go-enjin/be/features/pages/formats/njn/blocks/sidebar",
		"github.com/go-enjin/be/features/pages/formats/njn/blocks/toc",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/anchor",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/code",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/container",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/details",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/fa",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/fieldset",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/figure",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/footnote",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/img",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/inline",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/input",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/list",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/literal",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/optgroup",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/option",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/p",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/picture",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/pre",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/select",
		"github.com/go-enjin/be/features/pages/formats/njn/fields/table",
		"github.com/go-enjin/be/features/pages/formats/org",
		"github.com/go-enjin/be/features/pages/search",
		"github.com/go-enjin/be/features/requests/cloudflare",
		"github.com/go-enjin/be/features/requests/headers/proxy",
		"github.com/go-enjin/be/features/restrict/basic-auth",
	)
	return
}