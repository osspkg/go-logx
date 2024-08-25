/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx_test

import (
	"bytes"
	"testing"
	"time"

	"go.osspkg.com/logx"
)

func TestUnit_FormatString_Encode(t *testing.T) {
	tests := []struct {
		name    string
		args    *logx.Message
		want    []byte
		wantErr bool
	}{
		{
			name: "Case1",
			args: &logx.Message{
				Time:    time.Now(),
				Level:   "INF",
				Message: "Hello",
				Ctx: []interface{}{
					"err", "err\nmsg",
				},
			},
			want:    []byte("\"level\"=\"INF\"\t\"msg\"=\"Hello\"\t\"err\"=\"err\\nmsg\""),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w bytes.Buffer
			fo := logx.NewFormatString()
			err := fo.Encode(&w, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := w.Bytes()
			if !bytes.Contains(got, tt.want) {
				t.Errorf("Encode() got = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
