/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"io"
	"os"
)

// Option configures a logger when it is created.
type Option func(*loggerConfig)

type loggerConfig struct {
	level     uint32
	writer    io.Writer
	formatter Formatter
}

func newLoggerConfig(level uint32, formatter Formatter, options []Option) loggerConfig {
	cfg := loggerConfig{
		level:     level,
		writer:    os.Stdout,
		formatter: formatter,
	}
	for _, option := range options {
		if option != nil {
			option(&cfg)
		}
	}
	return cfg
}

// WithLevel sets the minimum level written by the logger.
func WithLevel(level uint32) Option {
	return func(cfg *loggerConfig) {
		cfg.level = level
	}
}

// WithOutput sets the logger's destination. A nil writer leaves the default
// destination unchanged.
func WithOutput(writer io.Writer) Option {
	return func(cfg *loggerConfig) {
		if writer != nil {
			cfg.writer = writer
		}
	}
}

// WithFormatter sets the formatter used by the logger. A nil formatter leaves
// the constructor's default formatter unchanged.
func WithFormatter(formatter Formatter) Option {
	return func(cfg *loggerConfig) {
		if formatter != nil {
			cfg.formatter = formatter
		}
	}
}
