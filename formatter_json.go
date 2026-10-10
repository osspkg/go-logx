/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"

	"go.osspkg.com/ioutils/pool"
)

type FormatJSON struct{}

func NewFormatJSON() *FormatJSON {
	return &FormatJSON{}
}

type jsonContext struct {
	values map[string]any
}

const (
	jsonInitialBufferSize   = 128
	maxPooledJSONBufferSize = 16 << 10
)

// rawJSONContextValue marks values from Message.Map that are already formatted.
type rawJSONContextValue string

type jsonBuffer struct {
	data []byte
}

// poolJSONBuffer reuses output buffers and drops oversized log records.
var poolJSONBuffer = sync.Pool{New: func() any {
	return &jsonBuffer{data: make([]byte, 0, jsonInitialBufferSize)}
}}

func (v *jsonContext) Reset() {
	clear(v.values)
}

var poolJSONContext = pool.New[*jsonContext](func() *jsonContext {
	return &jsonContext{values: make(map[string]any, 10)}
})

func prepareJSONMessage(message *Message) (Message, *jsonContext) {
	encoded := *message
	if len(message.Ctx) == 0 {
		return encoded, nil
	}

	context := poolJSONContext.Get()
	for key, value := range message.Map {
		context.values[key] = rawJSONContextValue(value)
	}
	for i := 0; i < len(message.Ctx); i += 2 {
		var value any
		if i+1 < len(message.Ctx) {
			value = message.Ctx[i+1]
		}
		context.values[typing(message.Ctx[i])] = value
	}
	return encoded, context
}

func appendJSONMessage(dst []byte, message Message) ([]byte, error) {
	return appendJSONMessageWithContext(dst, message, nil)
}

func appendJSONMessageWithContext(dst []byte, message Message, context map[string]any) ([]byte, error) {
	return appendJSONMessageFields(dst, message, context, nil)
}

func appendJSONMessageWithPairs(dst []byte, message Message, pairs []any) ([]byte, error) {
	return appendJSONMessageFields(dst, message, nil, pairs)
}

func appendJSONMessageFields(dst []byte, message Message, context map[string]any, pairs []any) ([]byte, error) {
	dst = append(dst, `{"time":`...)
	var err error
	dst, err = appendJSONTime(dst, message.Time)
	if err != nil {
		return nil, err
	}
	dst = append(dst, `,"level":`...)
	dst = appendJSONString(dst, message.Level)
	dst = append(dst, `,"msg":`...)
	dst = appendJSONString(dst, message.Message)
	if pairs != nil || len(message.Map) > 0 || len(context) > 0 {
		dst = append(dst, `,"ctx":{`...)
		first := true
		switch {
		case pairs != nil:
			for i := 0; i < len(pairs); i += 2 {
				if !first {
					dst = append(dst, ',')
				}
				first = false
				var value any
				if i+1 < len(pairs) {
					value = pairs[i+1]
				}
				dst = appendJSONString(dst, typing(pairs[i]))
				dst = append(dst, ':')
				dst = appendJSONContextValue(dst, value)
			}
		case context != nil:
			for key, value := range context {
				if !first {
					dst = append(dst, ',')
				}
				first = false
				dst = appendJSONString(dst, key)
				dst = append(dst, ':')
				dst = appendJSONContextValue(dst, value)
			}
		default:
			for key, value := range message.Map {
				if !first {
					dst = append(dst, ',')
				}
				first = false
				dst = appendJSONString(dst, key)
				dst = append(dst, ':')
				dst = appendJSONString(dst, value)
			}
		}
		dst = append(dst, '}')
	}
	return append(dst, '}'), nil
}

// canAppendJSONPairs avoids building a temporary map for the common one and
// two pair cases when string keys are unique and no map fields need merging.
func canAppendJSONPairs(message *Message) bool {
	if len(message.Map) != 0 {
		return false
	}
	switch len(message.Ctx) {
	case 2:
		_, ok := message.Ctx[0].(string)
		return ok
	case 4:
		first, firstOK := message.Ctx[0].(string)
		second, secondOK := message.Ctx[2].(string)
		return firstOK && secondOK && first != second
	default:
		return false
	}
}

func appendJSONTime(dst []byte, value time.Time) ([]byte, error) {
	dst = append(dst, '"')
	start := len(dst)
	dst = value.AppendFormat(dst, time.RFC3339Nano)
	dst = append(dst, '"')

	if dst[start+4] != '-' {
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}
	_, offset := value.Zone()
	if offset >= 24*60*60 || offset <= -24*60*60 {
		return nil, errors.New("Time.MarshalJSON: timezone hour outside of range [0,23]")
	}
	return dst, nil
}

