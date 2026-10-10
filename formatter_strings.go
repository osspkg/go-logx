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

func (v *FormatString) writeString(w *pooledBuffer, key, value string) {
	w.WriteByte('"') //nolint:errcheck
	writeQuotedString(w, key)
	w.WriteString("\"=\"") //nolint:errcheck
	writeQuotedString(w, value)
	w.WriteByte('"')     //nolint:errcheck
	w.WriteByte(v.delim) //nolint:errcheck
}

func writeFormatValue(w *pooledBuffer, value any) {
	if value == nil {
		w.WriteString("null") //nolint:errcheck
		return
	}
	if encoded, ok := appendScalarValue(w.AvailableBuffer(), value); ok {
		w.Write(encoded) //nolint:errcheck
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
	case []byte:
		writeQuotedString(w, string(v))
	case complex64:
		w.WriteString(strconv.FormatComplex(complex128(v), 'g', -1, 64)) //nolint:errcheck
	case complex128:
		w.WriteString(strconv.FormatComplex(v, 'g', -1, 128)) //nolint:errcheck
	default:
		writeFallback(w, value)
	}
}

func writeQuotedString(w *pooledBuffer, value string) {
	if isPlainString(value) {
		w.WriteString(value) //nolint:errcheck
		return
	}
	quoted := strconv.AppendQuote(w.AvailableBuffer(), value)
	w.Write(quoted[1 : len(quoted)-1]) //nolint:errcheck
}

// appendScalarValue formats built-in numeric values directly into dst. Both
// the formatter and typing use it to share the allocation-free conversion path.
func appendScalarValue(dst []byte, value any) ([]byte, bool) {
	switch value := value.(type) {
	case bool:
		return strconv.AppendBool(dst, value), true
	case int:
		return strconv.AppendInt(dst, int64(value), 10), true
	case int8:
		return strconv.AppendInt(dst, int64(value), 10), true
	case int16:
		return strconv.AppendInt(dst, int64(value), 10), true
	case int32:
		return strconv.AppendInt(dst, int64(value), 10), true
	case int64:
		return strconv.AppendInt(dst, value, 10), true
	case uint:
		return strconv.AppendUint(dst, uint64(value), 10), true
	case uint8:
		return strconv.AppendUint(dst, uint64(value), 10), true
	case uint16:
		return strconv.AppendUint(dst, uint64(value), 10), true
	case uint32:
		return strconv.AppendUint(dst, uint64(value), 10), true
	case uint64:
		return strconv.AppendUint(dst, value, 10), true
	case uintptr:
		return strconv.AppendUint(dst, uint64(value), 10), true
	case float32:
		return strconv.AppendFloat(dst, float64(value), 'g', -1, 32), true
	case float64:
		return strconv.AppendFloat(dst, value, 'g', -1, 64), true
	default:
		return dst, false
	}
}

func writeFallback(w *pooledBuffer, value any) {
	formatted := fmt.Appendf(w.AvailableBuffer(), "%#v", value)
	start, end := 0, len(formatted)
	for start < end && formatted[start] == '"' {
		start++
	}
	for end > start && formatted[end-1] == '"' {
		end--
	}
	w.Write(formatted[start:end]) //nolint:errcheck
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
	v.writeString(w, "level", m.Level)
	v.writeString(w, "msg", m.Message)

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
	if value, ok := v.(string); ok {
		return typingString(value)
	}
	var buffer [128]byte
	if encoded, ok := appendScalarValue(buffer[:0], v); ok {
		return string(encoded)
	}
	switch value := v.(type) {
	case complex64:
		return strconv.FormatComplex(complex128(value), 'g', -1, 64)
	case complex128:
		return strconv.FormatComplex(value, 'g', -1, 128)
	case error:
		return typingString(value.Error())
	case fmt.GoStringer:
		return typingString(value.GoString())
	case fmt.Stringer:
		return typingString(value.String())
	case encoding.TextMarshaler:
		if data, err := value.MarshalText(); err == nil {
			return typingString(string(data))
		}
	case encoding.BinaryMarshaler:
		if data, err := value.MarshalBinary(); err == nil {
			return typingString(string(data))
		}
	case []byte:
		return typingString(string(value))
	}

	s := fmt.Sprintf("%#v", v)
	return strings.Trim(s, "\"")
}

func typingString(value string) string {
	if isPlainString(value) {
		return value
	}
	s := fmt.Sprintf("%#v", value)
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
