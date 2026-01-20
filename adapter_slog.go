/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"io"
	"log/slog"
	"os"
	"sync/atomic"
)

type adapterSlog struct {
	level   atomic.Uint32
	handler func(w io.Writer) slog.Handler
	log     *slog.Logger
}

func NewSLogJsonAdapter() Logger {
	obj := &adapterSlog{
		handler: func(w io.Writer) slog.Handler {
			return slog.NewJSONHandler(w, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})
		},
	}
	obj.level.Store(LevelDebug)
	obj.SetOutput(os.Stdout)
	return obj
}

func NewSLogStringAdapter() Logger {
	obj := &adapterSlog{
		handler: func(w io.Writer) slog.Handler {
			return slog.NewTextHandler(w, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})
		},
	}
	obj.level.Store(LevelDebug)
	obj.SetOutput(os.Stdout)
	return obj
}

func (v *adapterSlog) SetOutput(out io.Writer) {
	v.log = slog.New(v.handler(out))
}

func (v *adapterSlog) SetFormatter(_ Formatter) {}

func (v *adapterSlog) SetLevel(l uint32) {
	v.level.Store(l)
}

func (v *adapterSlog) Fatal(message string, args ...interface{}) {
	v.log.Error(message, args...)
	os.Exit(1)
}

func (v *adapterSlog) Error(message string, args ...interface{}) {
	if v.level.Load() < LevelError {
		return
	}
	v.log.Error(message, args...)
}

func (v *adapterSlog) Warn(message string, args ...interface{}) {
	if v.level.Load() < LevelWarn {
		return
	}
	v.log.Warn(message, args...)
}

func (v *adapterSlog) Info(message string, args ...interface{}) {
	if v.level.Load() < LevelInfo {
		return
	}
	v.log.Info(message, args...)
}

func (v *adapterSlog) Debug(message string, args ...interface{}) {
	if v.level.Load() < LevelDebug {
		return
	}
	v.log.Debug(message, args...)
}
