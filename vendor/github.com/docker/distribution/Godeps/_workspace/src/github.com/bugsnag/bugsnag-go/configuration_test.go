package bugsnag

import (
	"testing"
)

func TestNotifyReleaseStages(t *testing.T) {

	var testCases = []struct {
		stage      string
		configured []string
		notify     bool
		msg        string
	}{
		{
			stage:  "production",
			notify: true,
			msg:    "Should notify in all release stages by default",
		},
		{
			stage:      "production",
			configured: []string{"development", "production"},
			notify:     true,
			msg:        "Failed to notify in configured release stage",
		},
		{
			stage:      "staging",
			configured: []string{"development", "production"},
			notify:     false,
			msg:        "Failed to prevent notification in excluded release stage",
		},
	}

	for _, testCase := range testCases {
		Configure(Configuration{ReleaseStage: testCase.stage, NotifyReleaseStages: testCase.configured})

		if Config.notifyInReleaseStage() != testCase.notify {
			t.Error(testCase.msg)
		}
	}
}

func TestProjectPackages(t *testing.T) {
	Configure(Configuration{ProjectPackages: []string{"main", "github.com/ConradIrwin/*"}})
	if !Config.isProjectPackage("main") {
		t.Error("literal project package doesn't work")
	}
	if !Config.isProjectPackage("github.com/ConradIrwin/foo") {
		t.Error("wildcard project package doesn't work")
	}
	if Config.isProjectPackage("runtime") {
		t.Error("wrong packges being marked in project")
	}
	if Config.isProjectPackage("github.com/ConradIrwin/foo/bar") {
		t.Error("wrong packges being marked in project")
	}

}
