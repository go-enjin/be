// Copyright (c) 2025  The Go-Enjin Authors
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

package logger

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/go-corelibs/context"

	"github.com/go-enjin/be/pkg/feature"
	"github.com/go-enjin/be/pkg/request"
)

// This code was extracted from pkg/net/gorilla-handlers, which in turn was forked from github.com/gorilla/handlers

const lowerhex = "0123456789abcdef"

func appendQuoted(buf []byte, s string) []byte {
	var runeTmp [utf8.UTFMax]byte
	for width := 0; len(s) > 0; s = s[width:] {
		r := rune(s[0])
		width = 1
		if r >= utf8.RuneSelf {
			r, width = utf8.DecodeRuneInString(s)
		}
		if width == 1 && r == utf8.RuneError {
			buf = append(buf, `\x`...)
			buf = append(buf, lowerhex[s[0]>>4])
			buf = append(buf, lowerhex[s[0]&0xF])
			continue
		}
		if r == rune('"') || r == '\\' { // always backslashed
			buf = append(buf, '\\')
			buf = append(buf, byte(r))
			continue
		}
		if strconv.IsPrint(r) {
			n := utf8.EncodeRune(runeTmp[:], r)
			buf = append(buf, runeTmp[:n]...)
			continue
		}
		switch r {
		case '\a':
			buf = append(buf, `\a`...)
		case '\b':
			buf = append(buf, `\b`...)
		case '\f':
			buf = append(buf, `\f`...)
		case '\n':
			buf = append(buf, `\n`...)
		case '\r':
			buf = append(buf, `\r`...)
		case '\t':
			buf = append(buf, `\t`...)
		case '\v':
			buf = append(buf, `\v`...)
		default:
			switch {
			case r < ' ':
				buf = append(buf, `\x`...)
				buf = append(buf, lowerhex[s[0]>>4])
				buf = append(buf, lowerhex[s[0]&0xF])
			case r > utf8.MaxRune:
				r = 0xFFFD
				fallthrough
			case r < 0x10000:
				buf = append(buf, `\u`...)
				for s := 12; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			default:
				buf = append(buf, `\U`...)
				for s := 28; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			}
		}
	}
	return buf
}

func prepareCommonLog(req *http.Request, url *url.URL, ts time.Time, status int, size int, duration time.Duration) (ctx context.Context) {
	ctx = context.Context{
		"rid":      "nil",
		"username": "-",
		"enjin-id": "-",
	}

	if url.User != nil {
		if name := url.User.Username(); name != "" {
			ctx["username"] = name
		}
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}
	ctx["host"] = host

	uri := req.RequestURI

	// Requests using the CONNECT method over HTTP/2.0 must use
	// the authority field (aka r.Host) to identify the target.
	// Refer: https://httpwg.github.io/specs/rfc7540.html#CONNECT
	if req.ProtoMajor == 2 && req.Method == "CONNECT" {
		uri = req.Host
	}
	if uri == "" {
		uri = url.RequestURI()
	}
	ctx["uri"] = uri

	if rid := request.GetRequestID(req); rid != "" {
		ctx["rid"] = rid
	}

	ctx["datetime"] = ts.Format("02/Jan/2006:15:04:05 -0700")

	ctx["req-host"] = req.Host
	if v := request.GetEnjinID(req); v != "" {
		ctx["enjin-id"] = v
	}

	if username, _, ok := req.BasicAuth(); ok {
		ctx["basic-auth"] = username
	}

	ctx["method"] = req.Method
	ctx["proto"] = req.Proto
	ctx["status"] = strconv.Itoa(status)
	ctx["size"] = strconv.Itoa(size)
	ctx["duration"] = duration

	if req.TLS != nil {
		ctx["scheme"] = "https"
	} else {
		ctx["scheme"] = "http"
	}

	return
}

func buildCommonLogLine(ctx context.Context) []byte {
	buf := make([]byte, 0)
	buf = append(buf, '[')
	buf = append(buf, ctx.String("rid")...)
	buf = append(buf, ']', ' ')
	buf = append(buf, ctx.String("host")...)
	buf = append(buf, " - "...)
	buf = append(buf, ctx.String("username")...)
	buf = append(buf, " ["...)
	buf = append(buf, ctx.String("datetime")...)
	buf = append(buf, `] `...)
	buf = append(buf, `- `...)
	buf = append(buf, ctx.String("req-host")...)
	buf = append(buf, fmt.Sprintf(" [%v]", ctx.String("enjin-id"))...)
	buf = append(buf, fmt.Sprintf(" [%v]", ctx.String("basic-auth"))...)
	buf = append(buf, ` "`...)
	buf = append(buf, ctx.String("method")...)
	buf = append(buf, " "...)
	buf = appendQuoted(buf, ctx.String("uri"))
	buf = append(buf, " "...)
	buf = append(buf, ctx.String("proto")...)
	buf = append(buf, `" `...)
	buf = append(buf, ctx.String("status")...)
	buf = append(buf, " "...)
	buf = append(buf, ctx.String("size")...)
	buf = append(buf, " "...)
	buf = append(buf, ctx.TimeDuration("duration").String()...)

	if ctx.Has("referer") {
		buf = append(buf, ` "`...)
		buf = appendQuoted(buf, ctx.String("referer"))
		buf = append(buf, `"`...)
	}

	if ctx.Has("user-agent") {
		buf = append(buf, ` "`...)
		buf = appendQuoted(buf, ctx.String("user-agent"))
		buf = append(buf, `"`...)
	}

	return buf
}

