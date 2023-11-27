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

package feature

type CSiteFeatureInfo struct {
	// Tag is the kebab-cased feature.Tag
	Tag string `json:"tag"`
	// Key is a different kebab-cased tag than Tag
	Key string `json:"key"`
	// Icon is a font-awesome icon class
	Icon string `json:"icon"`
	// Label is a translated string label
	Label string `json:"label"`
	// Usage is a sentence or two describing how to use this feature
	Usage string `json:"usage"`
	// Hint is used in multiple-step authentications, typically as a button's title attribute
	Hint string `json:"hint"`
	// Placeholder is the string of text displayed in the challenge form input tag
	Placeholder string `json:"placeholder"`
	// Backup indicates whether this feature is a backup feature or not
	Backup bool `json:"backup"`
}

func NewSiteFeatureInfo(tag, key, icon, label string) (info *CSiteFeatureInfo) {
	info = &CSiteFeatureInfo{
		Tag:   tag,
		Key:   key,
		Icon:  icon,
		Label: label,
	}
	return
}

type CSiteAuthMultiFactorInfo struct {
	CSiteFeatureInfo
	// Ready is true when the current user has at least one provisioned factor of this type
	Ready bool `json:"ready"`
	// Factors is the list of provisioned factors, of this type, ready for challenging the current user
	Factors []string `json:"factors"`
	// Claimed is the list of provisioned factors which are present in the current user's claims
	Claimed []string `json:"claimed"`
	// Submit indicates whether to submit the feature.Tag in order to initiate the challenge process
	Submit bool `json:"submit"`
}

func NewSiteAuthMultiFactorInfo(tag, key, icon, label string, factors ...string) (info *CSiteAuthMultiFactorInfo) {
	info = &CSiteAuthMultiFactorInfo{
		CSiteFeatureInfo: CSiteFeatureInfo{
			Tag:   tag,
			Key:   key,
			Icon:  icon,
			Label: label,
		},
		Ready:   len(factors) > 0,
		Factors: factors,
	}
	return
}
