// Copyright Project Harbor Authors. All rights reserved.

package logger

// Closer defines method to close the open io stream used by logger.
type Closer interface {
	// Close the opened io stream
	Close() error
}
