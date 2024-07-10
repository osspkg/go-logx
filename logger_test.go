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
	"go.osspkg.com/syncing"
)

func TestUnit_NewJSON(t *testing.T) {
	casecheck.NotNil(t, Default())

	filename, err := os.CreateTemp(os.TempDir(), "test_new_default-*.log")
	casecheck.NoError(t, err)

	SetOutput(filename)
	SetLevel(LevelDebug)
	casecheck.Equal(t, LevelDebug, GetLevel())

	go Infof("async %d", 1)
	go Warnf("async %d", 2)
	go Errorf("async %d", 3)
	go Debugf("async %d", 4)

	Infof("sync %d", 1)
	Warnf("sync %d", 2)
	Errorf("sync %d", 3)
	Debugf("sync %d", 4)

	WithFields(Fields{"ip": "0.0.0.0"}).Infof("context1")
	WithFields(Fields{"nil": nil}).Infof("context2")
	WithFields(Fields{"func": func() {}}).Infof("context3")

	WithField("ip", "0.0.0.0").Infof("context4")
	WithField("nil", nil).Infof("context5")
	WithField("func", func() {}).Infof("context6")

	WithError("err", nil).Infof("context7")
	WithError("err", fmt.Errorf("er1")).Infof("context8")

	<-time.After(time.Second * 1)
	Close()

	casecheck.NoError(t, filename.Close())
	data, err := os.ReadFile(filename.Name())
	casecheck.NoError(t, err)
	casecheck.NoError(t, os.Remove(filename.Name()))

	sdata := string(data)
	casecheck.Contains(t, sdata, `"lvl":"INF","msg":"async 1"`)
	casecheck.Contains(t, sdata, `"lvl":"WRN","msg":"async 2"`)
	casecheck.Contains(t, sdata, `"lvl":"ERR","msg":"async 3"`)
	casecheck.Contains(t, sdata, `"lvl":"DBG","msg":"async 4"`)
	casecheck.Contains(t, sdata, `"lvl":"INF","msg":"sync 1"`)
	casecheck.Contains(t, sdata, `"lvl":"WRN","msg":"sync 2"`)
	casecheck.Contains(t, sdata, `"lvl":"ERR","msg":"sync 3"`)
	casecheck.Contains(t, sdata, `"msg":"context1","ctx":{"ip":"0.0.0.0"}`)
	casecheck.Contains(t, sdata, `"msg":"context2","ctx":{"nil":null}`)
	casecheck.Contains(t, sdata, `"msg":"context3","ctx":{"func":"unsupported field value: (func())`)
	casecheck.Contains(t, sdata, `"msg":"context4","ctx":{"ip":"0.0.0.0"}`)
	casecheck.Contains(t, sdata, `"msg":"context5","ctx":{"nil":null}`)
	casecheck.Contains(t, sdata, `"msg":"context6","ctx":{"func":"unsupported field value: (func())`)
	casecheck.Contains(t, sdata, `"msg":"context7","ctx":{"err":null}`)
	casecheck.Contains(t, sdata, `"msg":"context8","ctx":{"err":"er1"}`)
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

	go l.Infof("async %d", 1)
	go l.Warnf("async %d", 2)
	go l.Errorf("async %d", 3)
	go l.Debugf("async %d", 4)

	l.Infof("sync %d", 1)
	l.Warnf("sync %d", 2)
	l.Errorf("sync %d", 3)
	l.Debugf("sync %d", 4)

	l.WithFields(Fields{"ip": "0.0.0.0"}).Infof("context1")
	l.WithFields(Fields{"nil": nil}).Infof("context2")
	l.WithFields(Fields{"func": func() {}}).Infof("context3")

	l.WithField("ip", "0.0.0.0").Infof("context4")
	l.WithField("nil", nil).Infof("context5")
	l.WithField("func", func() {}).Infof("context6")

	l.WithError("err", nil).Infof("context7")
	l.WithError("err", fmt.Errorf("er1")).Infof("context8")

	<-time.After(time.Second * 1)
	l.Close()

	casecheck.NoError(t, filename.Close())
	data, err := os.ReadFile(filename.Name())
	casecheck.NoError(t, err)
	casecheck.NoError(t, os.Remove(filename.Name()))

	sdata := string(data)
	casecheck.Contains(t, sdata, "lvl=INF	msg=\"async 1\"")
	casecheck.Contains(t, sdata, "lvl=WRN	msg=\"async 2\"")
	casecheck.Contains(t, sdata, "lvl=ERR	msg=\"async 3\"")
	casecheck.Contains(t, sdata, "lvl=DBG	msg=\"async 4\"")
	casecheck.Contains(t, sdata, "lvl=INF	msg=\"sync 1\"")
	casecheck.Contains(t, sdata, "lvl=WRN	msg=\"sync 2\"")
	casecheck.Contains(t, sdata, "lvl=ERR	msg=\"sync 3\"")
	casecheck.Contains(t, sdata, "lvl=DBG	msg=\"sync 4\"")
	casecheck.Contains(t, sdata, "lvl=INF	msg=\"context1\"	ip=\"0.0.0.0\"")
	casecheck.Contains(t, sdata, "lvl=INF	msg=\"context2\"	nil=<nil>")
	casecheck.Contains(t, sdata, "lvl=INF	msg=\"context3\"	func=\"unsupported field value: (func())(0x")
	casecheck.Contains(t, sdata, "lvl=INF	msg=\"context4\"	ip=\"0.0.0.0\"")
	casecheck.Contains(t, sdata, "lvl=INF	msg=\"context5\"	nil=<nil>")
	casecheck.Contains(t, sdata, "lvl=INF	msg=\"context6\"	func=\"unsupported field value: (func())(0x")
	casecheck.Contains(t, sdata, "lvl=INF	msg=\"context7\"	err=<nil>")
	casecheck.Contains(t, sdata, "lvl=INF	msg=\"context8\"	err=\"er1\"")
}

func BenchmarkNewJSON(b *testing.B) {
	b.ReportAllocs()

	ll := New()
	ll.SetOutput(io.Discard)
	ll.SetLevel(LevelDebug)
	ll.SetFormatter(NewFormatJSON())
	wg := syncing.NewGroup()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wg.Background(func() {
				ll.WithFields(Fields{"a": "b"}).Infof("hello")
				ll.WithField("a", "b").Infof("hello")
				ll.WithError("a", fmt.Errorf("b")).Infof("hello")
			})
		}
	})
	wg.Wait()
	ll.Close()
}

func BenchmarkNewString(b *testing.B) {
	b.ReportAllocs()

	ll := New()
	ll.SetOutput(io.Discard)
	ll.SetLevel(LevelDebug)
	ll.SetFormatter(NewFormatString())
	wg := syncing.NewGroup()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wg.Background(func() {
				ll.WithFields(Fields{"a": "b"}).Infof("hello")
				ll.WithField("a", "b").Infof("hello")
				ll.WithError("a", fmt.Errorf("b")).Infof("hello")
			})
		}
	})
	wg.Wait()
	ll.Close()
}
