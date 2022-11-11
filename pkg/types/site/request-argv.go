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

package site

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	beContext "github.com/go-enjin/be/pkg/context"
)

const (
	RequestArgvKey         beContext.RequestKey = "RequestArgv"
	RequestRedirectKey     string               = "RequestRedirect"
	RequestArgvIgnoredKey  string               = "RequestArgvIgnored"
	RequestArgvConsumedKey string               = "RequestArgvConsumed"
)

var (
	rxPageRequestSplit = regexp.MustCompile(`/:`)
	rxPageRequestSafe  = regexp.MustCompile(`/:.*$`)
	rxPageRequestArgv  = regexp.MustCompile(`^(/[^:]*)((?:/:[^/]+)*)(/\d+/\d+/??)?$`)
)

type RequestArgv struct {
	Path       string
	Argv       [][]string
	NumPerPage int
	PageNumber int
}

func (ra *RequestArgv) MustConsume() (must bool) {
	must = len(ra.Argv) > 0
	return
}

func (ra *RequestArgv) Set(r *http.Request) (req *http.Request) {
	req = r.WithContext(context.WithValue(r.Context(), RequestArgvKey, ra))
	return
}

func (ra *RequestArgv) String() (url string) {
	url = ra.Path
	for _, pieces := range ra.Argv {
		url += "/:" + strings.Join(pieces, ",")
	}
	if ra.NumPerPage > -1 {
		url += fmt.Sprintf("/%v", ra.NumPerPage)
	}
	if ra.PageNumber > -1 {
		url += fmt.Sprintf("/%v", ra.PageNumber)
	}
	return
}

func GetRequestArgv(r *http.Request) (reqArgv *RequestArgv) {
	reqArgv, _ = r.Context().Value(RequestArgvKey).(*RequestArgv)
	return
}

func ParseRequestArgv(path string) (reqArgv *RequestArgv) {
	var argv [][]string
	numPerPage, pageNumber := -1, -1
	if rxPageRequestArgv.MatchString(path) {
		m := rxPageRequestArgv.FindAllStringSubmatch(path, 1)
		path = m[0][1]
		if args := m[0][2]; args != "" {
			args = args[2:] // remove leading "/:"
			parts := rxPageRequestSplit.Split(args, -1)
			for _, part := range parts {
				var pieces []string
				for _, piece := range strings.Split(part, ",") {
					pieces = append(pieces, piece)
				}
				argv = append(argv, pieces)
			}
		}
		if pgntn := m[0][3]; pgntn != "" {
			pgntn = pgntn[1:] // remove leading "/"
			parts := strings.Split(pgntn, "/")
			switch len(parts) {
			case 1:
				if v, err := strconv.Atoi(parts[0]); err == nil && v >= 0 {
					pageNumber = v
				}
			case 2:
				if v, err := strconv.Atoi(parts[0]); err == nil && v >= 0 {
					numPerPage = v
				}
				if v, err := strconv.Atoi(parts[1]); err == nil && v >= 0 {
					pageNumber = v
				}
			}
		}
	} else {
		path = rxPageRequestSafe.ReplaceAllString(path, "")
	}
	reqArgv = &RequestArgv{
		Path:       path,
		Argv:       argv,
		NumPerPage: numPerPage,
		PageNumber: pageNumber,
	}
	return
}