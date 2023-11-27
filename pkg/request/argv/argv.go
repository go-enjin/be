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
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/go-enjin/golang-org-x-text/language"

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

const (
	PatternArgs  = `((?:/:[^/]+)+?)`
	PatternPgntn = `(/\d+/\d+/??)`
)

var (
	RxRequestSplit = regexp.MustCompile(`/:`)
	RxRequestCase0 = regexp.MustCompile(`^(/[^:]+?)` + PatternArgs + PatternPgntn + `$`)
	RxRequestCase1 = regexp.MustCompile(`^(/[^:]+?)` + PatternPgntn + `$`)
	RxRequestCase2 = regexp.MustCompile(`^(/[^:]+?)` + PatternArgs + `$`)
	RxRequestCase3 = regexp.MustCompile(`^(/[^:]+?)$`)
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

	var vPath, vArgs, vPgntn string

	switch {

	case RxRequestCase0.MatchString(path): // all three segments
		m := RxRequestCase0.FindAllStringSubmatch(path, 1)
		vPath, vArgs, vPgntn = m[0][1], m[0][2], m[0][3]

	case RxRequestCase1.MatchString(path): // first and third
		m := RxRequestCase1.FindAllStringSubmatch(path, 1)
		vPath, vPgntn = m[0][1], m[0][2]

	case RxRequestCase2.MatchString(path): // first and second
		m := RxRequestCase2.FindAllStringSubmatch(path, 1)
		vPath, vArgs = m[0][1], m[0][2]

	case RxRequestCase3.MatchString(path): // first only
		m := RxRequestCase3.FindAllStringSubmatch(path, 1)
		vPath = m[0][1]
	}

	path = strings.TrimSuffix(vPath, "/")

	if args := vArgs; args != "" {
		args = args[2:] // remove leading "/:"
		parts := RxRequestSplit.Split(args, -1)
		for _, part := range parts {
			argv = append(argv, strings.Split(part, ","))
		}
	}

	// log.WarnF("path=%v, uri=%v\nm=%v", path, r.RequestURI, argv)
	if pgntn := vPgntn; pgntn != "" {
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
