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

package index

import (
	"fmt"
	"html"
	"net/url"

	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/forms/nonce"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/request/argv"
	beStrings "github.com/go-enjin/be/pkg/strings"
)

func (f *CBlock) handleSearchRedirect(blockTag, nonceKey string, viewKeys []string, reqArgv *argv.RequestArgv) (redirect string, err error) {

	// tag := lang.GetTag(r)
	printer := lang.GetPrinterFromRequest(reqArgv.Request)
	var query string
	var foundNonce, foundQuery bool
	for k, v := range reqArgv.Request.URL.Query() {
		switch k {

		case "nonce":
			var value string
			if vv, e := url.QueryUnescape(v[0]); e != nil {
				log.ErrorF("error un-escaping url path: %v", e)
			} else {
				value = forms.StrictPolicy(vv)
			}
			if nonce.Validate(nonceKey, value) {
				foundNonce = true
			} else {
				err = fmt.Errorf(printer.Sprintf("search form expired"))
				break
			}

		case "query":
			if vv, e := url.QueryUnescape(v[0]); e != nil {
				log.ErrorF("error un-escaping url path: %v", e)
			} else {
				query = html.UnescapeString(vv)
				query = forms.StrictPolicy(vv)
			}
			foundQuery = true
		}
		if foundQuery && foundNonce || err != nil {
			break
		}
	}

	if err != nil {
		log.ErrorF("index block search redirect error: %v", err)
		err = nil
		target := blockTag
		var argvs [][]string
		for _, argvArgs := range reqArgv.Argv {
			var args []string
			for _, arg := range argvArgs {
				if target == blockTag && beStrings.StringInSlices(arg, viewKeys) {
					target += "-" + arg
				}
				if skip := arg != "" && arg[0] == '(' && arg[len(arg)-1] == ')'; !skip {
					args = append(args, arg)
				}
			}
			argvs = append(argvs, args)
		}
		reqArgv.Argv = argvs
		redirect = reqArgv.String() + "#" + target
		return
	}

	if foundQuery && foundNonce {
		for idx, argvArgs := range reqArgv.Argv {
			if len(argvArgs) > 0 && argvArgs[0] == blockTag {
				target := blockTag
				for jdx, arg := range argvArgs {
					if target == blockTag && beStrings.StringInSlices(arg, viewKeys) {
						target += "-" + arg
					}
					if arg != "" && arg[0] == '(' && arg[len(arg)-1] == ')' {
						if arg[1:len(arg)-1] != "" {
							reqArgv.Argv[idx] = append(reqArgv.Argv[idx][:jdx], reqArgv.Argv[idx][jdx+1:]...)
						}
					}
				}
				if query != "" {
					reqArgv.Argv[idx] = append(reqArgv.Argv[idx], "("+query+")")
				}
				redirect = reqArgv.String() + "#" + target
				return
			}
		}
	}

	return
}