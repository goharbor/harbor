package testing

import (
	"testing"
)

func ExampleParallelTest(t *testing.T) {
	var isReady = func(i Instance) bool {
		// Return true if Instance is ready to run tests.
		// Don't block here though.
		return true
	}

	// t is *testing.T coming from parent Test(t *testing.T)
	ParallelTest(t, []Version{{Image: "docker_image:9.6"}}, isReady,
		func(t *testing.T, i Instance) {
			// Run your test/s ...
			t.Fatal("...")
		})
}
