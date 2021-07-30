package context

import "context"

// Background returns a default context with logger discarded.
func Background() context.Context {
	ctx := context.Background()
	return WithLoggerDiscarded(ctx)
}
