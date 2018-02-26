package testutils

// TestBootstrapper is a simple implemented of the Bootstrapper interface
// to be used for tests
type TestBootstrapper struct {
	Booted bool
}

// Bootstrap sets Booted to true so tests can confirm it was called
func (tb *TestBootstrapper) Bootstrap() error {
	tb.Booted = true
	return nil
}
