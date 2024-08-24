/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
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
	Level   string            `json:"lvl" yaml:"lvl"`
	Message string            `json:"msg" yaml:"msg"`
	Ctx     []interface{}     `json:"-"`
	Map     map[string]string `json:"ctx,omitempty" yaml:"ctx,omitempty,inline"`
}

func newMessage() *Message {
	return &Message{
		Ctx: make([]interface{}, 0, 10),
		Map: make(map[string]string, 10),
	}
}

func (v *Message) Reset() {
	v.Ctx = v.Ctx[:0]
	for k := range v.Map {
		delete(v.Map, k)
	}
}

func (v *Message) CtxToMap() {
	count := len(v.Ctx)
	if count == 0 {
		return
	}
	if count%2 != 0 {
		v.Ctx = append(v.Ctx, nil)
		count++
	}
	for i := 0; i < count; i = i + 2 {
		v.Map[typing(v.Ctx[i])] = typing(v.Ctx[i+1])
	}
	v.Ctx = v.Ctx[:0]
}
