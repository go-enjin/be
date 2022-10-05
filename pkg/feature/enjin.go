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

package feature

import beStrings "github.com/go-enjin/be/pkg/strings"

type NjnClass uint

const (
	InlineNjnClass NjnClass = iota
	ContainerNjnClass
	AnyNjnClass
)

func (nc NjnClass) String() string {
	switch nc {
	case InlineNjnClass:
		return "inline"
	case ContainerNjnClass:
		return "container"
	case AnyNjnClass:
		return "any"
	}
	return "error"
}

type EnjinFeature interface {
	Feature

	NjnClass() (tagClass NjnClass)
	NjnCheckTag(tagName string) (allow bool)
	NjnCheckClass(tagClass NjnClass) (allow bool)
	NjnTagsAllowed() (allowed []string, ok bool)
	NjnTagsDenied() (denied []string, ok bool)
	NjnClassAllowed() (allowed NjnClass, ok bool)
}

type CEnjinFeature struct {
	CFeature
}

func (f *CEnjinFeature) NjnCheckTag(tagName string) (allow bool) {
	allowed, checkAllowed := f.NjnTagsAllowed()
	denied, checkDenied := f.NjnTagsDenied()
	switch {
	case checkAllowed && checkDenied:
		allow = !beStrings.StringInStrings(tagName, denied...) && beStrings.StringInStrings(tagName, allowed...)
	case checkAllowed:
		allow = beStrings.StringInStrings(tagName, allowed...)
	case checkDenied:
		allow = !beStrings.StringInStrings(tagName, denied...)
	default:
		allow = true
	}
	return
}

func (f *CEnjinFeature) NjnCheckClass(tagClass NjnClass) (allow bool) {
	if allowed, checkAllowed := f.NjnClassAllowed(); checkAllowed {
		switch allowed {
		case AnyNjnClass:
			allow = true
		default:
			allow = tagClass == allowed
		}
	} else {
		allow = true
	}
	return
}

func (f *CEnjinFeature) NjnTagsAllowed() (allowed []string, ok bool) {
	return
}

func (f *CEnjinFeature) NjnTagsDenied() (denied []string, ok bool) {
	return
}

func (f *CEnjinFeature) NjnClassAllowed() (allowed NjnClass, ok bool) {
	return
}