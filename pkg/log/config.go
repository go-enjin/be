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

import (
	"fmt"
	"log"
	"log/syslog"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

const (
	StandardTimestampFormat = "2006-01-02T15:04:05.000"
	DefaultTimestampFormat  = "20060102-150405.00"
)

type Configuration struct {
	DisableTimestamp bool
	TimestampFormat  string
	LoggingFormat    Format
	LogLevel         Level
	LogHook          string
	AppName          string
	PapertrailHost   string
	PapertrailPort   int
	PapertrailTag    string
}

var (
	Config = &Configuration{
		DisableTimestamp: false,
		TimestampFormat:  DefaultTimestampFormat,
		LoggingFormat:    FormatPretty,
		LogLevel:         LevelInfo,
		LogHook:          "stdout",
		AppName:          "",
		PapertrailHost:   "",
		PapertrailPort:   0,
		PapertrailTag:    "",
	}
)

func (c Configuration) Apply() {
	logger = logrus.New()

	switch c.LoggingFormat {
	case FormatJson:
		logger.SetFormatter(&logrus.JSONFormatter{
			DisableTimestamp: c.DisableTimestamp,
			TimestampFormat:  c.TimestampFormat,
		})
	case FormatText:
		logger.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp: c.DisableTimestamp,
			TimestampFormat:  c.TimestampFormat,
			DisableSorting:   true,
			DisableColors:    true,
			FullTimestamp:    !c.DisableTimestamp,
		})
	case FormatPretty:
		fallthrough
	default:
		logger.SetFormatter(&prefixed.TextFormatter{
			DisableTimestamp: c.DisableTimestamp,
			TimestampFormat:  c.TimestampFormat,
			ForceFormatting:  true,
			FullTimestamp:    true,
			DisableSorting:   true,
			DisableColors:    true,
		})
	}

	switch c.LogLevel {
	case LevelTrace:
		logger.SetLevel(logrus.TraceLevel)
	case LevelDebug:
		logger.SetLevel(logrus.DebugLevel)
	case LevelInfo:
		logger.SetLevel(logrus.InfoLevel)
	case LevelWarn:
		logger.SetLevel(logrus.WarnLevel)
	case LevelError:
		fallthrough
	default:
		logger.SetLevel(logrus.ErrorLevel)
	}

	switch c.LogHook {
	case "syslog":
		if hook, err := NewSyslogLocalHook(syslog.LOG_INFO, c.AppName); err != nil {
			panic(fmt.Sprintf("error setting up syslog hook: %v", err))
		} else {
			logger.AddHook(hook)
		}
	case "papertrail":
		if hook, err := NewPapertrailHook(c.PapertrailTag, c.PapertrailHost, c.PapertrailPort); err != nil {
			panic(fmt.Sprintf("error setting up papertrail hook: %v", err))
		} else {
			logger.AddHook(hook)
		}
	case "stdout":
		fallthrough
	default:
	}

	logLogger = log.New(logger.Writer(), c.AppName, 0)
}
