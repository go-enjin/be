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

func ErrorF(format string, argv ...interface{}) {
	ErrorDF(1, format, argv...)
}

func ErrorDF(depth int, format string, argv ...interface{}) {
	depth += 1
	logger.Errorf(prefixLogEntry(depth, format), argv...)
}

func WarnF(format string, argv ...interface{}) {
	WarnDF(1, format, argv...)
}

func WarnDF(depth int, format string, argv ...interface{}) {
	depth += 1
	logger.Warnf(prefixLogEntry(depth, format), argv...)
}

func InfoF(format string, argv ...interface{}) {
	InfoDF(1, format, argv...)
}

func InfoDF(depth int, format string, argv ...interface{}) {
	depth += 1
	logger.Infof(prefixLogEntry(depth, format), argv...)
}

func DebugF(format string, argv ...interface{}) {
	DebugDF(1, format, argv...)
}

func DebugDF(depth int, format string, argv ...interface{}) {
	depth += 1
	logger.Debugf(prefixLogEntry(depth, format), argv...)
}

func TraceF(format string, argv ...interface{}) {
	TraceDF(1, format, argv...)
}

func TraceDF(depth int, format string, argv ...interface{}) {
	depth += 1
	logger.Tracef(prefixLogEntry(depth, format), argv...)
}

func PanicF(format string, argv ...interface{}) {
	PanicDF(1, format, argv...)
}

func PanicDF(depth int, format string, argv ...interface{}) {
	depth += 1
	logger.Panicf(prefixLogEntry(depth, format), argv...)
}

func FatalF(format string, argv ...interface{}) {
	FatalDF(1, format, argv...)
}

func FatalDF(depth int, format string, argv ...interface{}) {
	depth += 1
	logger.Fatalf(prefixLogEntry(depth, format), argv...)
}