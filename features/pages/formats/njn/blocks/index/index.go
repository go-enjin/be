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
	"html/template"
	"strings"

	"github.com/go-enjin/golang-org-x-text/cases"
	"github.com/iancoleman/strcase"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms/nonce"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/page"
	"github.com/go-enjin/be/pkg/pageql"
	"github.com/go-enjin/be/pkg/pages"
	beStrings "github.com/go-enjin/be/pkg/strings"
	"github.com/go-enjin/be/pkg/types/site"
)

const (
	Tag feature.Tag = "NjnIndexBlock"
)

var (
	_ Block     = (*CBlock)(nil)
	_ MakeBlock = (*CBlock)(nil)
)

type Block interface {
	feature.EnjinBlock
}

type MakeBlock interface {
	Make() Block
}

type CBlock struct {
	feature.CEnjinBlock

	enjin feature.Internals
}

func New() (field MakeBlock) {
	f := new(CBlock)
	f.Init(f)
	return f
}

func (f *CBlock) Tag() feature.Tag {
	return Tag
}

func (f *CBlock) Init(this interface{}) {
	f.CEnjinBlock.Init(this)
}

func (f *CBlock) Make() Block {
	return f
}

func (f *CBlock) Setup(enjin feature.Internals) {
	f.enjin = enjin
}

func (f *CBlock) NjnClass() (tagClass feature.NjnClass) {
	tagClass = feature.InlineNjnClass
	return
}

func (f *CBlock) NjnBlockType() (name string) {
	name = "index"
	return
}

