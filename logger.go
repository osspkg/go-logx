/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
)

// Log base model
type Log struct {
	level     uint32
	writer    io.Writer
	formatter Formatter
}

// New init new logger
func New() *Log {
	return &Log{
		level:     LevelError,
		writer:    os.Stdout,
		formatter: NewFormatJSON(),
	}
}

func (l *Log) writeMessage(level uint32, call func(v *Message)) {
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

	err := l.formatter.Encode(l.writer, m)
	if err != nil {
		fmt.Println(err)
	}

}

// SetOutput change writer
func (l *Log) SetOutput(out io.Writer) {
	l.writer = out
}

func (l *Log) SetFormatter(f Formatter) {
	l.formatter = f
}

// SetLevel change Log level
func (l *Log) SetLevel(v uint32) {
	atomic.StoreUint32(&l.level, v)
}

// GetLevel getting Log level
func (l *Log) GetLevel() uint32 {
	return atomic.LoadUint32(&l.level)
}

func (l *Log) Info(message string, args ...interface{}) {
	l.writeMessage(LevelInfo, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
}

func (l *Log) Warn(message string, args ...interface{}) {
	l.writeMessage(LevelWarn, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
}

func (l *Log) Error(message string, args ...interface{}) {
	l.writeMessage(LevelError, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
}

func (l *Log) Debug(message string, args ...interface{}) {
	l.writeMessage(LevelDebug, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
}

func (l *Log) Fatal(message string, args ...interface{}) {
	l.writeMessage(levelFatal, func(v *Message) {
		v.Message = message
		v.Ctx = append(v.Ctx, args...)
	})
	os.Exit(1)
}
