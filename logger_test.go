/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"testing"

	"go.osspkg.com/casecheck"
	"go.osspkg.com/syncing"

	"go.osspkg.com/logx"
)

type mockWriter struct {
	b *bytes.Buffer
	l sync.Mutex
}

type testData struct {
	A string
	B int
}

func newMockWriter() *mockWriter {
	return &mockWriter{
		b: bytes.NewBuffer(nil),
		l: sync.Mutex{},
	}
}

func (v *mockWriter) Write(b []byte) (int, error) {
	v.l.Lock()
	defer v.l.Unlock()
	return v.b.Write(b)
}

func (v *mockWriter) String() string {
	v.l.Lock()
	defer v.l.Unlock()
	return v.b.String()
}

func TestUnit_NewJSON(t *testing.T) {
	casecheck.NotNil(t, logx.Default())

	wg := syncing.NewGroup(context.TODO())
	buff := newMockWriter()

	previous := logx.Default()
	logx.SetDefault(logx.New(logx.WithOutput(buff), logx.WithLevel(logx.LevelDebug)))
	defer logx.SetDefault(previous)

	wg.Background("", func(_ context.Context) { logx.Info("async", "id", 1) })
	wg.Background("", func(_ context.Context) { logx.Warn("async", "id", 2) })
	wg.Background("", func(_ context.Context) { logx.Error("async", "id", 3) })
	wg.Background("", func(_ context.Context) { logx.Debug("async", "id", 4) })

	logx.Info("sync", "id", 1)
	logx.Warn("sync", "id", 2)
	logx.Error("sync", "id", 3)
	logx.Debug("sync", "id", 4)

	logx.Info("context1", "ip", "0.0.0.0")
	logx.Info("context2", "nil", nil)
	logx.Info("context3", "func", func() {})
	logx.Info("context4", "err", fmt.Errorf("er1"))
	logx.Info("context5", "obj", testData{})

	wg.Wait()

	data := buff.String()
	casecheck.Contains(t, data, `"level":"INFO","msg":"async","ctx":{"id":"1"}`)
	casecheck.Contains(t, data, `"level":"WARN","msg":"async","ctx":{"id":"2"}`)
	casecheck.Contains(t, data, `"level":"ERROR","msg":"async","ctx":{"id":"3"}`)
	casecheck.Contains(t, data, `"level":"DEBUG","msg":"async","ctx":{"id":"4"}`)
	casecheck.Contains(t, data, `"level":"INFO","msg":"sync","ctx":{"id":"1"}`)
	casecheck.Contains(t, data, `"level":"WARN","msg":"sync","ctx":{"id":"2"}`)
	casecheck.Contains(t, data, `"level":"ERROR","msg":"sync","ctx":{"id":"3"}`)
	casecheck.Contains(t, data, `"level":"DEBUG","msg":"sync","ctx":{"id":"4"}`)
	casecheck.Contains(t, data, `"level":"INFO","msg":"context1","ctx":{"ip":"0.0.0.0"}`)
	casecheck.Contains(t, data, `"level":"INFO","msg":"context2","ctx":{"nil":"null"}`)
	casecheck.Contains(t, data, `"level":"INFO","msg":"context3","ctx":{"func":"(func())(0x`)
	casecheck.Contains(t, data, `"level":"INFO","msg":"context4","ctx":{"err":"er1"}`)
	casecheck.Contains(t, data, `"level":"INFO","msg":"context5","ctx":{"obj":"logx_test.testData{A:\"\", B:0}"}`)
}

func TestUnit_NewString(t *testing.T) {
	wg := syncing.NewGroup(context.TODO())
	buff := newMockWriter()

	l := logx.New(
		logx.WithFormatter(logx.NewFormatString()),
		logx.WithOutput(buff),
		logx.WithLevel(logx.LevelDebug),
	)
	casecheck.NotNil(t, l)
	casecheck.Equal(t, logx.LevelDebug, l.GetLevel())

	wg.Background("", func(_ context.Context) { l.Info("async", "id", 1) })
	wg.Background("", func(_ context.Context) { l.Warn("async", "id", 2) })
	wg.Background("", func(_ context.Context) { l.Error("async", "id", 3) })
	wg.Background("", func(_ context.Context) { l.Debug("async", "id", 4) })

	l.Info("sync", "id", 1)
	l.Warn("sync", "id", 2)
	l.Error("sync", "id", 3)
	l.Debug("sync", "id", 4)

	l.Info("context1", "ip", "0.0.0.0\n")
	l.Info("context2", "nil", nil)
	l.Info("context3", "func", func() {})
	l.Info("context4", "err", fmt.Errorf("er1"))
	l.Info("context5", "obj", testData{})

	wg.Wait()

	data := buff.String()
	casecheck.Contains(t, data, "\"level\"=\"INFO\"\t\"msg\"=\"async\"\t\"id\"=\"1\"")
	casecheck.Contains(t, data, "\"level\"=\"WARN\"\t\"msg\"=\"async\"\t\"id\"=\"2\"")
	casecheck.Contains(t, data, "\"level\"=\"ERROR\"\t\"msg\"=\"async\"\t\"id\"=\"3\"")
	casecheck.Contains(t, data, "\"level\"=\"DEBUG\"\t\"msg\"=\"async\"\t\"id\"=\"4\"")
	casecheck.Contains(t, data, "\"level\"=\"INFO\"\t\"msg\"=\"sync\"\t\"id\"=\"1\"")
	casecheck.Contains(t, data, "\"level\"=\"WARN\"\t\"msg\"=\"sync\"\t\"id\"=\"2\"")
	casecheck.Contains(t, data, "\"level\"=\"ERROR\"\t\"msg\"=\"sync\"\t\"id\"=\"3\"")
	casecheck.Contains(t, data, "\"level\"=\"DEBUG\"\t\"msg\"=\"sync\"\t\"id\"=\"4\"")
	casecheck.Contains(t, data, "\"level\"=\"INFO\"\t\"msg\"=\"context1\"\t\"ip\"=\"0.0.0.0\\n\"")
	casecheck.Contains(t, data, "\"level\"=\"INFO\"\t\"msg\"=\"context2\"\t\"nil\"=\"null\"")
	casecheck.Contains(t, data, "\"level\"=\"INFO\"\t\"msg\"=\"context3\"\t\"func\"=\"(func())(0x")
	casecheck.Contains(t, data, "\"level\"=\"INFO\"\t\"msg\"=\"context4\"\t\"err\"=\"er1\"")
	casecheck.Contains(t, data, "\"level\"=\"INFO\"\t\"msg\"=\"context5\"\t\"obj\"=\"logx_test.testData{A:\"\", B:0}\"")
}

