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
	"net/http"
	"sync"

	"github.com/Shopify/gomail"
	"github.com/asaskevich/govalidator"
	"github.com/mrz1836/go-sanitize"
	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/maps"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/forms"
	"github.com/go-enjin/be/pkg/globals"
	"github.com/go-enjin/be/pkg/log"
)

const Tag feature.Tag = "drivers-email-gomail"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

var (
	DefaultRetries int = 5
)

type Feature interface {
	feature.Feature
	feature.EmailSender

	RetrySendEmail(retries int, r *http.Request, account string, message *gomail.Message) (err error)
}

type MakeFeature interface {
	SetDefaultRetries(value int) MakeFeature
	AddAccount(name string, cfg SmtpConfig) MakeFeature

	Make() Feature
}

type CFeature struct {
	feature.CFeature

	retries int

	accounts map[string]SmtpConfig

	m  *sync.RWMutex
	wg sync.WaitGroup
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	f.Construct(f)
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.accounts = make(map[string]SmtpConfig)
	f.m = &sync.RWMutex{}
	f.retries = DefaultRetries
}

// SetRetries overrides the DefaultRetries value for accounts configured with
// this feature instance.
//
// Negative values apply the DefaultRetries, zero means "do not retry after the
// first attempt".
func (f *CFeature) SetDefaultRetries(value int) MakeFeature {
	if value < 0 {
		value = DefaultRetries
	}
	f.retries = value
	return f
}

func (f *CFeature) AddAccount(key string, cfg SmtpConfig) MakeFeature {
	f.accounts[key] = cfg
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	if err = f.CFeature.Build(b); err != nil {
		return
	}

	tag := f.Tag().String()
	accountFlags := []cli.Flag{
		&cli.IntFlag{
			Name:     globals.MakeFlagName(tag, "retries"),
			Usage:    "specify the default retries value",
			Category: tag,
			Value:    f.retries,
			EnvVars:  globals.MakeFlagEnvKeys(tag, "retries"),
		},
	}

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
				Name:     globals.MakeFlagName(tag, key+"-email"),
				Usage:    "specify the email address sending from",
				Category: tag,
				Value:    f.accounts[key].Email,
				EnvVars:  globals.MakeFlagEnvKeys(tag, key+"-email"),
			},
			&cli.StringFlag{
				Name:     globals.MakeFlagName(tag, key+"-display"),
				Usage:    "specify the display name sending from",
				Category: tag,
				Value:    f.accounts[key].Display,
				EnvVars:  globals.MakeFlagEnvKeys(tag, key+"-display"),
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
			&cli.IntFlag{
				Name:     globals.MakeFlagName(tag, key+"-retries"),
				Usage:    "specify the retries value",
				Category: tag,
				Value:    f.accounts[key].Retries,
				EnvVars:  globals.MakeFlagEnvKeys(tag, key+"-retries"),
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
			if err = f.SendEmail(nil, account, message); err != nil {
				return
			}
			f.wg.Wait() // f.SendMail uses goroutines for fast-path optimization
			fmt.Printf("test email sent to: %s\n", recipient)
			return
		},
	})
	return
}

func startupCheck[T string | int](ctx *cli.Context, tag, key string) (value T, flagName string, present bool) {
	flagName = globals.MakeFlagName(tag, key)
	if present = ctx.IsSet(flagName); present {
		v := ctx.Value(flagName)
		switch t := v.(type) {
		case string:
			if t != "" {
				value, _ = v.(T)
				return
			}
		case int:
			if t > 0 {
				value, _ = v.(T)
				return
			}
		}
	}
	return
}

func startupMust[T string | int](ctx *cli.Context, tag, key string) (value T, err error) {
	var present bool
	var flagName string
	if value, flagName, present = startupCheck[T](ctx, tag, key); !present {
		err = fmt.Errorf("missing --" + flagName)
	}
	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	tag := f.KebabTag

	for _, key := range maps.SortedKeys(f.accounts) {
		account := f.accounts[key]

		// required
		if account.Host, err = startupMust[string](ctx, tag, key+"-host"); err != nil {
			return
		} else if account.Port, err = startupMust[int](ctx, tag, key+"-port"); err != nil {
			return
		} else if account.Username, err = startupMust[string](ctx, tag, key+"-username"); err != nil {
			return
		} else if account.Password, err = startupMust[string](ctx, tag, key+"-password"); err != nil {
			return
		}

		// optional
		if display, flagName, present := startupCheck[string](ctx, tag, key+"-display"); !present {
			// nop
		} else if sanitized := forms.StrictSanitize(display); sanitized != display {
			err = fmt.Errorf("invalid --" + flagName + " value")
			return
		}

		if email, flagName, present := startupCheck[string](ctx, tag, key+"-email"); !present {
			// nop
		} else if sanitized := sanitize.Email(email, false); sanitized != email {
			err = fmt.Errorf("invalid --" + flagName + " value")
			return
		}

		if retries, _, present := startupCheck[int](ctx, tag, key+"-retries"); present {
			account.Retries = retries
		} else {
			account.Retries = f.retries
		}

		f.accounts[key] = account
	}
	return
}

func (f *CFeature) Shutdown() {
	f.wg.Wait()
}

func (f *CFeature) HasEmailAccount(account string) (present bool) {
	f.m.RLock()
	defer f.m.RUnlock()
	_, present = f.accounts[account]
	return
}

func (f *CFeature) SendEmail(r *http.Request, account string, message *gomail.Message) (err error) {
	var ok bool
	var cfg SmtpConfig
	f.m.RLock()
	if cfg, ok = f.accounts[account]; !ok {
		f.m.RUnlock()
		err = fmt.Errorf("account not found")
		return
	}
	f.m.RUnlock()
	if v := message.GetHeader("To"); len(v) == 0 {
		err = fmt.Errorf("message is missing the recipient, please set the \"To\" header before calling .SendEmail")
		return
	}
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		if cfg.Display == "" {
			message.SetHeader("From", cfg.Email)
		} else {
			message.SetHeader("From", fmt.Sprintf("%s <%s>", cfg.Display, cfg.Email))
		}

		var try int
		var err error
		for try = 0; try < 5; try++ {
			log.DebugRF(r, "dialing and sending message from: %v, to: %v (%d tries)", cfg.Email, message.GetHeader("To"), try)
			if err = f.dialAndSend(cfg, message); err == nil {
				return
			}
		}

		log.ErrorRF(r, "failed five times to send email: %v", err)
	}()
	return
}

func (f *CFeature) dialAndSend(cfg SmtpConfig, message *gomail.Message) (err error) {

	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	if err = dialer.DialAndSend(message); err != nil {
		err = fmt.Errorf(
			"error dialing and sending message from: %v, to: %v - %v",
			cfg.Email,
			message.GetHeader("To"),
			err,
		)
	}

	return
}
