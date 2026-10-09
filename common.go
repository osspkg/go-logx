/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

// Package logx provides leveled logging with JSON, text, and slog output.
package logx

const (
	levelFatal uint32 = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
)

var levels = map[uint32]string{
	levelFatal: "FATAL",
	LevelError: "ERROR",
	LevelWarn:  "WARN",
	LevelInfo:  "INFO",
	LevelDebug: "DEBUG",
}

// Logger base interface
type Logger interface {
	Fatal(message string, args ...any)
	Error(message string, args ...any)
	Warn(message string, args ...any)
	Info(message string, args ...any)
	Debug(message string, args ...any)
}
