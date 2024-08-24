// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package logx

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson4086215fDecodeGoOsspkgComLogx(in *jlexer.Lexer, out *Message) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "time":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.Time).UnmarshalJSON(data))
			}
		case "lvl":
			out.Level = string(in.String())
		case "msg":
			out.Message = string(in.String())
		case "ctx":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.Map = make(map[string]string)
				} else {
					out.Map = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v1 string
					v1 = string(in.String())
					(out.Map)[key] = v1
					in.WantComma()
				}
				in.Delim('}')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson4086215fEncodeGoOsspkgComLogx(out *jwriter.Writer, in Message) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"time\":"
		out.RawString(prefix[1:])
		out.Raw((in.Time).MarshalJSON())
	}
	{
		const prefix string = ",\"lvl\":"
		out.RawString(prefix)
		out.String(string(in.Level))
	}
	{
		const prefix string = ",\"msg\":"
		out.RawString(prefix)
		out.String(string(in.Message))
	}
	if len(in.Map) != 0 {
		const prefix string = ",\"ctx\":"
		out.RawString(prefix)
		{
			out.RawByte('{')
			v2First := true
			for v2Name, v2Value := range in.Map {
				if v2First {
					v2First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v2Name))
				out.RawByte(':')
				out.String(string(v2Value))
			}
			out.RawByte('}')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Message) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson4086215fEncodeGoOsspkgComLogx(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Message) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson4086215fEncodeGoOsspkgComLogx(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Message) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson4086215fDecodeGoOsspkgComLogx(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Message) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson4086215fDecodeGoOsspkgComLogx(l, v)
}
