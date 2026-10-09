/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"bytes"
	"io"

	"go.osspkg.com/ioutils/pool"
)

var newLine = []byte("\n")

const maxPooledBufferSize = 100_000

type pooledBuffer struct {
	bytes.Buffer
}

func (v *pooledBuffer) Reset() {
	if v.Cap() > maxPooledBufferSize {
		v.Buffer = *bytes.NewBuffer(make([]byte, 0, 1024))
		return
	}
	v.Buffer.Reset()
}

var poolBuffer = pool.New[*pooledBuffer](func() *pooledBuffer {
	return &pooledBuffer{Buffer: *bytes.NewBuffer(make([]byte, 0, 1024))}
})

type Formatter interface {
	Encode(w io.Writer, m *Message) error
}
