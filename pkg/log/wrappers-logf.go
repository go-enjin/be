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

package log

type LogfLogger Level

func NewLogf(level Level) LogfLogger {
	return LogfLogger(level)
}

func (l LogfLogger) Logf(format string, argv ...interface{}) {
	switch l {
	case LogfLogger(LevelInfo):
		InfoDF(1, format, argv...)
	case LogfLogger(LevelWarn):
		WarnDF(1, format, argv...)
	case LogfLogger(LevelError):
		ErrorDF(1, format, argv...)
	case LogfLogger(LevelDebug):
		DebugDF(1, format, argv...)
	case LogfLogger(LevelTrace):
		TraceDF(1, format, argv...)
	default:
		DebugDF(1, format, argv...)
	}
}
