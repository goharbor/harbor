package env

import (
	"context"
	"sync"
)

// Context keep some sharable materials and system controlling channels.
// The system context.Context interface is also included.
type Context struct {
	// The system context with cancel capability.
	SystemContext context.Context

	// Coordination signal
	WG *sync.WaitGroup

	// Report errors to bootstrap component
	// Once error is reported by lower components, the whole system should exit
	ErrorChan chan error

	// The base job context reference
	// It will be the parent conetext of job execution context
	JobContext JobContext
}