// appendJSONString appends a quoted JSON string, copying safe runs at once.
func appendJSONString(dst []byte, value string) []byte {
	dst = append(dst, '"')
	start := 0
	for i := 0; i < len(value); {
		char := value[i]
		if char < utf8.RuneSelf {
			if char >= 0x20 && char != '"' && char != '\\' && char != '<' && char != '>' && char != '&' {
				i++
				continue
			}
			dst = append(dst, value[start:i]...)
			switch char {
			case '"', '\\':
				dst = append(dst, '\\', char)
			case '\b':
				dst = append(dst, `\b`...)
			case '\f':
				dst = append(dst, `\f`...)
			case '\n':
				dst = append(dst, `\n`...)
			case '\r':
				dst = append(dst, `\r`...)
			case '\t':
				dst = append(dst, `\t`...)
			case '<', '>', '&':
				dst = append(dst, `\u003`...)
				dst = append(dst, "0123456789abcdef"[char&0x0f])
			default:
				dst = append(dst, `\u00`...)
				dst = append(dst, "0123456789abcdef"[char>>4], "0123456789abcdef"[char&0x0f])
			}
			i++
			start = i
			continue
		}

		r, size := utf8.DecodeRuneInString(value[i:])
		if r != utf8.RuneError && r != '\u2028' && r != '\u2029' {
			i += size
			continue
		}
		dst = append(dst, value[start:i]...)
		if r == utf8.RuneError {
			dst = append(dst, `\ufffd`...)
		} else {
			dst = append(dst, `\u202`...)
			dst = append(dst, "89"[r-'\u2028'])
		}
		i += size
		start = i
	}
	if start < len(value) {
		dst = append(dst, value[start:]...)
	}
	return append(dst, '"')
}

// appendJSONContextValue keeps context values quoted while formatting primitives in place.
func appendJSONContextValue(dst []byte, value any) []byte {
	switch value := value.(type) {
	case rawJSONContextValue:
		return appendJSONString(dst, string(value))
	case string:
		if isPlainString(value) {
			return appendJSONString(dst, value)
		}
	case bool:
		dst = append(dst, '"')
		dst = strconv.AppendBool(dst, value)
		return append(dst, '"')
	case int:
		dst = append(dst, '"')
		dst = strconv.AppendInt(dst, int64(value), 10)
		return append(dst, '"')
	case int8:
		dst = append(dst, '"')
		dst = strconv.AppendInt(dst, int64(value), 10)
		return append(dst, '"')
	case int16:
		dst = append(dst, '"')
		dst = strconv.AppendInt(dst, int64(value), 10)
		return append(dst, '"')
	case int32:
		dst = append(dst, '"')
		dst = strconv.AppendInt(dst, int64(value), 10)
		return append(dst, '"')
	case int64:
		dst = append(dst, '"')
		dst = strconv.AppendInt(dst, value, 10)
		return append(dst, '"')
	case uint:
		dst = append(dst, '"')
		dst = strconv.AppendUint(dst, uint64(value), 10)
		return append(dst, '"')
	case uint8:
		dst = append(dst, '"')
		dst = strconv.AppendUint(dst, uint64(value), 10)
		return append(dst, '"')
	case uint16:
		dst = append(dst, '"')
		dst = strconv.AppendUint(dst, uint64(value), 10)
		return append(dst, '"')
	case uint32:
		dst = append(dst, '"')
		dst = strconv.AppendUint(dst, uint64(value), 10)
		return append(dst, '"')
	case uint64:
		dst = append(dst, '"')
		dst = strconv.AppendUint(dst, value, 10)
		return append(dst, '"')
	case uintptr:
		dst = append(dst, '"')
		dst = strconv.AppendUint(dst, uint64(value), 10)
		return append(dst, '"')
	case float32:
		dst = append(dst, '"')
		dst = strconv.AppendFloat(dst, float64(value), 'g', -1, 32)
		return append(dst, '"')
	case float64:
		dst = append(dst, '"')
		dst = strconv.AppendFloat(dst, value, 'g', -1, 64)
		return append(dst, '"')
	case complex64:
		dst = append(dst, '"')
		dst = append(dst, strconv.FormatComplex(complex128(value), 'g', -1, 64)...)
		return append(dst, '"')
	case complex128:
		dst = append(dst, '"')
		dst = append(dst, strconv.FormatComplex(value, 'g', -1, 128)...)
		return append(dst, '"')
	case []byte:
		if isPlainJSONBytes(value) {
			dst = append(dst, '"')
			dst = append(dst, value...)
			return append(dst, '"')
		}
	}
	return appendJSONString(dst, typing(value))
}

func isPlainJSONBytes(value []byte) bool {
	for _, char := range value {
		if char < 0x20 || char > 0x7e || char == '"' || char == '\\' || char == '<' || char == '>' || char == '&' {
			return false
		}
	}
	return true
}

func (*FormatJSON) Encode(out io.Writer, m *Message) error {
	if m == nil {
		return errors.New("logx json encode: nil message")
	}

	fastPairs := canAppendJSONPairs(m)
	message := *m
	var context *jsonContext
	if !fastPairs {
		message, context = prepareJSONMessage(m)
	}
	if context != nil {
		defer poolJSONContext.Put(context)
	}

	buffer := poolJSONBuffer.Get().(*jsonBuffer)
	defer func() {
		if cap(buffer.data) > maxPooledJSONBufferSize {
			return
		}
		buffer.data = buffer.data[:0]
		poolJSONBuffer.Put(buffer)
	}()

	var encoded []byte
	var err error
	if fastPairs {
		encoded, err = appendJSONMessageWithPairs(buffer.data[:0], message, message.Ctx)
	} else if context == nil {
		encoded, err = appendJSONMessage(buffer.data[:0], message)
	} else {
		encoded, err = appendJSONMessageWithContext(buffer.data[:0], message, context.values)
	}
	if err != nil {
		return fmt.Errorf("logx json encode: %w", err)
	}
	buffer.data = append(encoded, newLine...)

	n, err := out.Write(buffer.data)
	if err != nil {
		return fmt.Errorf("logx json write: %w", err)
	}
	if n != len(buffer.data) {
		return fmt.Errorf("logx json write: %w", io.ErrShortWrite)
	}
	return nil
}
