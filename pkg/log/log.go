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
	"io"
	"log"

	"github.com/sirupsen/logrus"
)

var logLogger *log.Logger

var logger = logrus.New()

func ErrorWriter() *io.PipeWriter {
	return logger.WriterLevel(logrus.ErrorLevel)
}

func WarnWriter() *io.PipeWriter {
	return logger.WriterLevel(logrus.WarnLevel)
}

func InfoWriter() *io.PipeWriter {
	return logger.WriterLevel(logrus.InfoLevel)
}

func DebugWriter() *io.PipeWriter {
	return logger.WriterLevel(logrus.DebugLevel)
}

func TraceWriter() *io.PipeWriter {
	return logger.WriterLevel(logrus.TraceLevel)
}

func Logrus() *logrus.Logger {
	return logger
}

func Logger() *log.Logger {
	if logLogger == nil {
		logLogger = log.New(logger.Writer(), "", 0)
	}
	return logLogger
}

func PrefixedLogger(prefix string) (logging *log.Logger) {
	logging = log.New(logger.Writer(), prefix, 0)
	return
}
