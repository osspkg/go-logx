/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"
)

type adapterSlog struct {
	level     uint32
	writer    io.Writer
	formatter Formatter
	log       *slog.Logger
	writeMu   sync.Mutex
	writeGate serializedWriter
}

// NewSLogJsonAdapter creates a slog adapter with JSON output.
func NewSLogJsonAdapter(options ...Option) Logger {
	return newSlogAdapter(func(writer io.Writer, opts *slog.HandlerOptions) slog.Handler {
		return slog.NewJSONHandler(writer, opts)
	}, options...)
}

// NewSLogStringAdapter creates a slog adapter with text output.
func NewSLogStringAdapter(options ...Option) Logger {
	return newSlogAdapter(func(writer io.Writer, opts *slog.HandlerOptions) slog.Handler {
		return slog.NewTextHandler(writer, opts)
	}, options...)
}

func newSlogAdapter(newHandler func(io.Writer, *slog.HandlerOptions) slog.Handler, options ...Option) Logger {
	cfg := newLoggerConfig(LevelDebug, nil, options)
	adapter := &adapterSlog{
		level:     cfg.level,
		writer:    cfg.writer,
		formatter: cfg.formatter,
	}
	adapter.writeGate = serializedWriter{mu: &adapter.writeMu, writer: adapter.writer}
	if adapter.formatter == nil {
		adapter.log = slog.New(newHandler(cfg.writer, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return adapter
}

func (v *adapterSlog) Fatal(message string, args ...any) {
	v.writeMessage(levelFatal, slog.LevelError, true, message, args...)
	os.Exit(1)
}

func (v *adapterSlog) Error(message string, args ...any) {
	v.writeMessage(LevelError, slog.LevelError, false, message, args...)
}

func (v *adapterSlog) Warn(message string, args ...any) {
	v.writeMessage(LevelWarn, slog.LevelWarn, false, message, args...)
}

func (v *adapterSlog) Info(message string, args ...any) {
	v.writeMessage(LevelInfo, slog.LevelInfo, false, message, args...)
}

func (v *adapterSlog) Debug(message string, args ...any) {
	v.writeMessage(LevelDebug, slog.LevelDebug, false, message, args...)
}

func (v *adapterSlog) writeMessage(level uint32, slogLevel slog.Level, force bool, message string, args ...any) {
	if !force && v.level < level {
		return
	}

	if v.formatter == nil {
		v.log.Log(context.Background(), slogLevel, message, args...)
		return
	}

	msg := poolMessage.Get()
	defer poolMessage.Put(msg)
	msg.Level, msg.Message, msg.Time = levelName(level), message, time.Now()
	msg.Ctx = append(msg.Ctx, args...)

	err := v.encode(msg)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}

func (v *adapterSlog) encode(message *Message) error {
	switch v.formatter.(type) {
	case *FormatJSON, *FormatString:
		return v.formatter.Encode(&v.writeGate, message)
	}

	v.writeMu.Lock()
	defer v.writeMu.Unlock()
	return v.formatter.Encode(v.writer, message)
}
