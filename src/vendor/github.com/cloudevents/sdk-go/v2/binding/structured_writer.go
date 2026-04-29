/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package binding

import (
	"context"
	"io"

	"github.com/cloudevents/sdk-go/v2/binding/format"
)

// StructuredWriter is used to visit a structured Message and generate a new representation.
//
// Protocols that supports structured encoding should implement this interface to implement direct
// structured to structured encoding and event to structured encoding.
type StructuredWriter interface {
	// Event receives an io.Reader for the whole event.
	SetStructuredEvent(ctx context.Context, format format.Format, event io.Reader) error
}
