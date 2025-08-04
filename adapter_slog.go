/*
 *  Copyright (c) 2024-2025 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"context"
	"io"
	"log/slog"
	"os"
)

var slogLevelMap = map[uint32]slog.Level{
	LevelError: slog.LevelError,
	LevelWarn:  slog.LevelWarn,
	LevelInfo:  slog.LevelInfo,
	LevelDebug: slog.LevelDebug,
}

type adapterSlog struct {
	level   slog.Level
	handler func(w io.Writer, level slog.Level) slog.Handler
	log     *slog.Logger
}

func NewSLogJsonAdapter() Logger {
	obj := &adapterSlog{
		level: slog.LevelDebug,
		handler: func(w io.Writer, level slog.Level) slog.Handler {
			return slog.NewJSONHandler(w, &slog.HandlerOptions{
				Level: level,
			})
		},
	}
	obj.SetOutput(os.Stdout)
	return obj
}

func NewSLogStringAdapter() Logger {
	obj := &adapterSlog{
		level: slog.LevelDebug,
		handler: func(w io.Writer, level slog.Level) slog.Handler {
			return slog.NewTextHandler(w, &slog.HandlerOptions{
				Level: level,
			})
		},
	}
	obj.SetOutput(os.Stdout)
	return obj
}

func (v *adapterSlog) SetOutput(out io.Writer) {
	v.log = slog.New(v.handler(out, v.level))
}

func (v *adapterSlog) SetFormatter(_ Formatter) {}

func (v *adapterSlog) SetLevel(l uint32) {
	level, ok := slogLevelMap[l]
	if !ok {
		level = slog.LevelDebug
	}
	v.log.Enabled(context.TODO(), level)
}

func (v *adapterSlog) Fatal(message string, args ...interface{}) {
	v.log.Error(message, args...)
	os.Exit(1)
}

func (v *adapterSlog) Error(message string, args ...interface{}) {
	v.log.Error(message, args...)
}

func (v *adapterSlog) Warn(message string, args ...interface{}) {
	v.log.Warn(message, args...)
}

func (v *adapterSlog) Info(message string, args ...interface{}) {
	v.log.Info(message, args...)
}

func (v *adapterSlog) Debug(message string, args ...interface{}) {
	v.log.Debug(message, args...)
}
