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

package log

import "net/http"

func ErrorR(r *http.Request, err error) {
	Config.ErrorRDF(r, 1, "%v", err)
}

func ErrorRF(r *http.Request, format string, argv ...interface{}) {
	Config.ErrorRDF(r, 1, format, argv...)
}

func ErrorRDF(r *http.Request, depth int, format string, argv ...interface{}) {
	Config.ErrorRDF(r, depth+1, format, argv...)
}

func WarnRF(r *http.Request, format string, argv ...interface{}) {
	Config.WarnRDF(r, 1, format, argv...)
}

func WarnRDF(r *http.Request, depth int, format string, argv ...interface{}) {
	Config.WarnRDF(r, depth+1, format, argv...)
}

func InfoRF(r *http.Request, format string, argv ...interface{}) {
	Config.InfoRDF(r, 1, format, argv...)
}

func InfoRDF(r *http.Request, depth int, format string, argv ...interface{}) {
	Config.InfoRDF(r, depth+1, format, argv...)
}

func DebugRF(r *http.Request, format string, argv ...interface{}) {
	Config.DebugRDF(r, 1, format, argv...)
}

func DebugRDF(r *http.Request, depth int, format string, argv ...interface{}) {
	Config.DebugRDF(r, depth+1, format, argv...)
}

func TraceRF(r *http.Request, format string, argv ...interface{}) {
	Config.TraceRDF(r, 1, format, argv...)
}

func TraceRDF(r *http.Request, depth int, format string, argv ...interface{}) {
	Config.TraceRDF(r, depth+1, format, argv...)
}

func PanicRF(r *http.Request, format string, argv ...interface{}) {
	Config.PanicDF(1, format, argv...)
}

func PanicRDF(r *http.Request, depth int, format string, argv ...interface{}) {
	Config.PanicDF(depth+1, format, argv...)
}

func FatalRF(r *http.Request, format string, argv ...interface{}) {
	Config.FatalRDF(r, 1, format, argv...)
}

func FatalRDF(r *http.Request, depth int, format string, argv ...interface{}) {
	Config.FatalRDF(r, depth+1, format, argv...)
}
