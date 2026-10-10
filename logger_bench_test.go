/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx_test

import (
	"io"
	"os"
	"testing"

	"go.osspkg.com/logx"
)

func BenchmarkLoggerWriters(b *testing.B) {
	formatters := []struct {
		name string
		new  func() logx.Formatter
	}{
		{name: "json", new: func() logx.Formatter { return logx.NewFormatJSON() }},
		{name: "string", new: func() logx.Formatter { return logx.NewFormatString() }},
	}
	writers := []struct {
		name string
		new  func(*testing.B) io.Writer
	}{
		{name: "discard", new: func(*testing.B) io.Writer { return io.Discard }},
		{name: "devnull", new: benchmarkDevNullWriter},
	}

	for _, formatter := range formatters {
		b.Run(formatter.name, func(b *testing.B) {
			for _, output := range writers {
				b.Run(output.name, func(b *testing.B) {
					logger := logx.New(
						logx.WithOutput(output.new(b)),
						logx.WithLevel(logx.LevelDebug),
						logx.WithFormatter(formatter.new()),
					)
					b.ReportAllocs()
					b.RunParallel(func(pb *testing.PB) {
						for pb.Next() {
							logger.Info("sync", "id", 1, "obj", testData{})
						}
					})
				})
			}
		})
	}
}

func benchmarkDevNullWriter(b *testing.B) io.Writer {
	b.Helper()
	file, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = file.Close() })
	return file
}
