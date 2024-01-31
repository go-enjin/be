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

package mime

import (
	goMime "mime"

	"github.com/gabriel-vasile/mimetype"

	clPath "github.com/go-corelibs/path"
)

const (
	TextMimeType       = "text/plain"
	HtmlMimeType       = "text/html"
	CssMimeType        = "text/css"
	ScssMimeType       = "text/x-scss"
	JsonMimeType       = "application/json"
	JavaScriptMimeType = "text/javascript"

	DirectoryMimeType = "inode/directory"
	BinaryMimeType    = "application/octet-stream"

	EnjinMimeType     = "text/enjin"
	EnjinExtension    = "njn"
	OrgModeMimeType   = "text/org-mode"
	OrgModeExtension  = "org"
	MarkdownMimeType  = "text/markdown"
	MarkdownExtension = "md"
)

var (
	ExtensionLookup = map[string]string{
		"txt":  TextMimeType + "; charset=utf-8",
		"html": HtmlMimeType + "; charset=utf-8",
		"css":  CssMimeType + "; charset=utf-8",
		"scss": ScssMimeType + "; charset=utf-8",
		"json": JsonMimeType + "; charset=utf-8",
		"js":   JavaScriptMimeType + "; charset=utf-8",
	}
	PlainTextLookup = map[string]string{
		TextMimeType:       "utf-8",
		HtmlMimeType:       "utf-8",
		CssMimeType:        "utf-8",
		ScssMimeType:       "utf-8",
		JsonMimeType:       "utf-8",
		JavaScriptMimeType: "utf-8",
		EnjinMimeType:      "utf-8",
		OrgModeMimeType:    "utf-8",
		MarkdownMimeType:   "utf-8",
	}
)

func init() {
	_ = RegisterTextType(OrgModeMimeType, OrgModeExtension, nil)
	// support text/enjin as a text/plain mime type
	_ = RegisterTextType(EnjinMimeType, EnjinExtension, /*func(raw []byte, limit uint32) bool {
			_, content, _ := beMatter.ParseContent(string(raw))
			var ctx []map[string]interface{}
			if err := json.Unmarshal([]byte(content), &ctx); err == nil {
				return true
			}
			return false
		}*/nil)
	// support text/markdown as a text/plain mime type
	_ = RegisterTextType(MarkdownMimeType, MarkdownExtension, /*func(raw []byte, limit uint32) bool {
			ast := markdown.Parse(raw, nil)
			return ast != nil
		}*/nil)
}

func RegisterTextType(mimeType, extension string, detector func(raw []byte, limit uint32) bool) (err error) {
	if m, p, e := goMime.ParseMediaType(mimeType); e != nil {
		err = e
		return
	} else if _, ok := p["charset"]; !ok {
		mimeType = m + "; charset=utf-8"
	}
	ExtensionLookup[extension] = mimeType
	if detector != nil {
		mimetype.Lookup(TextMimeType).Extend(detector, mimeType, extension)
	}
	err = goMime.AddExtensionType(extension, mimeType)
	return
}

func PruneCharset(mimeType string) (pruned string) {
	pruned, _, _ = goMime.ParseMediaType(mimeType)
	return
}

func IsPlainText(mime string) (yes bool) {
	mime = PruneCharset(mime)
	if _, yes = PlainTextLookup[mime]; yes {
		return
	}
	if mt := mimetype.Lookup(mime); mt != nil {
		if yes = mt.Is(TextMimeType); yes {
			return
		}
		for check := mt; check != nil; check = check.Parent() {
			if yes = check.Is(TextMimeType); yes {
				return
			}
		}
	}
	return
}

func FromExtension(extension string) (mime string, ok bool) {
	if mime, ok = ExtensionLookup[extension]; ok {
		return
	}
	mime = goMime.TypeByExtension(extension)
	ok = mime != ""
	return
}

func FromPathOnly(path string) (mime string) {
	if path == "" {
		return
	} else if a, b := clPath.ExtExt(path); b != "" && a == "tmpl" {
		mime, _ = FromExtension(b)
	} else if a != "" {
		mime, _ = FromExtension(a)
	}
	return
}

// Mime returns the MIME type string of a local file
func Mime(path string) (mime string) {
	if clPath.IsDir(path) {
		mime = DirectoryMimeType
		return
	} else if !clPath.IsFile(path) {
		return
	} else if mime = FromPathOnly(path); mime != "" {
		return
	} else if mt, err := mimetype.DetectFile(path); err == nil {
		mime = mt.String()
	}
	return
}
