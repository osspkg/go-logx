/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/mailru/easyjson/jwriter"
	"go.osspkg.com/ioutils/pool"
)

type FormatJSON struct{}

func NewFormatJSON() *FormatJSON {
	return &FormatJSON{}
}

type jsonContext struct {
	values map[string]string
}

const (
	easyJSONInitialBufferSize = 128
	easyJSONPoolMinBufferSize = 512
)

func (v *jsonContext) Reset() {
	clear(v.values)
}

var poolJSONContext = pool.New[*jsonContext](func() *jsonContext {
	return &jsonContext{values: make(map[string]string, 10)}
})

func prepareJSONMessage(message *Message) (Message, *jsonContext) {
	encoded := *message
	if len(message.Ctx) == 0 {
		return encoded, nil
	}

	context := poolJSONContext.Get()
	for key, value := range message.Map {
		context.values[key] = value
	}
	for i := 0; i < len(message.Ctx); i += 2 {
		var value any
		if i+1 < len(message.Ctx) {
			value = message.Ctx[i+1]
		}
		context.values[typing(message.Ctx[i])] = typing(value)
	}
	encoded.Map = context.values
	return encoded, context
}

// encodeJSONMessage mirrors Message.MarshalEasyJSON while writing Time without
// allocating the intermediate slice returned by time.Time.MarshalJSON.
func encodeJSONMessage(out *jwriter.Writer, message Message) {
	out.RawByte('{')
	out.RawString(`"time":`)
	writeJSONTime(out, message.Time)
	if out.Error != nil {
		return
	}
	out.RawString(`,"level":`)
	out.String(message.Level)
	out.RawString(`,"msg":`)
	out.String(message.Message)
	if len(message.Map) > 0 {
		out.RawString(`,"ctx":{`)
		first := true
		for key, value := range message.Map {
			if !first {
				out.RawByte(',')
			}
			first = false
			out.String(key)
			out.RawByte(':')
			out.String(value)
		}
		out.RawByte('}')
	}
	out.RawByte('}')
}

func writeJSONTime(out *jwriter.Writer, value time.Time) {
	out.Buffer.EnsureSpace(37)
	out.Buffer.Buf = append(out.Buffer.Buf, '"')
	start := len(out.Buffer.Buf)
	out.Buffer.Buf = value.AppendFormat(out.Buffer.Buf, time.RFC3339Nano)
	out.Buffer.Buf = append(out.Buffer.Buf, '"')

	if out.Buffer.Buf[start+4] != '-' {
		out.Error = errors.New("Time.MarshalJSON: year outside of range [0,9999]")
		return
	}
	_, offset := value.Zone()
	if offset >= 24*60*60 || offset <= -24*60*60 {
		out.Error = errors.New("Time.MarshalJSON: timezone hour outside of range [0,23]")
	}
}

func dumpJSON(out io.Writer, writer *jwriter.Writer) error {
	if writer.Buffer.Size() <= easyJSONInitialBufferSize && cap(writer.Buffer.Buf) < easyJSONPoolMinBufferSize {
		data := writer.Buffer.BuildBytes()
		n, err := out.Write(data)
		if err != nil {
			return err
		}
		if n != len(data) {
			return io.ErrShortWrite
		}
		return nil
	}
	_, err := writer.DumpTo(out)
	return err
}

func (*FormatJSON) Encode(out io.Writer, m *Message) error {
	if m == nil {
		return errors.New("logx json encode: nil message")
	}

	message, context := prepareJSONMessage(m)
	if context != nil {
		defer poolJSONContext.Put(context)
	}

	var w jwriter.Writer
	encodeJSONMessage(&w, message)
	if w.Error != nil {
		encodeErr := w.Error
		_, _ = w.DumpTo(io.Discard)
		return fmt.Errorf("logx json encode: %w", encodeErr)
	}

	if err := dumpJSON(out, &w); err != nil {
		return fmt.Errorf("logx json write: %w", err)
	}

	if _, err := out.Write(newLine); err != nil {
		return fmt.Errorf("logx json write new line: %w", err)
	}

	return nil
}