func TestUnit_NewSlog(t *testing.T) {
	wg := syncing.NewGroup(context.TODO())
	buff := newMockWriter()

	l := logx.NewSLogJsonAdapter(
		logx.WithOutput(buff),
		logx.WithLevel(logx.LevelDebug),
	)
	casecheck.NotNil(t, l)

	wg.Background("", func(_ context.Context) { l.Info("async", "id", 1) })
	wg.Background("", func(_ context.Context) { l.Warn("async", "id", 2) })
	wg.Background("", func(_ context.Context) { l.Error("async", "id", 3) })
	wg.Background("", func(_ context.Context) { l.Debug("async", "id", 4) })

	l.Info("sync", "id", 1)
	l.Warn("sync", "id", 2)
	l.Error("sync", "id", 3)
	l.Debug("sync", "id", 4)

	l.Info("context1", "ip", "0.0.0.0\n")
	l.Info("context2", "nil", nil)
	l.Info("context3", "func", func() {})
	l.Info("context4", "err", fmt.Errorf("er1"))
	l.Info("context5", "obj", testData{})

	wg.Wait()

	data := buff.String()
	casecheck.Contains(t, data, "\"level\":\"INFO\",\"msg\":\"sync\",\"id\":1")
	casecheck.Contains(t, data, "\"level\":\"WARN\",\"msg\":\"sync\",\"id\":2")
	casecheck.Contains(t, data, "\"level\":\"ERROR\",\"msg\":\"sync\",\"id\":3")
	casecheck.Contains(t, data, "\"level\":\"DEBUG\",\"msg\":\"sync\",\"id\":4")
	casecheck.Contains(t, data, "\"level\":\"INFO\",\"msg\":\"async\",\"id\":1")
	casecheck.Contains(t, data, "\"level\":\"WARN\",\"msg\":\"async\",\"id\":2")
	casecheck.Contains(t, data, "\"level\":\"ERROR\",\"msg\":\"async\",\"id\":3")
	casecheck.Contains(t, data, "\"level\":\"DEBUG\",\"msg\":\"async\",\"id\":4")
	casecheck.Contains(t, data, "\"level\":\"INFO\",\"msg\":\"context1\",\"ip\":\"0.0.0.0\\n\"")
	casecheck.Contains(t, data, "\"level\":\"INFO\",\"msg\":\"context2\",\"nil\":null")
	casecheck.Contains(t, data, "\"level\":\"INFO\",\"msg\":\"context3\",\"func\":\"!ERROR:json: unsupported type: func()\"")
	casecheck.Contains(t, data, "\"level\":\"INFO\",\"msg\":\"context4\",\"err\":\"er1\"")
	casecheck.Contains(t, data, "\"level\":\"INFO\",\"msg\":\"context5\",\"obj\":{\"A\":\"\",\"B\":0}")
}

/*
goos: linux
goarch: amd64
pkg: go.osspkg.com/logx
cpu: 12th Gen Intel(R) Core(TM) i9-12900KF
BenchmarkNewJSON
BenchmarkNewJSON-24                 	 7479630	       162.4 ns/op	      32 B/op	       1 allocs/op
BenchmarkNewString
BenchmarkNewString-24               	 7665982	       154.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkNewSLogJsonAdapter
BenchmarkNewSLogJsonAdapter-24      	 7157626	       171.9 ns/op	      64 B/op	       1 allocs/op
BenchmarkNewSLogStringAdapter
BenchmarkNewSLogStringAdapter-24    	 6995372	       173.2 ns/op	      72 B/op	       2 allocs/op
*/

func BenchmarkNewJSON(b *testing.B) {
	ll := logx.New(
		logx.WithOutput(io.Discard),
		logx.WithLevel(logx.LevelDebug),
		logx.WithFormatter(logx.NewFormatJSON()),
	)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ll.Info("sync", "id", 1, "obj", testData{})
		}
	})
}

func BenchmarkNewString(b *testing.B) {
	ll := logx.New(
		logx.WithOutput(io.Discard),
		logx.WithLevel(logx.LevelDebug),
		logx.WithFormatter(logx.NewFormatString()),
	)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ll.Info("sync", "id", 1, "obj", testData{})
		}
	})
}

func BenchmarkNewSLogJsonAdapter(b *testing.B) {
	ll := logx.NewSLogJsonAdapter(logx.WithOutput(io.Discard), logx.WithLevel(logx.LevelDebug))

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ll.Info("sync", "id", 1, "obj", testData{})
		}
	})
}

func BenchmarkNewSLogStringAdapter(b *testing.B) {
	ll := logx.NewSLogStringAdapter(logx.WithOutput(io.Discard), logx.WithLevel(logx.LevelDebug))

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ll.Info("sync", "id", 1, "obj", testData{})
		}
	})
}
