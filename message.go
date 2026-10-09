/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"time"

	"go.osspkg.com/ioutils/pool"
)

//go:generate easyjson

var poolMessage = pool.New[*Message](func() *Message {
	return newMessage()
})

//easyjson:json
type Message struct {
	Time    time.Time         `json:"time" yaml:"time"`
	Level   string            `json:"level" yaml:"level"`
	Message string            `json:"msg" yaml:"msg"`
	Ctx     []any             `json:"-"`
	Map     map[string]string `json:"ctx,omitempty" yaml:"ctx,omitempty,inline"`
}

func newMessage() *Message {
	return &Message{
		Ctx: make([]any, 0, 10),
		Map: make(map[string]string, 10),
	}
}

func (v *Message) Reset() {
	clear(v.Ctx)
	clear(v.Map)
	v.Ctx = v.Ctx[:0]
}
