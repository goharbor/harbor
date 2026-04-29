/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package protocol

import (
	"context"
)

// Opener is the common interface for things that need to be opened.
type Opener interface {
	// OpenInbound is a blocking call and ctx is used to stop the Inbound message Receiver/Responder.
	// Closing the context won't close the Receiver/Responder, aka it won't invoke Close(ctx).
	OpenInbound(ctx context.Context) error
}

// Closer is the common interface for things that can be closed.
// After invoking Close(ctx), you cannot reuse the object you closed.
type Closer interface {
	Close(ctx context.Context) error
}
