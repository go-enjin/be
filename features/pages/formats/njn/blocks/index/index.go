//go:build !exclude_pages_formats && !exclude_pages_format_njn

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
	"html/template"
	"math"
	"net/http"
	"net/url"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/go-corelibs/slices"
	clStrings "github.com/go-corelibs/strings"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/lang"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
	"github.com/go-enjin/be/pkg/pageql"
	"github.com/go-enjin/be/pkg/pages"
	"github.com/go-enjin/be/pkg/request/argv"
)

// TODO: SearchWithin is way too heavy for quoted.fyi, does not use kws

const (
	Tag feature.Tag = "njn-blocks-index"
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
}

func New() (field MakeBlock) {
	f := new(CBlock)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = Tag
	return f
}

func (f *CBlock) Init(this interface{}) {
	f.CEnjinBlock.Init(this)
}

func (f *CBlock) Make() Block {
	return f
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

	numPerPage, pageIndex, pageNumber := 10, 0, 1
	if v, ee := maps.ExtractIntValue("index-num-per-page", data); ee != nil {
		err = ee
		return
	} else if v > 0 {
		numPerPage = v
	}
	blockNumPerPage := numPerPage

	var indexViews []string
	var indexViewTitles []string
	if views, ok := data["index-views"].(string); ok {
		for _, view := range strings.Split(views, ";") {
			before, after, _ := strings.Cut(view, "=")
			kebab := strcase.ToKebab(strings.TrimSpace(before))
			indexViews = append(indexViews, kebab)
			if after == "" {
				indexViewTitles = append(indexViewTitles, clStrings.ToSpacedCamel(before))
			} else {
				indexViewTitles = append(indexViewTitles, strings.TrimSpace(after))
			}
		}
	} else {
		err = fmt.Errorf("index blocks require an index-view property set")
		return
	}

	filters := makeFilters(data)

	reqArgv := re.RequestArgv()
	defTag := f.Enjin.SiteDefaultLanguage()
	langMode := f.Enjin.SiteLanguageMode()

	var csqp bool // correct search query paths
	decArgv := argv.DecomposeHttpRequest(reqArgv.Request)
	decArgv.Language = reqArgv.Language
	for idx, args := range decArgv.Argv {
		for jdx, arg := range args {
			if escCheck := strings.HasPrefix(arg, "%28") && strings.HasSuffix(arg, "%29"); escCheck {
				csqp = true
				arg = arg[3 : len(arg)-3]
				if arg == "" {
					decArgv.Argv[idx][jdx] = ""
				} else {
					arg, _ = url.PathUnescape(arg)
					arg = html.UnescapeString(arg)
					decArgv.Argv[idx][jdx] = "(" + arg + ")"
				}
			} else if braceCheck := arg != "" && arg[0] == '(' && arg[len(arg)-1] == ')'; braceCheck {
				arg = arg[1 : len(arg)-1]
				if arg == "" {
					csqp = true
					decArgv.Argv[idx][jdx] = ""
				} else {
					arg, _ = url.PathUnescape(arg)
					arg = html.UnescapeString(arg)
					decArgv.Argv[idx][jdx] = "(" + arg + ")"
				}
			}
		}
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
			re.RequestContext().SetSpecific(argv.RequestConsumedKey, true)

			var fixArgs []string
			var viewArgs []string
			for _, piece := range pieces[1:] {
				if updated := filters.SetPresent(piece); updated {
					fixArgs = append(fixArgs, piece)
				} else if slices.Within(piece, indexViews) {
					if argvView == "" {
						argvView = piece
						viewArgs = append(viewArgs, piece)
					} else {
						fixArgs = append(fixArgs, piece)
					}
				} else if searchEnabled && piece != "" && piece[0] == '(' && piece[len(piece)-1] == ')' {
					argvSearch = piece[1 : len(piece)-1] // trim '(' and ')'
					if unescaped, eee := url.PathUnescape(argvSearch); eee == nil {
						fixArgs = append(fixArgs, "("+unescaped+")")
						argvSearch = unescaped
					} else {
						log.ErrorF("error unescaping argv search: %v", eee)
					}
				} else {
					// 	fixArgs = append(fixArgs, piece)
				}
			}
			fixArgs = append(viewArgs, fixArgs...)
			if len(fixArgs) != len(pieces[1:]) {
				reqArgv.Argv[idx] = append([]string{pieces[0]}, fixArgs...)
				redirect = langMode.ToUrl(defTag, reqArgv.Language, reqArgv.String())
				return
			}
		}
	}

	if argvBlockPresent {
		if len(decArgv.Argv) > 0 && len(decArgv.Argv[0]) > 0 {
			reqUrl := reqArgv.String()
			decUrl := decArgv.String()
			if _, untranslated, ok := lang.ParseLangPath(reqUrl); ok {
				reqUrl = untranslated
			}
			if _, untranslated, ok := lang.ParseLangPath(decUrl); ok {
				decUrl = untranslated
			}
			if csqp || reqUrl != decUrl /*reqArgv.String() != decArgv.String()*/ {
				redirect = langMode.ToUrl(defTag, reqArgv.Language, decArgv.String())
				return
			}
		}
		if reqArgv.NumPerPage > -1 {
			numPerPage = reqArgv.NumPerPage
		}
		if reqArgv.PageNumber > -1 {
			reqArgv.NumPerPage = numPerPage
			pageIndex = reqArgv.PageNumber
			pageNumber = pageIndex + 1
		} else {
			reqArgv.PageNumber = -1
			reqArgv.NumPerPage = -1
		}
	}

	block["View"] = argvView
	block["NumPerPage"] = numPerPage

	block["SearchEnabled"] = searchEnabled
	if searchEnabled {
		block["SearchQuery"] = argvSearch
		block["SearchNonce"] = f.Enjin.CreateNonce(searchNonceKey)
		if argvBlockPresent {
			if searchRedirect, searchError := f.handleSearchRedirect(tag, searchNonceKey, indexViews, reqArgv); searchError != nil {
				block["SearchError"] = searchError.Error()
			} else if searchRedirect != "" {
				redirect = searchRedirect // already correct lang mode
				return
			}
		}
	}

	var query string
	if query, ok = data["index-query"].(string); !ok {
		err = fmt.Errorf("index blocks require an index-query property set")
		return
	}

	if _, perr := pageql.CompileQuery(query); perr != nil {
		err = fmt.Errorf("query error:\n%v", perr.Pretty())
		return
	}

	found := f.Enjin.MatchQL(query)
	totalFound := len(found)
	found = filters.FilterPages(found)
	totalFiltered := len(found)

	if searchEnabled && argvSearch != "" {
		if len(found) == 0 {
			// nope
		} else if matched, searchResults, e := pages.SearchWithin(argvSearch, totalFiltered, 0, found, f.Enjin.SiteDefaultLanguage(), reqArgv.Language, f.Enjin.SiteLanguageMode(), f.Enjin.MustGetTheme()); e != nil {
			log.ErrorF("error searching within... %v", err)
			found = nil
		} else {
			block["SearchWithinTotal"] = totalFiltered
			block["SearchResults"] = searchResults

			var updated []feature.Page

			searchRanked := true
			if ranked, ok := data["search-ranked"]; ok {
				searchRanked = maps.ExtractBoolValue(ranked)
			}

			if searchRanked {
				// use the order of .Hits to sort
				for _, hit := range searchResults.Hits {
					if pg, ok := matched[hit.ID]; ok {
						updated = append(updated, pg)
					}
				}
			} else {
				// use the already present found order
				for _, pg := range found {
					for _, hit := range searchResults.Hits {
						if hitPg, hitOk := matched[hit.ID]; hitOk && pg.Url() == hitPg.Url() {
							updated = append(updated, pg)
							break
						}
					}
				}
			}

			totalFiltered = len(updated)
			block["SearchTotal"] = totalFiltered
			log.DebugF("search found: %d (of %d total) hits for query: %v", len(updated), len(found), argvSearch)
			found = updated
		}
	}

	totalPages := int(math.Ceil(float64(totalFiltered) / float64(numPerPage)))

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

	if pgntn == "none" {
		if reqArgv.PageNumber >= 0 || reqArgv.NumPerPage >= 0 {
			reqArgv.NumPerPage = -1
			reqArgv.PageNumber = -1
			redirect = langMode.ToUrl(defTag, reqArgv.Language, reqArgv.String()) + "#" + tag
			return
		}
	}

	if pageIndex > totalPages {
		reqArgv.PageNumber = totalPages - 1
		redirect = langMode.ToUrl(defTag, reqArgv.Language, reqArgv.String())
		return
	}

	if numPerPage > 0 && totalFiltered > 0 {
		start := pageIndex * numPerPage
		end := start + numPerPage
		if start < end && end < totalFiltered {
			found = found[start:end]
		} else {
			found = found[start:]
		}
	}

	if r, hasReqInstance := re.RequestContext().Get("R").(*http.Request); hasReqInstance {
		f.Enjin.ApplyPageContextUpdaters(r, found...)
	}

	block["Results"] = found
	block["TotalPages"] = totalPages
	block["TotalFound"] = totalFound
	block["TotalFiltered"] = totalFiltered

	var builtViews Views
	for idx, viewKey := range indexViews {
		viewFilters := filters.Copy()
		if searchEnabled && argvSearch != "" {
			viewFilters.UpdateUrls(tag, reqArgv.Path, viewKey, "("+argvSearch+")")
		} else {
			viewFilters.UpdateUrls(tag, reqArgv.Path, viewKey)
		}

		view := makeView(idx, viewKey, indexViewTitles[idx], viewFilters)
		if argvView == viewKey {
			view.Present = true
		}

		cra := reqArgv.Copy()
		args := []string{tag, view.Key}
		for _, groups := range viewFilters {
			for _, filter := range groups {
				if filter.Present {
					args = append(args, filter.Key)
				}
			}
		}
		if searchEnabled {
			sfa := cra.Copy()
			sfa.NumPerPage = -1
			sfa.PageNumber = -1
			sfa.Argv = append([][]string{}, args)
			view.SearchAction = sfa.String()
		}
		if searchEnabled && argvSearch != "" {
			args = append(args, "("+argvSearch+")")
		}
		cra.Argv = append([][]string{}, args)

		view.Paginate = pgntn
		view.NumPerPage = numPerPage

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
				pageCra.PageNumber = pageIndex
				pageCra.NumPerPage = numPerPage
				if pageIndex > 0 {
					pageCra.PageNumber = 0
					view.FirstPage = pageCra.String() + "#" + tag + "-" + viewKey
					pageCra.PageNumber = pageIndex - 1
					view.PrevPage = pageCra.String() + "#" + tag + "-" + viewKey
				}

				view.PageIndex = pageIndex
				view.PageNumber = pageNumber
				view.TotalPages = totalPages

				if pageIndex < totalPages-1 {
					pageCra.PageNumber = pageIndex + 1
					view.NextPage = pageCra.String() + "#" + tag + "-" + viewKey
					pageCra.PageNumber = totalPages - 1
					view.LastPage = pageCra.String() + "#" + tag + "-" + viewKey
				}
			}

		} // end numPerPage < totalFiltered

		view.Url = strings.TrimSuffix(reqArgv.Path, "/") + "/:" + tag + "," + viewKey + "#" + tag + "-" + viewKey
		builtViews = append(builtViews, view)
	}
	block["Views"] = builtViews

	log.DebugF("index block found %v (total=%v) pages with query: %v", totalFiltered, totalFound, query)

	if heading, ok := re.PrepareBlockHeader(blockDataContent); ok {
		block["Heading"] = heading
	}

	if footer, ok := re.PrepareBlockFooter(blockDataContent); ok {
		block["Footer"] = footer
	}

	block["SiteContext"] = re.RequestContext()
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
