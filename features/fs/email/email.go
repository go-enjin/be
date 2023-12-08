//go:build fs_email || all

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

package email

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"path/filepath"
	"strings"
	textTemplate "text/template"

	"github.com/Shopify/gomail"
	"github.com/asaskevich/govalidator"
	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/feature/filesystem"
	"github.com/go-enjin/be/pkg/maps"
	bePath "github.com/go-enjin/be/pkg/path"
	"github.com/go-enjin/be/types/page/matter"
)

const Tag feature.Tag = "fs-email"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type Feature interface {
	filesystem.Feature[MakeFeature]
	feature.EmailProvider
}

type MakeFeature interface {
	filesystem.MakeFeature[MakeFeature]

	Make() Feature
}

type CFeature struct {
	filesystem.CFeature[MakeFeature]
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
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {

	tag := f.Tag().String()
	toFlag, tmplFlag, accountFlag := f.makeCliKeys()
	b.AddFlags(
		&cli.StringFlag{
			Name:     tmplFlag,
			Usage:    "specify template to send a test email",
			Category: tag,
		},
		&cli.StringFlag{
			Name:     toFlag,
			Usage:    "specify recipient of test email",
			Category: tag,
		},
		&cli.StringFlag{
			Name:     accountFlag,
			Usage:    "specify email account key",
			Category: tag,
		},
	)

	return
}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}

	toFlag, tmplFlag, accountFlag := f.makeCliKeys()
	if ctx.IsSet(toFlag) && ctx.IsSet(tmplFlag) && ctx.IsSet(accountFlag) {
		emailRecipient := ctx.String(toFlag)
		if !govalidator.IsEmail(emailRecipient) {
			err = fmt.Errorf("cannot send to invalid email address: %v", emailRecipient)
			return
		}
		emailTmpl := ctx.String(tmplFlag)
		emailAccount := ctx.String(accountFlag)
		bodyCtx := beContext.New()
		bodyCtx.Set("Recipient", emailRecipient)
		if p := f.Enjin.FindEmailAccount(emailAccount); p == nil {
			err = fmt.Errorf("email account not found: %v", emailAccount)
		} else if message, e := f.NewEmail(emailTmpl, bodyCtx); e != nil {
			err = fmt.Errorf("error making new email message: %v", e)
		} else {
			message.SetHeader("To", emailRecipient)
			message.SetHeader("Subject", f.Enjin.SiteName()+" test message")
			if err = p.SendEmail(nil, emailAccount, message); err != nil {
				err = fmt.Errorf("error sending email: %v", err)
			} else {
				err = fmt.Errorf("test email sent to: %v", emailRecipient)
			}
		}
		return
	}
	return
}

func (f *CFeature) Shutdown() {
	return
}

func (f *CFeature) ListTemplates() (names []string) {
	unique := make(map[string]struct{})
	for _, file := range f.MountPoints.ListFiles(".") {
		if strings.HasSuffix(file, ".tmpl") {
			file = bePath.Base(file)
			if _, present := unique[file]; present {
				unique[file] = struct{}{}
			}
		}
	}
	names = maps.SortedKeys(unique)
	return
}

func (f *CFeature) HasTemplate(name string) (present bool) {
	if present = f.Exists(name + ".html.tmpl"); present {
	} else if present = f.Exists(name + ".tmpl"); present {
	}
	return
}

func (f *CFeature) NewEmail(path string, bodyCtx beContext.Context) (message *gomail.Message, err error) {

	basename := bePath.CleanWithSlash(path)
	basename = strings.TrimSuffix(path, ".tmpl")
	basename = strings.TrimSuffix(path, ".html")

	textName := basename + ".tmpl"
	htmlName := basename + ".html.tmpl"

	pmText := f.Exists(textName)
	pmHtml := f.Exists(htmlName)

	if !pmText && !pmHtml {
		err = fmt.Errorf("email template with basename not found: %v", basename)
		return
	}

	var textBody, htmlBody string
	var textSubject, htmlSubject string
	var textMatter, htmlMatter beContext.Context
	if pmText {
		if textMatter, textBody, err = f.MakeEmailBody(textName, bodyCtx); err != nil {
			return
		}
		textSubject, _ = textMatter.Get("Subject").(string)
	}
	if pmHtml {
		if htmlMatter, htmlBody, err = f.MakeEmailBody(htmlName, bodyCtx); err != nil {
			return
		}
		htmlSubject, _ = htmlMatter.Get("Subject").(string)
	}

	message = gomail.NewMessage()

	if subject, ok := bodyCtx.Get("Subject").(string); ok {
		message.SetHeader("Subject", subject)
	} else if pmText && textSubject != "" {
		message.SetHeader("Subject", textSubject)
	} else if pmHtml && htmlSubject != "" {
		message.SetHeader("Subject", htmlSubject)
	}

	switch {
	case pmText && !pmHtml:
		message.SetBody("text/plain", textBody)
	case !pmText && pmHtml:
		message.SetBody("text/html", htmlBody)
	case pmText && pmHtml:
		message.SetBody("text/plain", textBody)
		message.AddAlternative("text/html", htmlBody)
	}

	attach := func(list []string) {
		for _, file := range list {
			if data, mime, e := f.Enjin.FindFile(file); e != nil {
				err = fmt.Errorf("error finding email attachment: %v - %v", file, e)
				return
			} else {
				bn := filepath.Base(file)
				message.AttachReader(bn, bytes.NewReader(data), gomail.SetHeader(map[string][]string{
					"Content-Type": {mime},
				}))
			}
		}
	}

	if textMatter != nil {
		if list, ok := textMatter.Get("attachments").([]string); ok {
			attach(list)
		}
	}

	if htmlMatter != nil {
		if list, ok := htmlMatter.Get("attachments").([]string); ok {
			attach(list)
		}
	}

	return
}

func (f *CFeature) MakeEmailBody(path string, ctx beContext.Context) (fm beContext.Context, body string, err error) {

	var pm *matter.PageMatter
	if pm, err = f.FindReadPageMatter(path); err != nil {
		err = fmt.Errorf("error finding email page matter: %v - %v", path, err)
		return
	}
	fm = pm.Matter

	render := beContext.New()
	render.Apply(pm.Matter)
	render.Apply(ctx)
	render.Apply(f.Enjin.Context(nil))

	if strings.HasSuffix(path, ".html.tmpl") {
		var t *htmlTemplate.Template
		if t, err = htmlTemplate.
			New(strcase.ToKebab(path)).
			Funcs(f.Enjin.MakeFuncMap(nil).AsHTML()).
			Parse(pm.Body); err != nil {
			err = fmt.Errorf("error parsing html.Template: %v", err)
			return
		}
		var buf bytes.Buffer
		if err = t.Execute(&buf, render); err != nil {
			err = fmt.Errorf("error executing html.Template: %v", err)
			return
		}
		body = buf.String()
		return
	}

	var t *textTemplate.Template
	if t, err = textTemplate.New(strcase.ToKebab(path)).Funcs(f.Enjin.MakeFuncMap(ctx).AsTEXT()).Parse(pm.Body); err != nil {
		err = fmt.Errorf("error parsing text.Template: %v", err)
		return
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, render); err != nil {
		err = fmt.Errorf("error executing text.Template: %v", err)
		return
	}
	body = buf.String()
	return
}

func (f *CFeature) makeCliKeys() (to, tmpl, account string) {
	kebabTag := f.Tag().Kebab()
	to = "test-" + kebabTag + "-to"
	tmpl = "test-" + kebabTag + "-tmpl"
	account = "test-" + kebabTag + "-account"
	return
}