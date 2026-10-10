/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package logx

import (
	"fmt"
	"reflect"
	"strconv"
)

// appendSimpleStruct handles the common flat struct shape without fmt's
// recursive printer. Other shapes keep fmt's exact %#v behavior.
func appendSimpleStruct(dst []byte, value any) ([]byte, bool) {
	if _, ok := value.(fmt.Formatter); ok {
		return dst, false
	}

	v := reflect.ValueOf(value)
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return dst, false
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() || field.Type.PkgPath() != "" {
			return dst, false
		}
		switch field.Type.Kind() {
		case reflect.String, reflect.Int:
		default:
			return dst, false
		}
	}

	dst = append(dst, t.String()...)
	dst = append(dst, '{')
	for i := 0; i < t.NumField(); i++ {
		if i != 0 {
			dst = append(dst, ',', ' ')
		}
		field := t.Field(i)
		dst = append(dst, field.Name...)
		dst = append(dst, ':')
		fieldValue := v.Field(i)
		if field.Type.Kind() == reflect.String {
			dst = strconv.AppendQuote(dst, fieldValue.String())
		} else {
			dst = strconv.AppendInt(dst, fieldValue.Int(), 10)
		}
	}
	return append(dst, '}'), true
}
