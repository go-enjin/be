// Copyright (c) 2024  The Go-Enjin Authors
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

package gorm

import (
	"database/sql"
	"time"

	"gorm.io/gorm/logger"
)

// Config is a buildable interface for configuring the sql.DB and
// logger.Config settings used by Gorm and SQL drivers
type Config interface {
	// SetConnMaxIdleTime configures the sql.DB setting of the same name
	// (default: 0 - no expiration)
	SetConnMaxIdleTime(maxIdle time.Duration) Config
	// SetConnMaxLifetime configures the sql.DB setting of the same name
	// (default: 0 - no expiration)
	SetConnMaxLifetime(maxLifetime time.Duration) Config
	// SetMaxIdleConns configures the sql.DB setting of the same name
	// (default: 2)
	SetMaxIdleConns(maxIdle int) Config
	// SetMaxOpenConns configures the sql.DB setting of the same name
	// (default: 0 - unlimited)
	SetMaxOpenConns(maxOpen int) Config
	// SetSlowThreshold specifies the corresponding logger.Config setting
	SetSlowThreshold(threshold time.Duration) Config
	// SetIgnoreRecordNotFoundError specifies the corresponding logger.Config setting
	SetIgnoreRecordNotFoundError(ignore bool) Config
	// SetParameterizedQueries specifies the corresponding logger.Config setting
	SetParameterizedQueries(parameterized bool) Config
	// SetLogLevel specifies the corresponding logger.Config setting
	SetLogLevel(level logger.LogLevel) Config

	make() *config
}

// NewConfig constructs a new buildable Config instance
func NewConfig() Config {
	return &settings{}
}

type settings struct {
	list []func(cfg *config)
}

func (s *settings) SetConnMaxIdleTime(maxIdle time.Duration) Config {
	s.list = append(s.list, func(cfg *config) {
		cfg.connMaxIdleTime = maxIdle
	})
	return s
}

func (s *settings) SetConnMaxLifetime(maxLifetime time.Duration) Config {
	s.list = append(s.list, func(cfg *config) {
		cfg.connMaxLifetime = maxLifetime
	})
	return s
}

func (s *settings) SetMaxIdleConns(maxIdle int) Config {
	s.list = append(s.list, func(cfg *config) {
		cfg.maxIdleConns = maxIdle
	})
	return s
}

func (s *settings) SetMaxOpenConns(maxOpen int) Config {
	s.list = append(s.list, func(cfg *config) {
		cfg.maxOpenConns = maxOpen
	})
	return s
}

func (s *settings) SetSlowThreshold(threshold time.Duration) Config {
	s.list = append(s.list, func(cfg *config) {
		cfg.slowThreshold = threshold
	})
	return s
}

func (s *settings) SetIgnoreRecordNotFoundError(ignore bool) Config {
	s.list = append(s.list, func(cfg *config) {
		cfg.ignoreRecordNotFoundError = ignore
	})
	return s
}

func (s *settings) SetParameterizedQueries(parameterized bool) Config {
	s.list = append(s.list, func(cfg *config) {
		cfg.parameterizedQueries = parameterized
	})
	return s
}

func (s *settings) SetLogLevel(level logger.LogLevel) Config {
	s.list = append(s.list, func(cfg *config) {
		cfg.logLevel = level
	})
	return s
}

func (s *settings) make() *config {
	cfg := &config{
		// defaults taken from https://pkg.go.dev/database/sql
		connMaxIdleTime: 0,
		connMaxLifetime: 0,
		maxIdleConns:    2,
		maxOpenConns:    0,
	}
	for _, fn := range s.list {
		fn(cfg)
	}
	return cfg
}

type config struct {
	// sql.DB settings
	connMaxIdleTime time.Duration
	connMaxLifetime time.Duration
	maxIdleConns    int
	maxOpenConns    int
	// logger.Config settings
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
	parameterizedQueries      bool
	logLevel                  logger.LogLevel
}

func (c *config) applySettings(db *sql.DB) {
	db.SetConnMaxIdleTime(c.connMaxIdleTime)
	db.SetConnMaxLifetime(c.connMaxLifetime)
	db.SetMaxIdleConns(c.maxIdleConns)
	db.SetMaxOpenConns(c.maxOpenConns)
}

func (c *config) loggerConfig() logger.Config {
	return logger.Config{
		Colorful:                  false, // always disable colour
		SlowThreshold:             c.slowThreshold,
		IgnoreRecordNotFoundError: c.ignoreRecordNotFoundError,
		ParameterizedQueries:      c.parameterizedQueries,
		LogLevel:                  c.logLevel,
	}
}
