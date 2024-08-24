/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"go.osspkg.com/casecheck"
)

func TestUnit_NewJSON(t *testing.T) {
	casecheck.NotNil(t, Default())

	filename, err := os.CreateTemp(os.TempDir(), "test_new_default-*.log")
	casecheck.NoError(t, err)

	SetOutput(filename)
	SetLevel(LevelDebug)
	casecheck.Equal(t, LevelDebug, GetLevel())

	go Info("async", "id", 1)
	go Warn("async", "id", 2)
	go Error("async", "id", 3)
	go Debug("async", "id", 4)

	Info("sync", "id", 1)
	Warn("sync", "id", 2)
	Error("sync", "id", 3)
	Debug("sync", "id", 4)

	Info("context1", "ip", "0.0.0.0")
	Info("context2", "nil", nil)
	Info("context3", "func", func() {})
	Info("context4", "err", fmt.Errorf("er1"))

	<-time.After(time.Second * 1)
	Close()

	casecheck.NoError(t, filename.Close())
	data, err := os.ReadFile(filename.Name())
	casecheck.NoError(t, err)
	casecheck.NoError(t, os.Remove(filename.Name()))

	sdata := string(data)
	casecheck.Contains(t, sdata, `"lvl":"INF","msg":"async","ctx":{"id":"1"}`)
	casecheck.Contains(t, sdata, `"lvl":"WRN","msg":"async","ctx":{"id":"2"}`)
	casecheck.Contains(t, sdata, `"lvl":"ERR","msg":"async","ctx":{"id":"3"}`)
	casecheck.Contains(t, sdata, `"lvl":"DBG","msg":"async","ctx":{"id":"4"}`)
	casecheck.Contains(t, sdata, `"lvl":"INF","msg":"sync","ctx":{"id":"1"}`)
	casecheck.Contains(t, sdata, `"lvl":"WRN","msg":"sync","ctx":{"id":"2"}`)
	casecheck.Contains(t, sdata, `"lvl":"ERR","msg":"sync","ctx":{"id":"3"}`)
	casecheck.Contains(t, sdata, `"lvl":"DBG","msg":"sync","ctx":{"id":"4"}`)
	casecheck.Contains(t, sdata, `"lvl":"INF","msg":"context1","ctx":{"ip":"0.0.0.0"}`)
	casecheck.Contains(t, sdata, `"lvl":"INF","msg":"context2","ctx":{"nil":"null"}`)
	casecheck.Contains(t, sdata, `"lvl":"INF","msg":"context3","ctx":{"func":"(func())(0x`)
	casecheck.Contains(t, sdata, `"lvl":"INF","msg":"context4","ctx":{"err":"er1"}`)
}

func TestUnit_NewString(t *testing.T) {
	l := New()

	casecheck.NotNil(t, l)
	l.SetFormatter(NewFormatString())

	filename, err := os.CreateTemp(os.TempDir(), "test_new_default-*.log")
	casecheck.NoError(t, err)

	l.SetOutput(filename)
	l.SetLevel(LevelDebug)
	casecheck.Equal(t, LevelDebug, l.GetLevel())

	go l.Info("async", "id", 1)
	go l.Warn("async", "id", 2)
	go l.Error("async", "id", 3)
	go l.Debug("async", "id", 4)

	l.Info("sync", "id", 1)
	l.Warn("sync", "id", 2)
	l.Error("sync", "id", 3)
	l.Debug("sync", "id", 4)

	l.Info("context1", "ip", "0.0.0.0\n")
	l.Info("context2", "nil", nil)
	l.Info("context3", "func", func() {})
	l.Info("context4", "err", fmt.Errorf("er1"))

	<-time.After(time.Second * 1)
	l.Close()

	casecheck.NoError(t, filename.Close())
	data, err := os.ReadFile(filename.Name())
	casecheck.NoError(t, err)
	casecheck.NoError(t, os.Remove(filename.Name()))

	sdata := string(data)
	casecheck.Contains(t, sdata, "lvl=INF\tmsg=\"async\"\tid=\"1\"")
	casecheck.Contains(t, sdata, "lvl=WRN\tmsg=\"async\"\tid=\"2\"")
	casecheck.Contains(t, sdata, "lvl=ERR\tmsg=\"async\"\tid=\"3\"")
	casecheck.Contains(t, sdata, "lvl=DBG\tmsg=\"async\"\tid=\"4\"")
	casecheck.Contains(t, sdata, "lvl=INF\tmsg=\"sync\"\tid=\"1\"")
	casecheck.Contains(t, sdata, "lvl=WRN\tmsg=\"sync\"\tid=\"2\"")
	casecheck.Contains(t, sdata, "lvl=ERR\tmsg=\"sync\"\tid=\"3\"")
	casecheck.Contains(t, sdata, "lvl=DBG\tmsg=\"sync\"\tid=\"4\"")
	casecheck.Contains(t, sdata, "lvl=INF\tmsg=\"context1\"\tip=\"0.0.0.0\\n\"")
	casecheck.Contains(t, sdata, "lvl=INF\tmsg=\"context2\"\tnil=\"null\"")
	casecheck.Contains(t, sdata, "lvl=INF\tmsg=\"context3\"\tfunc=\"(func())(0x")
	casecheck.Contains(t, sdata, "lvl=INF\tmsg=\"context4\"\terr=\"er1\"")
}

func BenchmarkNewJSON(b *testing.B) {
	b.ReportAllocs()

	ll := New()
	ll.SetOutput(io.Discard)
	ll.SetLevel(LevelDebug)
	ll.SetFormatter(NewFormatJSON())

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ll.Info("sync", "id", 1)
		}
	})
	ll.Close()
}

func BenchmarkNewString(b *testing.B) {
	b.ReportAllocs()

	ll := New()
	ll.SetOutput(io.Discard)
	ll.SetLevel(LevelDebug)
	ll.SetFormatter(NewFormatString())

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ll.Info("sync", "id", 1)
		}
	})
	ll.Close()
}
