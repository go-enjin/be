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

package log

import (
	"log"

	"github.com/sirupsen/logrus"
)

const (
	StandardTimestampFormat = "2006-01-02T15:04:05.000"
	DefaultTimestampFormat  = "20060102-150405.00"
)

var initialLogger = logrus.New()

var Config *config = &config{
	Configuration: Configuration{
		DisableTimestamp: false,
		TimestampFormat:  DefaultTimestampFormat,
		LoggingFormat:    FormatPretty,
		LogLevel:         LevelInfo,
		LogHook:          "stdout",
		AppName:          "",
		RemoteHost:       "",
		RemotePort:       0,
		LogTag:           "",
		logger:           initialLogger,
		writer:           log.New(initialLogger.Writer(), "", 0),
	},
	_private: true,
}

type config struct {
	Configuration

	_private bool
}
