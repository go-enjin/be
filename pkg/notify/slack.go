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

package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

var RxSlackWebhook = regexp.MustCompile(`^\s*(https://hooks.slack.com/services/[A-Z0-9]{9}/[A-Z0-9]{11}/[a-zA-Z0-9]{24})\s*$`)

var RxSlackChannel = regexp.MustCompile(`^\s*([A-Z0-9]{9}/[A-Z0-9]{11}/[a-zA-Z0-9]{24})\s*$`)

func SlackUrl(channel string) (webhook string) {
	if RxSlackWebhook.MatchString(channel) {
		webhook = channel
	} else if RxSlackChannel.MatchString(channel) {
		webhook = fmt.Sprintf("https://hooks.slack.com/services/%s", channel)
	}
	return
}

func SlackF(slack, format string, argv ...interface{}) (err error) {
	if slack == "" {
		// nop
		return
	}
	channel := SlackUrl(slack)
	if channel == "" {
		err = fmt.Errorf("invalid slack channel (or url): %v", slack)
		return
	}
	data := map[string]string{
		"text": fmt.Sprintf(format, argv...),
	}
	body, _ := json.Marshal(data)
	bb := bytes.NewBuffer(body)
	mt := "application/json"
	_, err = http.Post(channel, mt, bb)
	return
}
