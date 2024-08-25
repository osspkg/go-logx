package logx

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"
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

func (v *FormatString) write(w *bytes.Buffer, key, value interface{}) {
	w.WriteByte('"')
	w.WriteString(typing(key))
	w.WriteString("\"=\"")
	w.WriteString(typing(value))
	w.WriteByte('"')
	w.WriteByte(v.delim)
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
	w.Write(newLine)
	if _, err := w.WriteTo(out); err != nil {
		return fmt.Errorf("logx string write: %w", err)
	}
	return nil
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
