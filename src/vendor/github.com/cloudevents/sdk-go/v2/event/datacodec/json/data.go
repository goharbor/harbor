/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package json

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

// Decode takes `in` as []byte.
// If Event sent the payload as base64, Decoder assumes that `in` is the
// decoded base64 byte array.
func Decode(ctx context.Context, in []byte, out interface{}) error {
	if in == nil {
		return nil
	}
	if out == nil {
		return fmt.Errorf("out is nil")
	}

	if err := json.Unmarshal(in, out); err != nil {
		return fmt.Errorf("[json] found bytes \"%s\", but failed to unmarshal: %s", string(in), err.Error())
	}
	return nil
}

// Encode attempts to json.Marshal `in` into bytes. Encode will inspect `in`
// and returns `in` unmodified if it is detected that `in` is already a []byte;
// Or json.Marshal errors.
func Encode(ctx context.Context, in interface{}) ([]byte, error) {
	if in == nil {
		return nil, nil
	}

	it := reflect.TypeOf(in)
	switch it.Kind() {
	case reflect.Slice:
		if it.Elem().Kind() == reflect.Uint8 {

			if b, ok := in.([]byte); ok && len(b) > 0 {
				// check to see if it is a pre-encoded byte string.
				if b[0] == byte('"') || b[0] == byte('{') || b[0] == byte('[') {
					return b, nil
				}
			}

		}
	}

	return json.Marshal(in)
}
