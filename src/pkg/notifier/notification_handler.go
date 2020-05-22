package notifier

// NotificationHandler defines what operations a notification handler
// should have.
type NotificationHandler interface {
	// Handle the event when it coming.
	// value might be optional, it depends on usages.
	Handle(value interface{}) error

	// IsStateful returns whether the handler is stateful or not.
	// If handler is stateful, it will not be triggered in parallel.
	// Otherwise, the handler will be triggered concurrently if more
	// than one same handler are matched the topics.
	IsStateful() bool
}
