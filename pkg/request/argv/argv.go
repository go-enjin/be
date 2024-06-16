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

package argv

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/go-corelibs/x-text/language"

	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request"
)

const (
	RequestKey         request.Key = "RequestArgv"
	RequestRedirectKey string      = "RequestRedirect"
	RequestIgnoredKey  string      = "RequestArgvIgnored"
	RequestConsumedKey string      = "RequestArgvConsumed"
)

type Argv struct {
	Path       string
	Argv       [][]string
	NumPerPage int
	PageNumber int
	Language   language.Tag
	Request    *http.Request
}

func (ra *Argv) MustConsume() (must bool) {
	must = len(ra.Argv) > 0
	return
}

func (ra *Argv) Set(r *http.Request) (req *http.Request) {
	req = r.Clone(context.WithValue(r.Context(), RequestKey, ra))
	return
}

func (ra *Argv) Copy() (reqArg *Argv) {
	var argv [][]string
	for _, group := range ra.Argv {
		argv = append(argv, append([]string{}, group...))
	}
	reqArg = &Argv{
		Path:       ra.Path,
		Argv:       argv,
		NumPerPage: ra.NumPerPage,
		PageNumber: ra.PageNumber,
		Request:    ra.Request,
	}
	return
}

func (ra *Argv) String() (argvUrl string) {
	argvUrl = strings.TrimSuffix(ra.Path, "/")
	for _, pieces := range ra.Argv {
		argvUrl += "/:"
		for idx, piece := range pieces {
			if piece != "" {
				if idx > 0 {
					argvUrl += ","
				}
				if piece != "" && piece[0] == '(' && piece[len(piece)-1] == ')' {
					argvUrl += "(" + url.PathEscape(piece[1:len(piece)-1]) + ")"
				} else {
					argvUrl += url.PathEscape(piece)
				}
			}
		}
	}
	if ra.NumPerPage > -1 && ra.PageNumber > -1 {
		argvUrl += fmt.Sprintf("/%v/%v/", ra.NumPerPage, ra.PageNumber)
	}
	return
}

func Get(r *http.Request) (reqArgv *Argv) {
	reqArgv, _ = r.Context().Value(RequestKey).(*Argv)
	return
}

func DecomposeHttpRequest(r *http.Request) (reqArgv *Argv) {
	if r == nil {
		reqArgv = &Argv{}
		return
	}
	path := forms.TrimQueryParams(r.RequestURI)
	var argv [][]string
	numPerPage, pageNumber := -1, -1

	// path, args, pgntn

	if m := rxRequestPageSize.FindAllStringSubmatch(path, 1); len(m) == 1 {
		numPerPage, _ = strconv.Atoi(m[0][1])
		pageNumber, _ = strconv.Atoi(m[0][2])
		path = strings.TrimSuffix(path, m[0][0])
	} else if m = rxRequestPageOnly.FindAllStringSubmatch(path, 1); len(m) == 1 {
		pageNumber, _ = strconv.Atoi(m[0][1])
		path = strings.TrimSuffix(path, m[0][0])
	}

	for {
		if m := rxRequestArgv.FindAllStringSubmatch(path, 1); len(m) == 1 {
			parts := strings.Split(m[0][1], ",")
			argv = append(argv, parts)
			path = strings.TrimSuffix(path, m[0][0])
		} else {
			break
		}
	}

	path = strings.TrimSuffix(path, "/")

	reqArgv = &Argv{
		Path:       path,
		Argv:       argv,
		NumPerPage: numPerPage,
		PageNumber: pageNumber,
		Request:    r,
	}
	return
}

func DecodeHttpRequest(r *http.Request) (reqArgv *Argv) {
	reqArgv = DecomposeHttpRequest(r)
	for idx, argv := range reqArgv.Argv {
		var args []string
		for _, arg := range argv {
			if cleaned, err := url.PathUnescape(arg); err != nil {
				log.ErrorF("error unescaping argument: %v", arg)
			} else {
				args = append(args, html.UnescapeString(cleaned))
			}
		}
		reqArgv.Argv[idx] = args
	}
	return
}
