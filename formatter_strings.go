/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"encoding"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type FormatString struct {
	delim byte
}

// FormatStringOption configures a string formatter when it is created.
type FormatStringOption func(*FormatString)

// NewFormatString creates a key/value formatter. It separates fields with tabs
// unless another delimiter is supplied.
func NewFormatString(options ...FormatStringOption) *FormatString {
	formatter := &FormatString{delim: '\t'}
	for _, option := range options {
		if option != nil {
			option(formatter)
		}
	}
	return formatter
}

// WithDelimiter sets the separator used between formatted fields.
func WithDelimiter(delimiter byte) FormatStringOption {
	return func(formatter *FormatString) {
		formatter.delim = delimiter
	}
}

func (v *FormatString) write(w *pooledBuffer, key, value any) {
	w.WriteByte('"') //nolint:errcheck
	writeFormatValue(w, key)
	w.WriteString("\"=\"") //nolint:errcheck
	writeFormatValue(w, value)
	w.WriteByte('"')     //nolint:errcheck
	w.WriteByte(v.delim) //nolint:errcheck
}

func writeFormatValue(w *pooledBuffer, value any) {
	if value == nil {
		w.WriteString("null") //nolint:errcheck
		return
	}
	switch v := value.(type) {
	case error:
		writeQuotedString(w, v.Error())
	case fmt.GoStringer:
		writeQuotedString(w, v.GoString())
	case fmt.Stringer:
		writeQuotedString(w, v.String())
	case encoding.TextMarshaler:
		if data, err := v.MarshalText(); err == nil {
			writeQuotedString(w, string(data))
		} else {
			writeFallback(w, value)
		}
	case encoding.BinaryMarshaler:
		if data, err := v.MarshalBinary(); err == nil {
			writeQuotedString(w, string(data))
		} else {
			writeFallback(w, value)
		}
	case string:
		writeQuotedString(w, v)
	case bool:
		w.Write(strconv.AppendBool(w.AvailableBuffer(), v)) //nolint:errcheck
	case int:
		w.Write(strconv.AppendInt(w.AvailableBuffer(), int64(v), 10)) //nolint:errcheck
	case int8:
		w.Write(strconv.AppendInt(w.AvailableBuffer(), int64(v), 10)) //nolint:errcheck
	case int16:
		w.Write(strconv.AppendInt(w.AvailableBuffer(), int64(v), 10)) //nolint:errcheck
	case int32:
		w.Write(strconv.AppendInt(w.AvailableBuffer(), int64(v), 10)) //nolint:errcheck
	case int64:
		w.Write(strconv.AppendInt(w.AvailableBuffer(), v, 10)) //nolint:errcheck
	case uint:
		writeUintDecimal(w, uint64(v))
	case uint8:
		writeUintDecimal(w, uint64(v))
	case uint16:
		writeUintDecimal(w, uint64(v))
	case uint32:
		writeUintDecimal(w, uint64(v))
	case uint64:
		writeUintDecimal(w, v)
	case uintptr:
		writeUintDecimal(w, uint64(v))
	case float32:
		w.Write(strconv.AppendFloat(w.AvailableBuffer(), float64(v), 'g', -1, 32)) //nolint:errcheck
	case float64:
		w.Write(strconv.AppendFloat(w.AvailableBuffer(), v, 'g', -1, 64)) //nolint:errcheck
	case []byte:
		writeQuotedString(w, string(v))
	default:
		writeFallback(w, value)
	}
}

func writeQuotedString(w *pooledBuffer, value string) {
	quoted := strconv.AppendQuote(w.AvailableBuffer(), value)
	w.Write(quoted[1 : len(quoted)-1]) //nolint:errcheck
}

func writeUintDecimal(w *pooledBuffer, value uint64) {
	w.Write(strconv.AppendUint(w.AvailableBuffer(), value, 10)) //nolint:errcheck
}

func writeFallback(w *pooledBuffer, value any) {
	formatted := fmt.Sprintf("%#v", value)
	w.WriteString(strings.Trim(formatted, "\"")) //nolint:errcheck
}

func (v *FormatString) writeTime(w *pooledBuffer, value time.Time) {
	w.WriteString("\"time\"=\"")                                   //nolint:errcheck
	w.Write(value.AppendFormat(w.AvailableBuffer(), time.RFC3339)) //nolint:errcheck
	w.WriteByte('"')                                               //nolint:errcheck
	w.WriteByte(v.delim)                                           //nolint:errcheck
}

func (v *FormatString) Encode(out io.Writer, m *Message) error {
	w := poolBuffer.Get()
	defer poolBuffer.Put(w)

	v.writeTime(w, m.Time)
	v.write(w, "level", m.Level)
	v.write(w, "msg", m.Message)

	if count := len(m.Ctx); count > 0 {
		for i := 0; i < count; i += 2 {
			var value any
			if i+1 < count {
				value = m.Ctx[i+1]
			}
			v.write(w, m.Ctx[i], value)
		}
	}
	w.Write(newLine) //nolint:errcheck

	if _, err := w.WriteTo(out); err != nil {
		return fmt.Errorf("logx string write: %w", err)
	}
	return nil
}

func typing(v any) string {
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

	switch value := v.(type) {
	case string:
		if isPlainString(value) {
			return value
		}
	case bool:
		return strconv.FormatBool(value)
	case int:
		return strconv.FormatInt(int64(value), 10)
	case int8:
		return strconv.FormatInt(int64(value), 10)
	case int16:
		return strconv.FormatInt(int64(value), 10)
	case int32:
		return strconv.FormatInt(int64(value), 10)
	case int64:
		return strconv.FormatInt(value, 10)
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(value, 10)
	case uintptr:
		return strconv.FormatUint(uint64(value), 10)
	case float32:
		return strconv.FormatFloat(float64(value), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(value, 'g', -1, 64)
	case complex64:
		return strconv.FormatComplex(complex128(value), 'g', -1, 64)
	case complex128:
		return strconv.FormatComplex(value, 'g', -1, 128)
	}

	s := fmt.Sprintf("%#v", v)
	return strings.Trim(s, "\"")
}

func isPlainString(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] < 0x20 || value[i] > 0x7e || value[i] == '"' || value[i] == '\\' {
			return false
		}
	}
	return true
}
