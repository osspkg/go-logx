/*
 *  Copyright (c) 2024-2025 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"go.osspkg.com/ioutils/data"
	"go.osspkg.com/ioutils/pool"
)

type FormatJSON struct{}

func NewFormatJSON() *FormatJSON {
	return &FormatJSON{}
}

func (*FormatJSON) Encode(out io.Writer, m *Message) error {
	m.CtxToMap()

	w := poolJson.Get()
	defer func() {
		poolJson.Put(w)
	}()

	if err := w.Encoder.Encode(m); err != nil {
		return fmt.Errorf("logx json encode: %w", err)
	}

	if _, err := w.Buffer.Seek(-1, io.SeekEnd); err != nil {
		return fmt.Errorf("logx json seek: %w", err)
	}

	if b := w.Buffer.Next(1); !bytes.Equal(b, newLine) {
		if _, err := w.Buffer.Write(newLine); err != nil {
			return fmt.Errorf("logx json write new line: %w", err)
		}
	}

	if _, err := w.Buffer.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("logx json seek: %w", err)
	}

	if _, err := w.Buffer.WriteTo(out); err != nil {
		return fmt.Errorf("logx json write: %w", err)
	}

	return nil
}

var poolJson = pool.New[*jsonWriter](func() *jsonWriter {
	obj := &jsonWriter{
		Buffer: data.NewBuffer(1024),
	}
	obj.Encoder = json.NewEncoder(obj.Buffer)
	return obj
})

type jsonWriter struct {
	Buffer  *data.Buffer
	Encoder *json.Encoder
}

func (v *jsonWriter) Reset() {
	v.Buffer.Reset()
}