func (f *CBlock) PrepareBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (block map[string]interface{}, redirect string, err error) {
	if blockType != "index" {
		err = fmt.Errorf("%v does not implement %v block type", f.Tag(), blockType)
		return
	}

	var blockDataContent map[string]interface{}
	if blockDataContent, err = re.PrepareGenericBlockData(data["content"]); err != nil {
		return
	}
	delete(blockDataContent, "section")

	block = re.PrepareGenericBlock("index", data)

	var ok bool
	var tag string
	if tag, ok = block["Tag"].(string); !ok {
		err = fmt.Errorf("index block missing data-block-tag")
		return
	}

	numPerPage, pageNumber := 10, 0
	if numPerPage, err = maps.ExtractIntValue("index-num-per-page", data); err != nil {
		return
	}
	blockNumPerPage := numPerPage

	var indexViews []string
	if views, ok := data["index-views"].(string); ok {
		for _, view := range strings.Split(views, ",") {
			indexViews = append(indexViews, strcase.ToKebab(strings.TrimSpace(view)))
		}
	} else {
		err = fmt.Errorf("index blocks require an index-view property set")
		return
	}

	var orderBy, sortDir string
	if orderBy, ok = data["index-order-by"].(string); ok {
		orderBy = strcase.ToCamel(orderBy)
	} else {
		orderBy = "Title"
	}

	if sortDir, err = maps.ExtractEnumValue("index-sort-dir", true, []string{"ASC", "DSC"}, data); err != nil {
		return
	}

	var maxResults int
	if maxResults, err = maps.ExtractIntValue("index-max-results", data); err != nil {
		return
	}

	filters := makeFilters(data)

	reqArgv := re.RequestArgv()
	// log.WarnF("reqArgv=%#v", reqArgv.String())

	if reqArgv.NumPerPage > -1 {
		numPerPage = reqArgv.NumPerPage
	}
	if reqArgv.PageNumber > -1 {
		reqArgv.NumPerPage = numPerPage
		pageNumber = reqArgv.PageNumber
		reqArgv.PageNumber = pageNumber
	} else {
		reqArgv.PageNumber = -1
		reqArgv.NumPerPage = -1
	}

	var csqp bool // correct search query paths
	decArgv := site.DecomposeHttpRequest(reqArgv.Request)
	for idx, argv := range decArgv.Argv {
		for jdx, arg := range argv {
			if check := strings.HasPrefix(arg, "%28") && strings.HasSuffix(arg, "%29"); check {
				csqp = check
				arg = arg[3 : len(arg)-3]
				if arg == "" {
					decArgv.Argv[idx][jdx] = ""
				} else {
					decArgv.Argv[idx][jdx] = "(" + arg + ")"
				}
			} else if check := arg != "" && arg[0] == '(' && arg[len(arg)-1] == ')'; check {
				arg = arg[1 : len(arg)-1]
				if arg == "" {
					csqp = check
					decArgv.Argv[idx][jdx] = ""
				} else {
					decArgv.Argv[idx][jdx] = "(" + arg + ")"
				}
			}
		}
	}
	if csqp {
		redirect = decArgv.String()
		return
	}

	searchEnabled := false
	searchNonceKey := fmt.Sprintf("index-block--%v--search-form", tag)
	if check, ok := data["search-enabled"]; ok {
		searchEnabled = maps.ExtractBoolValue(check)
	}

	var argvBlockPresent bool
	var argvView string
	var argvSearch string

	for idx, pieces := range reqArgv.Argv {
		if pieces[0] != "" && pieces[0] == tag {
			argvBlockPresent = true
			re.RequestContext().SetSpecific(site.RequestArgvConsumedKey, true)

			var fixArgs []string
			var viewArgs []string
			for _, piece := range pieces[1:] {
				if updated := filters.SetPresent(piece); updated {
					fixArgs = append(fixArgs, piece)
				} else if beStrings.StringInSlices(piece, indexViews) {
					if argvView == "" {
						argvView = piece
						viewArgs = append(viewArgs, piece)
					} else {
						fixArgs = append(fixArgs, piece)
					}
				} else if searchEnabled && piece != "" && piece[0] == '(' && piece[len(piece)-1] == ')' {
					argvSearch = piece[1 : len(piece)-1] // trim '(' and ')'
					fixArgs = append(fixArgs, "("+argvSearch+")")
				} else {
					// 	fixArgs = append(fixArgs, piece)
				}
			}
			fixArgs = append(viewArgs, fixArgs...)
			if len(fixArgs) != len(pieces[1:]) {
				reqArgv.Argv[idx] = append([]string{pieces[0]}, fixArgs...)
				// re.RequestContext().SetSpecific(site.RequestRedirectKey, reqArgv.String())
				redirect = reqArgv.String()
				return
			}
		}
	}

	block["View"] = argvView
	block["NumPerPage"] = numPerPage

	block["SearchEnabled"] = searchEnabled
	if searchEnabled {
		block["SearchQuery"] = argvSearch
		block["SearchNonce"] = nonce.Make(searchNonceKey)
		if argvBlockPresent {
			if searchRedirect, searchError := f.handleSearchRedirect(tag, searchNonceKey, indexViews, reqArgv); searchError != nil {
				block["SearchError"] = searchError.Error()
			} else if searchRedirect != "" {
				redirect = searchRedirect
				return
			}
		}
	}

	var query string
	if query, ok = data["index-query"].(string); !ok {
		err = fmt.Errorf("index blocks require an index-query property set")
		return
	}

	if err = pageql.Validate(query); err != nil {
		err = fmt.Errorf("query error: %v - %v", query, err)
		return
	}

	found := f.enjin.MatchQL(query)
	found = page.SortPages(found, orderBy, sortDir)

	totalFound := len(found)

	found = filters.FilterPages(found)

	if maxResults > 0 && totalFound > maxResults {
		found = found[:maxResults]
	}

	totalFiltered := len(found)

	if searchEnabled && argvSearch != "" {
		if len(found) == 0 {
			// nope
		} else if matched, searchResults, e := pages.SearchWithin(argvSearch, totalFiltered, 0, found, f.enjin.SiteDefaultLanguage(), reqArgv.Language, f.enjin.SiteLanguageMode()); e != nil {
			log.ErrorF("error searching within... %v", err)
			found = nil
		} else {
			block["SearchWithinTotal"] = totalFiltered
			block["SearchResults"] = searchResults
			var updated []*page.Page
			for _, hit := range searchResults.Hits {
				if pg, ok := matched[hit.ID]; ok {
					updated = append(updated, pg)
				}
			}
			totalFiltered = len(updated)
			block["SearchTotal"] = totalFiltered

			log.DebugF("search found: %d (of %d total) hits for query: %v", len(updated), len(found), argvSearch)
			searchRanked := true
			if ranked, ok := data["search-ranked"]; ok {
				searchRanked = maps.ExtractBoolValue(ranked)
			}
			if searchRanked {
				found = updated
			} else {
				found = page.SortPages(updated, orderBy, sortDir)
			}
		}
	}

	totalPages := totalFiltered / numPerPage

	if pageNumber > totalPages {
		reqArgv.PageNumber = totalPages - 1
		pageNumber = reqArgv.PageNumber
		// re.RequestContext().SetSpecific(site.RequestRedirectKey, reqArgv.String())
		redirect = reqArgv.String()
		return
	}
	if numPerPage > 0 && totalFiltered > 0 {
		start := pageNumber * numPerPage
		end := start + numPerPage
		if start < end && end < totalFiltered {
			found = found[start:end]
		} else {
			found = found[start:]
		}
	}

	block["Results"] = found

	var pgntn string
	if pgntn, ok = data["index-pagination"].(string); ok {
		pgntn = strings.ToLower(pgntn)
		switch pgntn {
		case "none", "more", "page":
		default:
			err = fmt.Errorf("invalid index block pagination value: %v", pgntn)
			return
		}
	} else {
		pgntn = "none"
	}

	var builtViews Views
	for idx, viewKey := range indexViews {
		viewFilters := filters.Copy()
		if searchEnabled && argvSearch != "" {
			viewFilters.UpdateUrls(tag, reqArgv.Path, viewKey, "("+argvSearch+")")
		} else {
			viewFilters.UpdateUrls(tag, reqArgv.Path, viewKey)
		}

		titleCase := cases.Title(reqArgv.Language)
		viewTitle := titleCase.String(strcase.ToDelimited(viewKey, ' '))
		view := makeView(idx, viewKey, viewTitle, viewFilters)
		if argvView == viewKey {
			view.Present = true
		}

		cra := reqArgv.Copy()
		argv := []string{tag, view.Key}
		for _, groups := range viewFilters {
			for _, filter := range groups {
				if filter.Present {
					argv = append(argv, filter.Key)
				}
			}
		}
		if searchEnabled {
			sfa := cra.Copy()
			sfa.NumPerPage = -1
			sfa.PageNumber = -1
			sfa.Argv = append([][]string{}, argv)
			view.SearchAction = sfa.String()
		}
		if searchEnabled && argvSearch != "" {
			argv = append(argv, "("+argvSearch+")")
		}
		cra.Argv = append([][]string{}, argv)

		view.Paginate = pgntn

		if numPerPage < totalFiltered {

			// more pagination
			moreCra := cra.Copy()
			count := numPerPage / blockNumPerPage
			count += 1
			moreCra.PageNumber = 0
			moreCra.NumPerPage = count * blockNumPerPage
			if moreCra.NumPerPage > totalFiltered {
				moreCra.NumPerPage = totalFiltered
			}
			view.NextMore = moreCra.String() + "#" + tag + "-" + viewKey
			// page pagination
			if totalPages > 0 {
				pageCra := cra.Copy()
				pageCra.PageNumber = pageNumber
				pageCra.NumPerPage = numPerPage
				if pageNumber > 0 {
					pageCra.PageNumber = 0
					view.FirstPage = pageCra.String() + "#" + tag + "-" + viewKey
					pageCra.PageNumber = pageNumber - 1
					view.PrevPage = pageCra.String() + "#" + tag + "-" + viewKey
				}

				view.PageNumber = pageNumber + 1
				view.TotalPages = totalPages

				if pageNumber < totalPages-1 {
					pageCra.PageNumber = pageNumber + 1
					view.NextPage = pageCra.String() + "#" + tag + "-" + viewKey
					pageCra.PageNumber = totalPages - 1
					view.LastPage = pageCra.String() + "#" + tag + "-" + viewKey
				}
			}
		}
		view.Url = strings.TrimSuffix(reqArgv.Path, "/") + "/:" + tag + "," + viewKey + "#" + tag + "-" + viewKey
		builtViews = append(builtViews, view)
	}
	block["Views"] = builtViews

	log.DebugF("index block found %v (total=%v, max=%v) pages with query: %v", totalFiltered, totalFound, maxResults, query)

	if heading, ok := re.PrepareBlockHeader(blockDataContent); ok {
		block["Heading"] = heading
	}

	if footer, ok := re.PrepareBlockFooter(blockDataContent); ok {
		block["Footer"] = footer
	}

	return
}

func (f *CBlock) RenderPreparedBlock(re feature.EnjinRenderer, block map[string]interface{}) (html template.HTML, err error) {
	html, err = re.RenderNjnTemplate("block/index", block)
	return
}

func (f *CBlock) ProcessBlock(re feature.EnjinRenderer, blockType string, data map[string]interface{}) (html template.HTML, redirect string, err error) {
	if block, redir, e := f.PrepareBlock(re, blockType, data); e != nil {
		err = e
		return
	} else if redir != "" {
		redirect = redir
		return
	} else {
		html, err = f.RenderPreparedBlock(re, block)
	}
	return
}