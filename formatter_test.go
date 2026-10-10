/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"testing"
	"time"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/logx"
)

type testTextMarshaler []byte

func (value testTextMarshaler) MarshalText() ([]byte, error) {
	return value, nil
}

type testBinaryMarshaler []byte

func (value testBinaryMarshaler) MarshalBinary() ([]byte, error) {
	return value, nil
}

func TestFormatJSONEscapedContextValue(t *testing.T) {
	cases := []struct {
		name  string
		value string
	}{
		{name: "quote", value: `quoted "text"`},
		{name: "backslash", value: `path\to\file`},
		{name: "newline", value: "first\nsecond"},
		{name: "tab", value: "first\tsecond"},
		{name: "unicode", value: "snowman ☃"},
		{name: "backtick", value: "`quoted`"},
		{name: "invalid utf8", value: string([]byte{'a', 0xff, 'b'})},
	}

	for _, tc := range cases {
		values := []struct {
			name  string
			value any
		}{
			{name: "string", value: tc.value},
			{name: "bytes", value: []byte(tc.value)},
			{name: "text marshaler", value: testTextMarshaler([]byte(tc.value))},
			{name: "binary marshaler", value: testBinaryMarshaler([]byte(tc.value))},
		}
		for _, item := range values {
			t.Run(tc.name+"/"+item.name, func(t *testing.T) {
				var output bytes.Buffer
				message := &logx.Message{Ctx: []any{"payload", item.value}}
				if err := logx.NewFormatJSON().Encode(&output, message); err != nil {
					t.Fatal(err)
				}

				var decoded struct {
					Ctx map[string]string `json:"ctx"`
				}
				if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
					t.Fatal(err)
				}
				want := strings.Trim(fmt.Sprintf("%#v", tc.value), "\"")
				if got := decoded.Ctx["payload"]; got != want {
					t.Errorf("context value = %q, want %q", got, want)
				}
			})
		}
	}
}

func TestFormatJSONFallbackContextValue(t *testing.T) {
	value := struct {
		Text string
		ID   int
	}{Text: "quoted \"text\"\n", ID: 7}
	cases := []struct {
		name  string
		value any
		want  string
	}{
		{name: "struct", value: value, want: strings.Trim(fmt.Sprintf("%#v", value), "\"")},
		{name: "nil", value: nil, want: "null"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := logx.NewFormatJSON().Encode(&output, &logx.Message{Ctx: []any{"payload", tc.value}}); err != nil {
				t.Fatal(err)
			}
			var decoded struct {
				Ctx map[string]string `json:"ctx"`
			}
			if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
				t.Fatal(err)
			}
			if got := decoded.Ctx["payload"]; got != tc.want {
				t.Errorf("context value = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatJSONContextPairs(t *testing.T) {
	cases := []struct {
		name    string
		message logx.Message
		want    map[string]string
	}{
		{
			name:    "three unique keys",
			message: logx.Message{Ctx: []any{"a", 1, "b", 2, "c", 3}},
			want:    map[string]string{"a": "1", "b": "2", "c": "3"},
		},
		{
			name:    "eight unique keys",
			message: logx.Message{Ctx: []any{"a", 1, "b", 2, "c", 3, "d", 4, "e", 5, "f", 6, "g", 7, "h", 8}},
			want:    map[string]string{"a": "1", "b": "2", "c": "3", "d": "4", "e": "5", "f": "6", "g": "7", "h": "8"},
		},
		{
			name:    "duplicate key keeps last value",
			message: logx.Message{Ctx: []any{"a", 1, "b", 2, "a", 3}},
			want:    map[string]string{"a": "3", "b": "2"},
		},
		{
			name:    "context overrides map",
			message: logx.Message{Map: map[string]string{"a": "old", "b": "old"}, Ctx: []any{"a", "new", "c", 3, "d", 4}},
			want:    map[string]string{"a": "new", "b": "old", "c": "3", "d": "4"},
		},
		{
			name:    "non string key",
			message: logx.Message{Ctx: []any{"a", 1, 7, 2, "c", 3}},
			want:    map[string]string{"a": "1", "7": "2", "c": "3"},
		},
		{
			name:    "escaped key",
			message: logx.Message{Ctx: []any{"a", 1, "b\n", 2, "c", 3}},
			want:    map[string]string{"a": "1", strings.Trim(fmt.Sprintf("%#v", "b\n"), "\""): "2", "c": "3"},
		},
		{
			name:    "odd context length",
			message: logx.Message{Ctx: []any{"a", 1, "b", 2, "c"}},
			want:    map[string]string{"a": "1", "b": "2", "c": "null"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := logx.NewFormatJSON().Encode(&output, &tc.message); err != nil {
				t.Fatal(err)
			}
			var decoded struct {
				Ctx map[string]string `json:"ctx"`
			}
			if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
				t.Fatal(err)
			}
			if !maps.Equal(decoded.Ctx, tc.want) {
				t.Errorf("context = %#v, want %#v", decoded.Ctx, tc.want)
			}
		})
	}
}

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
				Ctx: []any{
					"err", "err\nmsg",
				},
			},
			want:    []byte("\"level\"=\"INF\"\t\"msg\"=\"Hello\"\t\"err\"=\"err\\nmsg\""),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := bytes.NewBuffer(nil)
			fo := logx.NewFormatString()
			err := fo.Encode(w, tt.args)
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

func TestUnit_debug(t *testing.T) {
	t.SkipNow()

	var w bytes.Buffer
	fj := logx.NewFormatJSON()
	fj.Encode(&w, &logx.Message{Ctx: []any{"a\"\na", "a\nb\n"}, Map: make(map[string]string)})
	fj.Encode(&w, &logx.Message{Message: "a\"\nb\n"})
	fj.Encode(&w, &logx.Message{})

	fs := logx.NewFormatString()
	fs.Encode(&w, &logx.Message{Ctx: []any{"a\"\na", "a\nb\n"}})
	fs.Encode(&w, &logx.Message{Message: "a\"\nb\n"})
	fs.Encode(&w, &logx.Message{})

	result := string(w.Bytes())
	wait := `{"time":"0001-01-01T00:00:00Z","level":"","msg":""}
{"time":"0001-01-01T00:00:00Z","level":"","msg":""}
{"time":"0001-01-01T00:00:00Z","level":"","msg":""}
"time"="0001-01-01T00:00:00Z"	"level"=""	"msg"=""	"a\na"="a\nb\n"	
"time"="0001-01-01T00:00:00Z"	"level"=""	"msg"="a\nb\n"	
`

	casecheck.Equal(t, wait, result)
}
