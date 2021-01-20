package exporter

import "github.com/goharbor/harbor/src/common/utils/test"

var isDBInitialized = false

// initDBOnce initializes DB only once during all the tests.
func initDBOnce() {
	if !isDBInitialized {
		test.InitDatabaseFromEnv()

		isDBInitialized = true
	}
}
