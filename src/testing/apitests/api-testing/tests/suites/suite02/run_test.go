package suite02

import (
	"testing"

	"github.com/goharbor/harbor/src/testing/apitests/api-testing/envs"
)

// TestRun : Start to run the case
func TestRun(t *testing.T) {
	// Initialize env
	if err := envs.ConcourseCILdapEnv.Load(); err != nil {
		t.Fatal(err.Error())
	}

	suite := ConcourseCiSuite02{}
	report := suite.Run(&envs.ConcourseCILdapEnv)
	report.Print()
	if report.IsFail() {
		t.Fail()
	}
}
