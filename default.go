/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"io"
)

var std Logger = New()

func SetDefault(log Logger) {
	std = log
}

// Default logger
func Default() Logger {
	return std
}

// SetOutput change writer
func SetOutput(out io.Writer) {
	std.SetOutput(out)
}

func SetFormatter(f Formatter) {
	std.SetFormatter(f)
}

// SetLevel change Log level
func SetLevel(v uint32) {
	std.SetLevel(v)
}

// Info message
func Info(format string, args ...interface{}) {
	std.Info(format, args...)
}

// Warn message
func Warn(format string, args ...interface{}) {
	std.Warn(format, args...)
}

// Error message
func Error(format string, args ...interface{}) {
	std.Error(format, args...)
}

// Debug message
func Debug(format string, args ...interface{}) {
	std.Debug(format, args...)
}

// Fatal message and exit
func Fatal(format string, args ...interface{}) {
	std.Fatal(format, args...)
}
