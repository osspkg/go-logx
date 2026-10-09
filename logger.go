/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Log base model
type Log struct {
	level     uint32
	writeMu   sync.Mutex
	writer    io.Writer
	formatter Formatter
}

// New creates a logger with the supplied configuration options.
func New(options ...Option) *Log {
	cfg := newLoggerConfig(LevelError, NewFormatJSON(), options)
	return &Log{
		level:     cfg.level,
		writer:    cfg.writer,
		formatter: cfg.formatter,
	}
}

func (l *Log) writeMessage(level uint32, call func(v *Message)) {
	if l.GetLevel() < level {
		return
	}

	m := poolMessage.Get()
	defer func() {
		poolMessage.Put(m)
	}()

	call(m)

	lvl, ok := levels[level]
	if !ok {
		lvl = "UNK"
	}
	m.Level, m.Time = lvl, time.Now()

	l.writeMu.Lock()
	defer l.writeMu.Unlock()

	err := l.formatter.Encode(l.writer, m)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}

}

// GetLevel returns the minimum level written by the logger.
func (l *Log) GetLevel() uint32 {
	return l.level
}

func (l *Log) Info(message string, args ...any) {
	l.writeMessage(LevelInfo, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
}

func (l *Log) Warn(message string, args ...any) {
	l.writeMessage(LevelWarn, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
}

func (l *Log) Error(message string, args ...any) {
	l.writeMessage(LevelError, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
}

func (l *Log) Debug(message string, args ...any) {
	l.writeMessage(LevelDebug, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
}

func (l *Log) Fatal(message string, args ...any) {
	l.writeMessage(levelFatal, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
	os.Exit(1)
}
