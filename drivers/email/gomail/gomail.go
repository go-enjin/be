//go:build driver_email_gomail || drivers_email || drivers || all

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

package gomail

import (
	"fmt"
	"sync"

	"github.com/Shopify/gomail"
	"github.com/asaskevich/govalidator"
	"github.com/urfave/cli/v2"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
	"github.com/go-enjin/be/pkg/maps"
)

const Tag feature.Tag = "drivers-email-gomail"

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

	AddAccount(name string, cfg SmtpConfig) MakeFeature
}

type CFeature struct {
	feature.CFeature

	accounts map[string]SmtpConfig

	sync.RWMutex
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.accounts = make(map[string]SmtpConfig)
}

func (f *CFeature) AddAccount(key string, cfg SmtpConfig) MakeFeature {
	f.accounts[key] = cfg
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	tag := f.Tag().String()
	var accountFlags []cli.Flag
	for _, key := range maps.SortedKeys(f.accounts) {
		accountFlags = append(accountFlags,
			&cli.StringFlag{
				Name:     globals.MakeFlagName(tag, key+"-host"),
				Usage:    "specify the hostname",
				Category: tag,
				Value:    f.accounts[key].Host,
				EnvVars:  globals.MakeFlagEnvKeys(tag, key+"-host"),
			},
			&cli.IntFlag{
				Name:     globals.MakeFlagName(tag, key+"-port"),
				Usage:    "specify the port number",
				Category: tag,
				Value:    f.accounts[key].Port,
				EnvVars:  globals.MakeFlagEnvKeys(tag, key+"-port"),
			},
			&cli.StringFlag{
				Name:     globals.MakeFlagName(tag, key+"-username"),
				Usage:    "specify the username",
				Category: tag,
				Value:    f.accounts[key].Username,
				EnvVars:  globals.MakeFlagEnvKeys(tag, key+"-username"),
			},
			&cli.StringFlag{
				Name:     globals.MakeFlagName(tag, key+"-password"),
				Usage:    "specify the password",
				Category: tag,
				// Value:    f.accounts[key].Password,
				EnvVars: globals.MakeFlagEnvKeys(tag, key+"-password"),
			},
			&cli.StringFlag{
				Name:     globals.MakeFlagName(tag, key+"-email"),
				Usage:    "specify the email address sending from",
				Category: tag,
				Value:    f.accounts[key].Email,
				EnvVars:  globals.MakeFlagEnvKeys(tag, key+"-email"),
			},
		)
	}
	b.AddFlags(accountFlags...)

	b.AddCommands(&cli.Command{
		Name:      "test-gomail-send",
		Usage:     "Send a test message from the email/gomail feature",
		ArgsUsage: globals.BinName + " send-test-email [options] <account-key> <recipient>",
		Flags:     accountFlags,
		Action: func(ctx *cli.Context) (err error) {
			if err = f.Startup(ctx); err != nil {
				return
			}
			argv := ctx.Args().Slice()
			if len(argv) != 2 {
				cli.ShowCommandHelpAndExit(ctx, "test-gomail-send", 1)
			}
			if len(f.accounts) == 0 {
				err = fmt.Errorf("please add at least one email sender account")
				return
			}
			account := argv[0]
			if _, ok := f.accounts[account]; !ok {
				err = fmt.Errorf("account must be one of: %v", maps.SortedKeys(f.accounts))
				return
			}
			var recipient string
			if recipient = argv[1]; recipient == "" {
				err = fmt.Errorf("missing recipient argument")
				return
			} else if !govalidator.IsEmail(recipient) {
				err = fmt.Errorf("not an email address: %v", recipient)
				return
			}
			message := gomail.NewMessage()
			message.SetHeader("To", recipient)
			message.SetHeader("Subject", "Test message")
			message.SetBody("text/plain", "This is a test of sending emails from the "+account+" account.")
			err = f.SendEmail(account, message)
			return
		},
	})
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	tag := f.Tag().String()
	check := func(key string) (v interface{}, err error) {
		if flagName := globals.MakeFlagName(tag, key); ctx.IsSet(flagName) {
			v = ctx.Value(flagName)
			switch t := v.(type) {
			case string:
				if t != "" {
					return
				}
			case int:
				if t > 0 {
					return
				}
			}
			err = fmt.Errorf("missing --" + flagName)
		}
		return
	}

	for _, key := range maps.SortedKeys(f.accounts) {
		account := f.accounts[key]
		var v interface{}

		if v, err = check(key + "-host"); err != nil {
			return
		} else if s, ok := v.(string); ok {
			account.Host = s
		}

		if v, err = check(key + "-port"); err != nil {
			return
		} else if i, ok := v.(int); ok {
			account.Port = i
		}

		if v, err = check(key + "-username"); err != nil {
			return
		} else if s, ok := v.(string); ok {
			account.Username = s
		}

		if v, err = check(key + "-password"); err != nil {
			return
		} else if s, ok := v.(string); ok {
			account.Password = s
		}

		if v, err = check(key + "-email"); err != nil {
			return
		} else if s, ok := v.(string); ok {
			account.Email = s
		}

		f.accounts[key] = account
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

func (f *CFeature) SendEmail(account string, message *gomail.Message) (err error) {
	f.RLock()
	defer f.RUnlock()
	if cfg, ok := f.accounts[account]; ok {
		if v := message.GetHeader("To"); len(v) == 0 {
			err = fmt.Errorf("gomail.Message missing recipient, please set the \"To\" header before calling SendEmail")
			return
		}
		message.SetHeader("From", cfg.Email)
		dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
		log.DebugF("dialing and sending message from: %v, to: %v", cfg.Email, message.GetHeader("To"))
		err = dialer.DialAndSend(message)
	} else {
		err = fmt.Errorf("account not found")
	}
	return
}