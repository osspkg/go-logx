/*
 *  Copyright (c) 2024-2025 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx_test

import (
	"bytes"
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

	wg := syncing.NewGroup()
	buff := newMockWriter()

	logx.SetOutput(buff)
	logx.SetLevel(logx.LevelDebug)

	wg.Background(func() { logx.Info("async", "id", 1) })
	wg.Background(func() { logx.Warn("async", "id", 2) })
	wg.Background(func() { logx.Error("async", "id", 3) })
	wg.Background(func() { logx.Debug("async", "id", 4) })

	logx.Info("sync", "id", 1)
	logx.Warn("sync", "id", 2)
	logx.Error("sync", "id", 3)
	logx.Debug("sync", "id", 4)

	logx.Info("context1", "ip", "0.0.0.0")
	logx.Info("context2", "nil", nil)
	logx.Info("context3", "func", func() {})
	logx.Info("context4", "err", fmt.Errorf("er1"))

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
}

func TestUnit_NewString(t *testing.T) {
	wg := syncing.NewGroup()
	buff := newMockWriter()

	l := logx.New()
	casecheck.NotNil(t, l)
	l.SetFormatter(logx.NewFormatString())
	l.SetOutput(buff)
	l.SetLevel(logx.LevelDebug)
	casecheck.Equal(t, logx.LevelDebug, l.GetLevel())

	wg.Background(func() { l.Info("async", "id", 1) })
	wg.Background(func() { l.Warn("async", "id", 2) })
	wg.Background(func() { l.Error("async", "id", 3) })
	wg.Background(func() { l.Debug("async", "id", 4) })

	l.Info("sync", "id", 1)
	l.Warn("sync", "id", 2)
	l.Error("sync", "id", 3)
	l.Debug("sync", "id", 4)

	l.Info("context1", "ip", "0.0.0.0\n")
	l.Info("context2", "nil", nil)
	l.Info("context3", "func", func() {})
	l.Info("context4", "err", fmt.Errorf("er1"))

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
}

func TestUnit_NewSlog(t *testing.T) {
	wg := syncing.NewGroup()
	buff := newMockWriter()

	l := logx.NewSLogJsonAdapter()
	casecheck.NotNil(t, l)
	l.SetFormatter(logx.NewFormatString())
	l.SetOutput(buff)
	l.SetLevel(logx.LevelDebug)

	wg.Background(func() { l.Info("async", "id", 1) })
	wg.Background(func() { l.Warn("async", "id", 2) })
	wg.Background(func() { l.Error("async", "id", 3) })
	wg.Background(func() { l.Debug("async", "id", 4) })

	l.Info("sync", "id", 1)
	l.Warn("sync", "id", 2)
	l.Error("sync", "id", 3)
	l.Debug("sync", "id", 4)

	l.Info("context1", "ip", "0.0.0.0\n")
	l.Info("context2", "nil", nil)
	l.Info("context3", "func", func() {})
	l.Info("context4", "err", fmt.Errorf("er1"))

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
}

func BenchmarkNewJSON(b *testing.B) {
	ll := logx.New()
	ll.SetOutput(io.Discard)
	ll.SetLevel(logx.LevelDebug)
	ll.SetFormatter(logx.NewFormatJSON())

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ll.Info("sync", "id", 1)
		}
	})
}

func BenchmarkNewString(b *testing.B) {
	ll := logx.New()
	ll.SetOutput(io.Discard)
	ll.SetLevel(logx.LevelDebug)
	ll.SetFormatter(logx.NewFormatString())

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ll.Info("sync", "id", 1)
		}
	})
}

func BenchmarkNewSLogJsonAdapter(b *testing.B) {
	ll := logx.NewSLogJsonAdapter()
	ll.SetOutput(io.Discard)
	ll.SetLevel(logx.LevelDebug)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ll.Info("sync", "id", 1)
		}
	})
}

func BenchmarkNewSLogStringAdapter(b *testing.B) {
	ll := logx.NewSLogStringAdapter()
	ll.SetOutput(io.Discard)
	ll.SetLevel(logx.LevelDebug)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ll.Info("sync", "id", 1)
		}
	})
}
