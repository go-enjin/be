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

import (
	"context"
	"net/http"

	beContext "github.com/go-enjin/be/pkg/context"
)

type UserNotices []*UserNotice

type UserNotice struct {
	Type    string           `json:"type"`
	Dismiss bool             `json:"dismiss"`
	Summary string           `json:"summary"`
	Actions []UserNoticeLink `json:"actions"`
}

type UserNoticeLink struct {
	Text   string `json:"text"`
	Href   string `json:"href"`
	Target string `json:"target,omitempty"`
}

const (
	UserNoticesRequestKey beContext.RequestKey = "user-notices"
)

func GetUserNotices(r *http.Request) (notices UserNotices) {
	if v := r.Context().Value(UserNoticesRequestKey); v != nil {
		notices, _ = v.(UserNotices)
	}
	return
}

func AddUserNotices(r *http.Request, notes ...*UserNotice) (modified *http.Request) {
	notices := GetUserNotices(r)
	notices = append(notices, notes...)
	modified = r.Clone(context.WithValue(r.Context(), UserNoticesRequestKey, notices))
	return
}

func MakeInfoNotice(message string, dismiss bool, actions ...UserNoticeLink) (notice *UserNotice) {
	return &UserNotice{
		Type:    "info",
		Dismiss: dismiss,
		Summary: message,
		Actions: actions,
	}
}

func MakeWarnNotice(message string, dismiss bool, actions ...UserNoticeLink) (notice *UserNotice) {
	return &UserNotice{
		Type:    "warn",
		Dismiss: dismiss,
		Summary: message,
		Actions: actions,
	}
}

func MakeErrorNotice(message string, dismiss bool, actions ...UserNoticeLink) (notice *UserNotice) {
	return &UserNotice{
		Type:    "error",
		Dismiss: dismiss,
		Summary: message,
		Actions: actions,
	}
}