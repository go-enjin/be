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

	"github.com/go-enjin/be/pkg/request"
)

type UserNotices []*UserNotice

type UserNotice struct {
	Type    string           `json:"type"`
	Dismiss bool             `json:"dismiss"`
	Summary string           `json:"summary"`
	Content []string         `json:"content"`
	Actions []UserNoticeLink `json:"actions"`
}

type UserNoticeLink struct {
	Text   string `json:"text"`
	Href   string `json:"href"`
	Target string `json:"target,omitempty"`
}

const (
	UserNoticesRequestKey request.Key = "user-notices"
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

func FilterUserNotices(r *http.Request, fn func(notice *UserNotice) (keep bool)) (m *http.Request) {
	var keepers UserNotices
	notices, _ := r.Context().Value(UserNoticesRequestKey).(UserNotices)
	if len(notices) > 0 {
		for _, notice := range notices {
			if fn(notice) {
				keepers = append(keepers, notice)
			}
		}
	}
	m = r
	return
}

func NewNotice(key string, dismiss bool, summary string, actions ...UserNoticeLink) (notice *UserNotice) {
	if key == "" {
		key = "info"
	}
	return &UserNotice{
		Type:    key,
		Dismiss: dismiss,
		Summary: summary,
		Actions: actions,
	}
}

func parseNoticeArgv(argv []interface{}) (content []string, actions []UserNoticeLink) {
	for _, arg := range argv {
		switch t := arg.(type) {
		case string:
			content = append(content, t)
		case UserNoticeLink:
			actions = append(actions, t)
		}
	}
	return
}

func MakeImportantNotice(dismiss bool, message string, argv ...interface{}) (notice *UserNotice) {
	content, actions := parseNoticeArgv(argv)
	notice = NewNotice("important", dismiss, message, actions...)
	notice.Content = content
	return
}

func AddImportantNotice(r *http.Request, dismiss bool, message string, argv ...interface{}) (modified *http.Request) {
	modified = AddUserNotices(r, MakeImportantNotice(dismiss, message, argv...))
	return
}

func MakeInfoNotice(dismiss bool, message string, argv ...interface{}) (notice *UserNotice) {
	content, actions := parseNoticeArgv(argv)
	notice = NewNotice("info", dismiss, message, actions...)
	notice.Content = content
	return
}

func AddInfoNotice(r *http.Request, dismiss bool, message string, argv ...interface{}) (modified *http.Request) {
	modified = AddUserNotices(r, MakeInfoNotice(dismiss, message, argv...))
	return
}

func MakeWarnNotice(dismiss bool, message string, argv ...interface{}) (notice *UserNotice) {
	content, actions := parseNoticeArgv(argv)
	notice = NewNotice("warn", dismiss, message, actions...)
	notice.Content = content
	return
}

func AddWarnNotice(r *http.Request, dismiss bool, message string, argv ...interface{}) (modified *http.Request) {
	modified = AddUserNotices(r, MakeWarnNotice(dismiss, message, argv...))
	return
}

func MakeErrorNotice(dismiss bool, message string, argv ...interface{}) (notice *UserNotice) {
	content, actions := parseNoticeArgv(argv)
	notice = NewNotice("error", dismiss, message, actions...)
	notice.Content = content
	return
}

func AddErrorNotice(r *http.Request, dismiss bool, message string, argv ...interface{}) (modified *http.Request) {
	modified = AddUserNotices(r, MakeErrorNotice(dismiss, message, argv...))
	return
}