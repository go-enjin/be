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
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-enjin/golang-org-x-text/language"
	"golang.org/x/net/html"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/log"
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
	rxPageRequestArgv  = regexp.MustCompile(`^(/[^:]*)((?:/:[^/]+)*)(/\d+/\d+/?)?$`)
)

type RequestArgv struct {
	Path       string
	Argv       [][]string
	NumPerPage int
	PageNumber int
	Language   language.Tag
	Request    *http.Request
}

func (ra *RequestArgv) MustConsume() (must bool) {
	must = len(ra.Argv) > 0
	return
}

func (ra *RequestArgv) Set(r *http.Request) (req *http.Request) {
	req = r.WithContext(context.WithValue(r.Context(), RequestArgvKey, ra))
	return
}

func (ra *RequestArgv) Copy() (reqArg *RequestArgv) {
	var argv [][]string
	for _, group := range ra.Argv {
		argv = append(argv, append([]string{}, group...))
	}
	reqArg = &RequestArgv{
		Path:       ra.Path,
		Argv:       argv,
		NumPerPage: ra.NumPerPage,
		PageNumber: ra.PageNumber,
		Request:    ra.Request,
	}
	return
}

func (ra *RequestArgv) String() (argvUrl string) {
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

func GetRequestArgv(r *http.Request) (reqArgv *RequestArgv) {
	reqArgv, _ = r.Context().Value(RequestArgvKey).(*RequestArgv)
	return
}

func DecomposeHttpRequest(r *http.Request) (reqArgv *RequestArgv) {
	path := forms.TrimQueryParams(r.RequestURI)
	var argv [][]string
	numPerPage, pageNumber := -1, -1
	if rxPageRequestArgv.MatchString(path) {
		m := rxPageRequestArgv.FindAllStringSubmatch(path, 1)
		path = strings.TrimSuffix(m[0][1], "/")
		if args := m[0][2]; args != "" {
			args = args[2:] // remove leading "/:"
			parts := rxPageRequestSplit.Split(args, -1)
			for _, part := range parts {
				argv = append(argv, strings.Split(part, ","))
			}
		}
		// log.WarnF("path=%v, uri=%v\nm=%v", path, r.RequestURI, argv)
		if pgntn := m[0][3]; pgntn != "" {
			pgntn = strings.TrimPrefix(strings.TrimSuffix(pgntn, "/"), "/")
			parts := strings.Split(pgntn, "/")
			// log.WarnF("pgntn: %v - %#v", pgntn, parts)
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
		Request:    r,
	}
	return
}

func DecodeHttpRequest(r *http.Request) (reqArgv *RequestArgv) {
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