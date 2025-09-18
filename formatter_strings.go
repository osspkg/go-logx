/*
 *  Copyright (c) 2024-2025 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"encoding"
	"fmt"
	"io"
	"strings"
	"time"

	"go.osspkg.com/ioutils/data"
)

type FormatString struct {
	delim byte
}

func NewFormatString() *FormatString {
	return &FormatString{delim: '\t'}
}

func (v *FormatString) SetDelimiter(d byte) {
	v.delim = d
}

func (v *FormatString) write(w *data.Buffer, key, value interface{}) {
	w.WriteByte('"')             //nolint:errcheck
	w.WriteString(typing(key))   //nolint:errcheck
	w.WriteString("\"=\"")       //nolint:errcheck
	w.WriteString(typing(value)) //nolint:errcheck
	w.WriteByte('"')             //nolint:errcheck
	w.WriteByte(v.delim)         //nolint:errcheck
}

func (v *FormatString) Encode(out io.Writer, m *Message) error {
	w := poolBuffer.Get()
	defer func() {
		poolBuffer.Put(w)
	}()

	v.write(w, "time", m.Time.Format(time.RFC3339))
	v.write(w, "level", m.Level)
	v.write(w, "msg", m.Message)

	if count := len(m.Ctx); count > 0 {
		if count%2 != 0 {
			m.Ctx = append(m.Ctx, nil)
			count++
		}
		for i := 0; i < count; i = i + 2 {
			v.write(w, m.Ctx[i], m.Ctx[i+1])
		}
	}
	w.Write(newLine) //nolint:errcheck
	if _, err := w.WriteTo(out); err != nil {
		return fmt.Errorf("logx string write: %w", err)
	}
	return nil
}

func typing(v interface{}) string {
	if v == nil {
		return "null"
	}
	switch vv := v.(type) {
	case error:
		v = vv.Error()
	case fmt.GoStringer:
		v = vv.GoString()
	case fmt.Stringer:
		v = vv.String()
	case encoding.TextMarshaler:
		if b, err := vv.MarshalText(); err == nil {
			v = string(b)
		}
	case encoding.BinaryMarshaler:
		if b, err := vv.MarshalBinary(); err == nil {
			v = string(b)
		}
	case []byte:
		v = string(vv)
	default:
	}

	s := fmt.Sprintf("%#v", v)
	return strings.Trim(s, "\"")
}
