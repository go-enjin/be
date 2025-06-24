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
	// "log/syslog"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	syslog "github.com/dmachard/go-clientsyslog"
	"github.com/go-enjin/be/pkg/globals"

	"github.com/sirupsen/logrus"
)

type SyslogHook struct {
	writer    *syslog.Writer
	dsn       *SyslogDSN
	priority  syslog.Priority
	tag       string
	m         *sync.RWMutex
	reconnect func() error
}

func NewSyslogHook(dsn *SyslogDSN, priority syslog.Priority, tag string) (hook *SyslogHook, err error) {
	return NewSyslogHookWith(dsn, priority, tag, nil, nil)
}

func NewSyslogHookWith(dsn *SyslogDSN, priority syslog.Priority, tag string, tlsConfig *tls.Config, dialer *net.Dialer) (hook *SyslogHook, err error) {
	hook = &SyslogHook{
		dsn:      dsn,
		priority: priority,
		tag:      tag,
		m:        &sync.RWMutex{},
	}

	hook.reconnect = func() (ee error) {
		if hook.writer != nil {
			_ = hook.writer.Close()
		}

		switch hook.dsn.network {
		case "", "unix", "local":
			// no reconnect checking necessary
			hook.writer, ee = syslog.Dial("", "", hook.priority, hook.tag)
			hook.writer.SetProgram(globals.BinName)
			return

		case "udp", "tcp":

		case "tcp+tls":
			if tlsConfig == nil {
				tlsConfig = &tls.Config{
					InsecureSkipVerify: hook.dsn.Options.Bool("SkipVerifyTLS"),
				}
			}
		}

		if dialer == nil {
			dialer = &net.Dialer{
				Timeout:   time.Second * 5,
				KeepAlive: time.Second,
				KeepAliveConfig: net.KeepAliveConfig{
					Enable:   true,
					Idle:     time.Second,
					Interval: time.Second,
					Count:    2,
				},
			}
		}

		if hook.writer, err = syslog.DialWithCustomDialer("custom", hook.dsn.Raddr(), hook.priority, hook.tag, func(network, addr string) (conn net.Conn, eee error) {
			if tlsConfig != nil {
				return tls.DialWithDialer(dialer, "tcp", hook.dsn.Raddr(), tlsConfig)
			}
			return dialer.Dial(hook.dsn.network, hook.dsn.Raddr())
		}); err != nil {
			return fmt.Errorf("syslog dialer error: %w", err)
		}

		if hook.writer != nil && ee == nil {
			hook.writer.SetProgram(globals.BinName)
		}
		return
	}

	err = hook.reconnect()
	return
}

func (hook *SyslogHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		ErrorF("Unable to read entry, %v", err)
		return err
	}

	switch entry.Level {
	case logrus.PanicLevel:
		return hook.writer.Crit(line)
	case logrus.FatalLevel:
		return hook.writer.Crit(line)
	case logrus.ErrorLevel:
		return hook.writer.Err(line)
	case logrus.WarnLevel:
		return hook.writer.Warning(line)
	case logrus.InfoLevel:
		return hook.writer.Info(line)
	case logrus.DebugLevel, logrus.TraceLevel:
		return hook.writer.Debug(line)
	default:
		return nil
	}
}

func (hook *SyslogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *SyslogHook) Reconnect() (err error) {
	if hook.reconnect != nil {
		return hook.reconnect()
	}
	return
}
