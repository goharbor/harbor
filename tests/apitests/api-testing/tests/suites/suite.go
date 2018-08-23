package suites

import (
	"github.com/goharbor/harbor/tests/apitests/api-testing/envs"
	"github.com/goharbor/harbor/tests/apitests/api-testing/lib"
)

//Suite : Run a group of test cases
type Suite interface {
	Run(onEnvironment envs.Environment) *lib.Report
}
