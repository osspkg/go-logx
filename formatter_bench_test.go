/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx_test

import (
	"io"
	"testing"
	"time"
	"unsafe"

	"go.osspkg.com/logx"
)

func BenchmarkFormatStringEncode(b *testing.B) {
	benchmarkFormatters(b, logx.NewFormatString())
}

func BenchmarkFormatJSONEncode(b *testing.B) {
	benchmarkFormatters(b, logx.NewFormatJSON())
}

func BenchmarkFormatJSONContextPairs(b *testing.B) {
	cases := []struct {
		name string
		ctx  []any
	}{
		{name: "three", ctx: []any{"a", 1, "b", 2, "c", 3}},
		{name: "four", ctx: []any{"a", 1, "b", 2, "c", 3, "d", 4}},
		{name: "eight", ctx: []any{"a", 1, "b", 2, "c", 3, "d", 4, "e", 5, "f", 6, "g", 7, "h", 8}},
		{name: "duplicate", ctx: []any{"a", 1, "b", 2, "a", 3}},
	}
	formatter := logx.NewFormatJSON()
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			message := logx.Message{
				Time:  time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				Level: "INFO", Message: "sync", Ctx: tc.ctx,
			}
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if err := formatter.Encode(io.Discard, &message); err != nil {
						b.Error(err)
						return
					}
				}
			})
		})
	}
}

func benchmarkFormatters(b *testing.B, formatter logx.Formatter) {
	plain := "plain ascii payload"
	escaped := "quote=\" slash=\\ newline=\n unicode=☃"
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
		{
			name: "string_plain",
			message: logx.Message{
				Time:  time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				Level: "INFO", Message: "sync", Ctx: []any{"payload", plain},
			},
		},
		{
			name: "string_escaped",
			message: logx.Message{
				Time:  time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				Level: "INFO", Message: "sync", Ctx: []any{"payload", escaped},
			},
		},
		{
			name: "bytes_plain",
			message: logx.Message{
				Time:  time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				Level: "INFO", Message: "sync", Ctx: []any{"payload", stringBytesView(plain)},
			},
		},
		{
			name: "bytes_escaped",
			message: logx.Message{
				Time:  time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				Level: "INFO", Message: "sync", Ctx: []any{"payload", stringBytesView(escaped)},
			},
		},
		{
			name: "text_marshaler",
			message: logx.Message{
				Time:  time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				Level: "INFO", Message: "sync", Ctx: []any{"payload", benchmarkTextMarshaler(stringBytesView(escaped))},
			},
		},
		{
			name: "binary_marshaler",
			message: logx.Message{
				Time:  time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				Level: "INFO", Message: "sync", Ctx: []any{"payload", benchmarkBinaryMarshaler(stringBytesView(escaped))},
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

// stringBytesView returns a read-only byte view used by formatter benchmarks.
// The formatter implementations only read input values and never mutate them.
func stringBytesView(value string) []byte {
	if value == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(value), len(value))
}

type benchmarkTextMarshaler []byte

func (value benchmarkTextMarshaler) MarshalText() ([]byte, error) {
	return value, nil
}

type benchmarkBinaryMarshaler []byte

func (value benchmarkBinaryMarshaler) MarshalBinary() ([]byte, error) {
	return value, nil
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

func BenchmarkNewPayloadFormats(b *testing.B) {
	plain := "plain ascii payload"
	escaped := "quote=\" slash=\\ newline=\n unicode=☃"
	payloads := []struct {
		name  string
		value any
	}{
		{name: "string_plain", value: plain},
		{name: "string_escaped", value: escaped},
		{name: "bytes_plain", value: stringBytesView(plain)},
		{name: "bytes_escaped", value: stringBytesView(escaped)},
		{name: "text_marshaler", value: benchmarkTextMarshaler(stringBytesView(escaped))},
		{name: "binary_marshaler", value: benchmarkBinaryMarshaler(stringBytesView(escaped))},
	}
	loggers := []struct {
		name string
		new  func() logx.Logger
	}{
		{
			name: "json",
			new: func() logx.Logger {
				return logx.New(logx.WithOutput(io.Discard), logx.WithLevel(logx.LevelDebug), logx.WithFormatter(logx.NewFormatJSON()))
			},
		},
		{
			name: "string",
			new: func() logx.Logger {
				return logx.New(logx.WithOutput(io.Discard), logx.WithLevel(logx.LevelDebug), logx.WithFormatter(logx.NewFormatString()))
			},
		},
		{
			name: "slog_json",
			new: func() logx.Logger {
				return logx.NewSLogJsonAdapter(logx.WithOutput(io.Discard), logx.WithLevel(logx.LevelDebug))
			},
		},
		{
			name: "slog_string",
			new: func() logx.Logger {
				return logx.NewSLogStringAdapter(logx.WithOutput(io.Discard), logx.WithLevel(logx.LevelDebug))
			},
		},
	}

	for _, loggerCase := range loggers {
		b.Run(loggerCase.name, func(b *testing.B) {
			for _, payload := range payloads {
				b.Run(payload.name, func(b *testing.B) {
					benchmarkLogger(b, loggerCase.new(), "payload", payload.value)
				})
			}
		})
	}
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
