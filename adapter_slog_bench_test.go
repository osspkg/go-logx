/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"go.osspkg.com/logx"
)

func BenchmarkSlogTextPaths(b *testing.B) {
	adapter := logx.NewSLogStringAdapter(logx.WithOutput(io.Discard), logx.WithLevel(logx.LevelDebug))
	standard := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	standardInterface := slogInfoAsInterface(standard)
	cases := []struct {
		name string
		log  func()
	}{
		{name: "adapter", log: func() { adapter.Info("sync", "id", 1, "obj", testData{}) }},
		{name: "logger_log", log: func() { standard.Log(context.Background(), slog.LevelInfo, "sync", "id", 1, "obj", testData{}) }},
		{name: "logger_info", log: func() { standard.Info("sync", "id", 1, "obj", testData{}) }},
		{name: "logger_interface", log: func() { standardInterface.Info("sync", "id", 1, "obj", testData{}) }},
		{name: "logger_attrs", log: func() {
			standard.LogAttrs(context.Background(), slog.LevelInfo, "sync", slog.Int("id", 1), slog.Any("obj", testData{}))
		}},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					tc.log()
				}
			})
		})
	}
}

// Keep dynamic dispatch in the comparison with the Logger return type.
//go:noinline
func slogInfoAsInterface(logger *slog.Logger) interface{ Info(string, ...any) } {
	return logger
}
