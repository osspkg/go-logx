/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx_test

import (
	"io"
	"testing"
	"time"

	"go.osspkg.com/logx"
)

func BenchmarkFormatStringEncode(b *testing.B) {
	benchmarkFormatters(b, logx.NewFormatString())
}

func BenchmarkFormatJSONEncode(b *testing.B) {
	benchmarkFormatters(b, logx.NewFormatJSON())
}

func benchmarkFormatters(b *testing.B, formatter logx.Formatter) {
	messages := []struct {
		name    string
		message logx.Message
	}{
		{
			name: "primitives",
			message: logx.Message{
				Time:    time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				Level:   "INFO",
				Message: "sync",
				Ctx:     []any{"id", 1, "status", "ok"},
			},
		},
		{
			name: "struct",
			message: logx.Message{
				Time:    time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				Level:   "INFO",
				Message: "sync",
				Ctx:     []any{"id", 1, "obj", testData{A: "x", B: 2}},
			},
		},
	}

	for _, tc := range messages {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if err := formatter.Encode(io.Discard, &tc.message); err != nil {
						b.Error(err)
						return
					}
				}
			})
		})
	}
}

func BenchmarkNewStringPrimitives(b *testing.B) {
	benchmarkLogger(b, logx.New(
		logx.WithOutput(io.Discard),
		logx.WithLevel(logx.LevelDebug),
		logx.WithFormatter(logx.NewFormatString()),
	), "id", 1, "status", "ok")
}

func BenchmarkNewJSONPrimitives(b *testing.B) {
	benchmarkLogger(b, logx.New(
		logx.WithOutput(io.Discard),
		logx.WithLevel(logx.LevelDebug),
		logx.WithFormatter(logx.NewFormatJSON()),
	), "id", 1, "status", "ok")
}

func benchmarkLogger(b *testing.B, logger logx.Logger, args ...any) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("sync", args...)
		}
	})
}
