package sweeper

// Interface defines the operations a sweeper should have
type Interface interface {
	// Sweep the outdated log entries if necessary
	//
	// If failed, an non-nil error will return
	// If succeeded, count of swept log entries is returned
	Sweep() (int, error)

	// Return the sweeping duration with day unit.
	Duration() int
}
