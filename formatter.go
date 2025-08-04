/*
 *  Copyright (c) 2024-2025 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"bytes"
	"io"

	"go.osspkg.com/ioutils/pool"
)

var newLine = []byte("\n")

var poolBuffer = pool.New[*bytes.Buffer](func() *bytes.Buffer {
	return bytes.NewBuffer(make([]byte, 0, 1024))
})

type Formatter interface {
	Encode(w io.Writer, m *Message) error
}
