/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"io"
	"os"
	"sync/atomic"
	"time"

	"go.osspkg.com/syncing"
)

// log base model
type log struct {
	level     uint32
	writer    io.Writer
	formatter Formatter
	channel   chan []byte
	mux       syncing.Lock
	wg        syncing.Group
	closed    bool
}

// New init new logger
func New() Logger {
	object := &log{
		level:     LevelError,
		writer:    os.Stdout,
		formatter: NewFormatJSON(),
		channel:   make(chan []byte, 100000),
		mux:       syncing.NewLock(),
		wg:        syncing.NewGroup(),
		closed:    false,
	}
	object.wg.Background(func() {
		object.queue()
	})
	return object
}

func (l *log) SendMessage(level uint32, call func(v *Message)) {
	if l.GetLevel() < level {
		return
	}

	m := poolMessage.Get()
	defer func() {
		poolMessage.Put(m)
	}()

	call(m)

	lvl, ok := levels[level]
	if !ok {
		lvl = "UNK"
	}
	m.Level, m.Time = lvl, time.Now()

	l.mux.RLock(func() {
		if l.closed {
			return
		}

		b, err := l.formatter.Encode(m)
		if err != nil {
			b = []byte(err.Error())
		}

		select {
		case l.channel <- b:
		default:
		}
	})
}

func (l *log) queue() {
	for b := range l.channel {
		l.mux.RLock(func() {
			l.writer.Write(b) //nolint:errcheck
		})
	}
}

// Close waiting for all messages to finish recording
func (l *log) Close() {
	l.mux.Lock(func() {
		l.closed = true
		close(l.channel)
	})
	l.wg.Wait()
}

// SetOutput change writer
func (l *log) SetOutput(out io.Writer) {
	l.mux.Lock(func() {
		l.writer = out
	})
}

func (l *log) SetFormatter(f Formatter) {
	l.mux.Lock(func() {
		l.formatter = f
	})
}

// SetLevel change log level
func (l *log) SetLevel(v uint32) {
	atomic.StoreUint32(&l.level, v)
}

// GetLevel getting log level
func (l *log) GetLevel() uint32 {
	return atomic.LoadUint32(&l.level)
}

func (l *log) Info(message string, args ...interface{}) {
	l.SendMessage(LevelInfo, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
}

func (l *log) Warn(message string, args ...interface{}) {
	l.SendMessage(LevelWarn, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
}

func (l *log) Error(message string, args ...interface{}) {
	l.SendMessage(LevelError, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
}

func (l *log) Debug(message string, args ...interface{}) {
	l.SendMessage(LevelDebug, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
}

func (l *log) Fatal(message string, args ...interface{}) {
	l.SendMessage(levelFatal, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
	l.Close()
	os.Exit(1)
}
