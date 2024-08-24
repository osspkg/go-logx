/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.osspkg.com/ioutils/pool"
)

type Formatter interface {
	Encode(m *Message) ([]byte, error)
}

var newLine = []byte("\n")

// //////////////////////////////////////////////////////////////////////////////

type FormatJSON struct{}

func NewFormatJSON() *FormatJSON {
	return &FormatJSON{}
}

func (*FormatJSON) Encode(m *Message) ([]byte, error) {
	m.CtxToMap()
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// //////////////////////////////////////////////////////////////////////////////

var poolBuff = pool.New[*bytes.Buffer](func() *bytes.Buffer {
	return bytes.NewBuffer(make([]byte, 0, 1024))
})

type FormatString struct {
	delim string
}

func NewFormatString() *FormatString {
	return &FormatString{delim: "\t"}
}

func (v *FormatString) SetDelimiter(d string) {
	v.delim = d
}

func (v *FormatString) Encode(m *Message) ([]byte, error) {
	buff := poolBuff.Get()
	defer func() {
		poolBuff.Put(buff)
	}()

	fmt.Fprintf(buff, "time=%s%slvl=%s%smsg=%#v",
		m.Time.Format(time.RFC3339), v.delim, m.Level, v.delim, m.Message)

	if count := len(m.Ctx); count > 0 {
		if count%2 != 0 {
			m.Ctx = append(m.Ctx, nil)
			count++
		}
		for i := 0; i < count; i = i + 2 {
			fmt.Fprintf(buff, "%s%s=\"%s\"", v.delim, typing(m.Ctx[i]), typing(m.Ctx[i+1]))
		}
	}
	buff.Write(newLine)

	return append(make([]byte, 0, buff.Len()), buff.Bytes()...), nil
}

func typing(v interface{}) (s string) {
	if v == nil {
		s = "null"
		return
	}
	switch vv := v.(type) {
	case error:
		s = vv.Error()
	case fmt.GoStringer:
		s = vv.GoString()
	case fmt.Stringer:
		s = vv.String()
	default:
		s = fmt.Sprintf("%#v", v)
	}
	s = strings.Trim(s, "\"")
	return
}
