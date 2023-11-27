//go:build page_shortcodes || pages || all

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

package shortcodes

import (
	"bytes"

	scanner "github.com/go-enjin/go-stdlib-text-scanner"
	"github.com/urfave/cli/v2"

	beContext "github.com/go-enjin/be/pkg/context"
	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/log"
)

const Tag feature.Tag = "page-shortcodes"

var (
	_ Feature     = (*CFeature)(nil)
	_ MakeFeature = (*CFeature)(nil)
)

type ShortcodeFeature interface {
	LookupShortcode(name string) (shortcode Shortcode, ok bool)
	feature.PageShortcodeProcessor
}

type Feature interface {
	feature.Feature
	ShortcodeFeature
}

type MakeFeature interface {
	Make() Feature

	Defaults() MakeFeature

	Add(shortcodes ...Shortcode) MakeFeature
}

type CFeature struct {
	feature.CFeature

	known   map[string]Shortcode
	aliases map[string]string
}

func New() MakeFeature {
	return NewTagged(Tag)
}

func NewTagged(tag feature.Tag) MakeFeature {
	f := new(CFeature)
	f.Init(f)
	f.PackageTag = Tag
	f.FeatureTag = tag
	return f
}

func (f *CFeature) Init(this interface{}) {
	f.CFeature.Init(this)
	f.known = make(map[string]Shortcode)
	f.aliases = make(map[string]string)
}

func (f *CFeature) Add(shortcodes ...Shortcode) MakeFeature {
	for _, shortcode := range shortcodes {
		if shortcode.RenderFn == nil && shortcode.InlineFn == nil {
			log.DebugDF(1, "ignoring shortcode missing both .RenderFn or .InlineFn: %#+v", shortcode)
			continue
		}
		f.known[shortcode.Name] = shortcode
		for _, alias := range shortcode.Aliases {
			if _, present := f.known[alias]; !present {
				f.aliases[alias] = shortcode.Name
			}
		}
	}
	return f
}

func (f *CFeature) Defaults() MakeFeature {
	f.Add(
		BoldShortcode,
		ItalicShortcode,
		UnderlineShortcode,
		ColorShortcode,
		StrikethroughShortcode,
		SuperscriptShortcode,
		SubscriptShortcode,
		UrlShortcode,
		CodeShortcode,
		QuoteShortcode,
		ImageShortcode,
	)
	return f
}

func (f *CFeature) Make() Feature {
	return f
}

func (f *CFeature) Build(b feature.Buildable) (err error) {
	return
}

func (f *CFeature) Setup(enjin feature.Internals) {
	f.CFeature.Setup(enjin)

}

func (f *CFeature) Startup(ctx *cli.Context) (err error) {
	if err = f.CFeature.Startup(ctx); err != nil {
		return
	}
	return
}

func (f *CFeature) LookupShortcode(name string) (shortcode Shortcode, ok bool) {
	if shortcode, ok = f.known[name]; !ok {
		if actual, aliased := f.aliases[name]; aliased {
			shortcode, ok = f.known[actual]
		}
	}
	return
}

func (f *CFeature) TranslateShortcodes(content string, ctx beContext.Context) (modified string) {
	modified = f.ProcessShortcodes(content).Render(ctx)
	return
}

func (f *CFeature) ProcessShortcodes(input string) (nodes Nodes) {

	buf := bytes.NewBuffer([]byte(input))
	var scan scanner.Scanner
	scan.Filename = "content"
	scan.Init(buf)
	scan.Mode |= scanner.ScanInts
	scan.Mode |= scanner.ScanChars
	scan.Mode |= scanner.ScanIdents
	scan.Mode |= scanner.ScanFloats
	scan.Mode |= scanner.ScanStrings
	scan.Mode |= scanner.ScanRawStrings
	scan.Mode |= scanner.KeepComments
	scan.Whitespace ^= 1<<' ' | 1<<'\t' | 1<<'\n'
	scan.Error = func(s *scanner.Scanner, msg string) {
		// nop
	}

	var stream Nodes
	var current *Node

	keepAndResetCurrent := func() {
		if current != nil {
			stream = append(stream, current)
			current = nil
		}
	}

	keepAppendText := func(text string) {
		if text == "" {
			return
		}
		if current != nil {
			if current.Name == "" {
				current.Raw += text
				current.Content += text
			} else {
				keepAndResetCurrent()
				current = newNode(f, "", text)
			}
		} else {
			current = newNode(f, "", text)
		}
	}

	keepAppendTextAndResetCurrent := func(text string) {
		keepAppendText(text)
		keepAndResetCurrent()
	}

	for scan.Peek() != scanner.EOF {

		text, raw, maybeTag, maybeTagRaw, closing := slurpToNextTag(&scan)

		if maybeTag != "" {

			// possibly a tag
			if closing {
				// possibly a closing tag
				if ctRaw, ctName, ctOk := parseClosingTag(maybeTagRaw); ctOk {

					if _, ok := f.LookupShortcode(ctName); ok {
						// valid closing tag
						keepAppendTextAndResetCurrent(text)

						// append closing tag
						current = newNode(f, ctName, "")
						current.Raw = ctRaw
						current.isClosing = true
						keepAndResetCurrent()

					} else {
						// unknown closing tag
						keepAppendText(raw)
					}

				} else {
					// not a closing tag, retain as is, keeping text node current
					keepAppendText(raw)
				}

			} else {
				// possibly an opening tag
				if otRaw, otName, otAttrs, otOk := parseOpeningTag(maybeTagRaw); otOk {
					if sc, ok := f.LookupShortcode(otName); ok {
						// valid opening tag
						keepAppendTextAndResetCurrent(text)

						// keep the new opening tag
						current = newNode(f, otName, otRaw)
						current.Attributes = otAttrs
						current.Shortcode = &sc
						keepAndResetCurrent()
					} else {
						// unknown opening tag
						keepAppendText(raw)
					}

				} else {
					// not a valid opening tag
					keepAppendText(raw)
				}
			}

		} else {
			// not a tag at all
			keepAppendText(raw)
		}

	}

	if current != nil {
		stream = append(stream, current)
	}

	nodes = stream.Collapse()
	return
}
