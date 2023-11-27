// Copyright (c) 2023  The Go-Enjin Authors
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
	"sort"
	"time"
)

/*
	Note: the types within this file have been copied from the bleve project and
	simplified for generic use cases where bleve is not built-in.

	See: "github.com/blevesearch/bleve/v2"

	The idea is to be able to provide bleve as a generic driver rather than the
	only means of full-text-search. This feature design pattern has not
	stabilized and this file is here for internal development

	Both Go-Enjin and bleve share the same Apache 2.0 licensing.
*/

type IndexErrMap map[string]error

type HighlightRequest struct {
	Style  *string  `json:"style"`
	Fields []string `json:"fields"`
}

// type FacetRequest struct {
//     Size           int              `json:"size"`
//     Field string                    `json:"field"`
//     NumericRanges  []*numericRange  `json:"numeric_ranges,omitempty"`
//     DateTimeRanges []*dateTimeRange `json:"date_ranges,omitempty"`
// }

// type FacetsRequest map[string]*FacetRequest

type SearchStatus struct {
	Total      int         `json:"total"`
	Failed     int         `json:"failed"`
	Successful int         `json:"successful"`
	Errors     IndexErrMap `json:"errors,omitempty"`
}

type SearchRequest struct {
	// Query            query.Query       `json:"query"`
	Size      int               `json:"size"`
	From      int               `json:"from"`
	Highlight *HighlightRequest `json:"highlight"`
	Fields    []string          `json:"fields"`
	// Facets           FacetsRequest     `json:"facets"`
	Explain bool `json:"explain"`
	// Sort             search.SortOrder  `json:"sort"`
	IncludeLocations bool     `json:"includeLocations"`
	Score            string   `json:"score,omitempty"`
	SearchAfter      []string `json:"search_after"`
	SearchBefore     []string `json:"search_before"`
	sortFunc         func(sort.Interface)
}

type SearchResult struct {
	Status    *SearchStatus           `json:"status"`
	Request   *SearchRequest          `json:"request"`
	Hits      DocumentMatchCollection `json:"hits"`
	Total     uint64                  `json:"total_hits"`
	BytesRead uint64                  `json:"bytesRead,omitempty"`
	MaxScore  float64                 `json:"max_score"`
	Took      time.Duration           `json:"took"`
	Facets    FacetResults            `json:"facets"`
}

type FacetResults map[string]*FacetResult

type FacetResult struct {
	Field   string `json:"field"`
	Total   int    `json:"total"`
	Missing int    `json:"missing"`
	Other   int    `json:"other"`
	// Terms         *TermFacets        `json:"terms,omitempty"`
	// NumericRanges NumericRangeFacets `json:"numeric_ranges,omitempty"`
	// DateRanges    DateRangeFacets    `json:"date_ranges,omitempty"`
}

type DocumentMatchCollection []*DocumentMatch

type DocumentMatch struct {
	Index string `json:"index,omitempty"`
	ID    string `json:"id"`
	// IndexInternalID    index.IndexInternalID  `json:"-"`
	Score              float64                `json:"score"`
	Expl               *Explanation           `json:"explanation,omitempty"`
	Locations          FieldTermLocationMap   `json:"locations,omitempty"`
	Fragments          FieldFragmentMap       `json:"fragments,omitempty"`
	Sort               []string               `json:"sort,omitempty"`
	Fields             map[string]interface{} `json:"fields,omitempty"`
	HitNumber          uint64                 `json:"-"`
	FieldTermLocations []FieldTermLocation    `json:"-"`
}

type Explanation struct {
	Value    float64        `json:"value"`
	Message  string         `json:"message"`
	Children []*Explanation `json:"children,omitempty"`
}

type FieldTermLocation struct {
	Field    string
	Term     string
	Location Location
}

type FieldFragmentMap map[string][]string

type FieldTermLocationMap map[string]TermLocationMap

type TermLocationMap map[string]Locations

type Locations []*Location

type Location struct {
	Pos            uint64         `json:"pos"`
	Start          uint64         `json:"start"`
	End            uint64         `json:"end"`
	ArrayPositions ArrayPositions `json:"array_positions"`
}

type ArrayPositions []uint64
