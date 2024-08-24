/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import "io"

const (
	levelFatal uint32 = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
)

var levels = map[uint32]string{
	levelFatal: "FAT",
	LevelError: "ERR",
	LevelWarn:  "WRN",
	LevelInfo:  "INF",
	LevelDebug: "DBG",
}

type Sender interface {
	SendMessage(level uint32, call func(v *Message))
	Close()
}

// Writer interface
type Writer interface {
	Fatal(format string, args ...interface{})
	Error(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Info(format string, args ...interface{})
	Debug(format string, args ...interface{})
}

// Logger base interface
type Logger interface {
	SetOutput(out io.Writer)
	SetFormatter(f Formatter)
	SetLevel(v uint32)
	GetLevel() uint32
	Close()

	Writer
}
