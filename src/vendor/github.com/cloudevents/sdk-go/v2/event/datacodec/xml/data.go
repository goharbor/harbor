/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package xml

import (
	"context"
	"encoding/xml"
	"fmt"
)

// Decode takes `in` as []byte.
// If Event sent the payload as base64, Decoder assumes that `in` is the
// decoded base64 byte array.
func Decode(ctx context.Context, in []byte, out interface{}) error {
	if in == nil {
		return nil
	}

	if err := xml.Unmarshal(in, out); err != nil {
		return fmt.Errorf("[xml] found bytes, but failed to unmarshal: %s %s", err.Error(), string(in))
	}
	return nil
}

// Encode attempts to xml.Marshal `in` into bytes. Encode will inspect `in`
// and returns `in` unmodified if it is detected that `in` is already a []byte;
// Or xml.Marshal errors.
func Encode(ctx context.Context, in interface{}) ([]byte, error) {
	if b, ok := in.([]byte); ok {
		// check to see if it is a pre-encoded byte string.
		if len(b) > 0 && b[0] == byte('"') {
			return b, nil
		}
	}

	return xml.Marshal(in)
}
