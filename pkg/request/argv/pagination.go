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

package argv

type Pagination struct {
	// BasePath is the URL path to use when generating pagination links
	BasePath string `json:"BasePath"`
	// PageNumber is the current page number in human format (lists start at 1)
	PageNumber int `json:"PageNumber"`
	// NumPerPage is the number of results per page requested
	NumPerPage int `json:"NumPerPage"`
	// PageIndex is the current page number in computer format (lists start at 0)
	PageIndex int `json:"PageIndex"`
	// LastIndex is the last page number in computer format
	LastIndex int `json:"LastIndex"`
	// TotalItems is the total number of items available
	TotalItems int `json:"TotalItems"`
	// TotalPages is the total number of pagination pages available
	TotalPages int `json:"TotalPages"`
	// StartItem is the number of the first item being presented
	StartItem int `json:"StartItem"`
	// EndItem is the number of the last item being presented
	EndItem int `json:"EndItem"`
	// SearchQuery is an optional plain text string to be included with generated output
	SearchQuery string `json:"SearchQuery"`
}