func buildCommonJsonLine(ctx context.Context) []byte {
	buf := make([]byte, 0)
	buf = append(buf, `{"rid":"`...)
	buf = append(buf, ctx.String("rid")...)
	buf = append(buf, `","remote":"`...)
	buf = append(buf, ctx.String("host")...)
	buf = append(buf, `","username":"`...)
	buf = append(buf, ctx.String("username")...)
	buf = append(buf, `","datetime":"`...)
	buf = append(buf, ctx.String("datetime")...)
	buf = append(buf, `","domain":"`...)
	buf = append(buf, ctx.String("req-host")...)
	buf = append(buf, `","enjin-id":"`...)
	buf = append(buf, ctx.String("enjin-id")...)
	buf = append(buf, `","basic-auth":"`...)
	buf = append(buf, ctx.String("basic-auth")...)
	buf = append(buf, `","scheme":"`...)
	buf = append(buf, ctx.String("scheme")...)
	buf = append(buf, `","method":"`...)
	buf = append(buf, ctx.String("method")...)
	buf = append(buf, `","path":"`...)
	buf = appendQuoted(buf, ctx.String("uri"))
	buf = append(buf, `","proto":"`...)
	buf = append(buf, ctx.String("proto")...)
	buf = append(buf, `","status":`...)
	buf = append(buf, ctx.String("status")...)
	buf = append(buf, `,"size":`...)
	buf = append(buf, ctx.String("size")...)
	buf = append(buf, `,"duration":"`...)
	buf = append(buf, ctx.TimeDuration("duration").String()...)
	buf = append(buf, `"`...)

	if ctx.Has("user-agent") {
		buf = append(buf, `,"useragent":"`...)
		buf = append(buf, ctx.String("user-agent")...)
		buf = append(buf, `"`...)
	}
	if ctx.Has("referer") {
		buf = append(buf, `,"referer":"`...)
		buf = append(buf, ctx.String("referer")...)
		buf = append(buf, `"`...)
	}

	buf = append(buf, `}`...)
	return buf
}

func writeLog(writer io.Writer, params feature.LoggerContext) {
	ctx := prepareCommonLog(params.Request(), params.URL(), params.TimeStamp(), params.StatusCode(), params.Size(), params.Duration())
	buf := buildCommonLogLine(ctx)
	buf = append(buf, '\n')
	_, _ = writer.Write(buf)
}

// writeCombinedLog writes a log entry for req to w in Apache Combined Log Format.
// ts is the timestamp with which the entry should be logged.
// status and size are used to provide the response HTTP status and size.
func writeCombinedLog(writer io.Writer, params feature.LoggerContext) {
	ctx := prepareCommonLog(params.Request(), params.URL(), params.TimeStamp(), params.StatusCode(), params.Size(), params.Duration())
	ctx["referer"] = params.Request().Referer()
	ctx["user-agent"] = params.Request().UserAgent()

	buf := buildCommonLogLine(ctx)
	buf = append(buf, ` "`...)
	buf = appendQuoted(buf, params.Request().Referer())
	buf = append(buf, `" "`...)
	buf = appendQuoted(buf, params.Request().UserAgent())
	buf = append(buf, '"', '\n')
	_, _ = writer.Write(buf)
}

func writeJsonLog(writer io.Writer, params feature.LoggerContext) {
	ctx := prepareCommonLog(params.Request(), params.URL(), params.TimeStamp(), params.StatusCode(), params.Size(), params.Duration())
	buf := buildCommonJsonLine(ctx)
	buf = append(buf, '\n')
	_, _ = writer.Write(buf)
}

func writeCombinedJsonLog(writer io.Writer, params feature.LoggerContext) {
	ctx := prepareCommonLog(params.Request(), params.URL(), params.TimeStamp(), params.StatusCode(), params.Size(), params.Duration())
	ctx["referer"] = params.Request().Referer()
	ctx["user-agent"] = params.Request().UserAgent()
	buf := buildCommonJsonLine(ctx)
	buf = append(buf, '\n')
	_, _ = writer.Write(buf)
}
