package logx

import (
	"encoding/json"
	"io"
	"testing"
	"time"
)

func BenchmarkJSONEncoderComparison(b *testing.B) {
	cases := []struct {
		name string
		msg  Message
	}{
		{
			name: "primitives",
			msg: Message{
				Time:    time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				Level:   "INFO",
				Message: "sync",
				Map:     map[string]string{"id": "123", "status": "ok"},
			},
		},
		{
			name: "struct",
			msg: Message{
				Time:    time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				Level:   "INFO",
				Message: "sync",
				Map:     map[string]string{"id": "123", "obj": `{"A":"x","B":2}`},
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name+"/current", func(b *testing.B) {
			formatter := NewFormatJSON()
			b.ReportAllocs()
			for b.Loop() {
				if err := formatter.Encode(io.Discard, &tc.msg); err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(tc.name+"/encoding_json", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				data, err := json.Marshal(&tc.msg)
				if err != nil {
					b.Fatal(err)
				}
				if _, err := io.Discard.Write(data); err != nil {
					b.Fatal(err)
				}
				if _, err := io.Discard.Write(newLine); err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(tc.name+"/append_bytes", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				var buffer [128]byte
				data, err := appendJSONMessage(buffer[:0], tc.msg)
				if err != nil {
					b.Fatal(err)
				}
				data = append(data, '\n')
				if _, err := io.Discard.Write(data); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
