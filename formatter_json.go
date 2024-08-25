package logx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

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
	w.Buffer.Write(newLine)
	if _, err := w.Buffer.WriteTo(out); err != nil {
		return fmt.Errorf("logx json write: %w", err)
	}
	return nil
}

var poolJson = pool.New[*jsonWriter](func() *jsonWriter {
	obj := &jsonWriter{
		Buffer: bytes.NewBuffer(nil),
	}
	obj.Encoder = json.NewEncoder(obj.Buffer)
	return obj
})

type jsonWriter struct {
	Buffer  *bytes.Buffer
	Encoder *json.Encoder
}

func (v *jsonWriter) Reset() {
	v.Buffer.Reset()
}
