//go:build driver_email_fakemail || drivers_email || drivers || all

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

package fakemail

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/Shopify/gomail"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-corelibs/strings"
)

const Tag feature.Tag = "drivers-email-fakemail"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	feature.Feature
	feature.EmailSender
}

type MakeFeature interface {
	Make() Feature

	AddAccount(name string) MakeFeature
}

type CFeature struct {
	feature.CFeature

	accounts map[string]struct{}

	sync.RWMutex
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.CFeature.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.accounts = make(map[string]struct{})
}

func (f *CFeature) AddAccount(key string) MakeFeature {
	f.accounts[key] = struct{}{}
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	err = f.CFeature.Build(b)
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) HasEmailAccount(account string) (present bool) {
	f.RLock()
	defer f.RUnlock()
	_, present = f.accounts[account]
	return
}

func (f *CFeature) SendEmail(r *http.Request, account string, message *gomail.Message) (err error) {
	f.RLock()
	defer f.RUnlock()
	if _, ok := f.accounts[account]; ok {
		if v := message.GetHeader("To"); len(v) == 0 {
			err = fmt.Errorf("gomail.Message missing recipient, please set the \"To\" header before calling SendEmail")
			return
		}
		buf := strings.NewByteBuffer()
		_, _ = message.WriteTo(buf)
		log.WarnRF(r, "fakemail should have sent the following email:\n# BEGIN EMAIL\n%v\n# END EMAIL", buf.String())
	} else {
		err = fmt.Errorf("account not found")
	}
	return
}
