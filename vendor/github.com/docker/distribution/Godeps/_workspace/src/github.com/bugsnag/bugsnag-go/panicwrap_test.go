// +build !appengine

package bugsnag

import (
	"github.com/bitly/go-simplejson"
	"github.com/mitchellh/osext"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestPanicHandler(t *testing.T) {
	startTestServer()

	exePath, err := osext.Executable()
	if err != nil {
		t.Fatal(err)
	}

	// Use the same trick as panicwrap() to re-run ourselves.
	// In the init() block below, we will then panic.
	cmd := exec.Command(exePath, os.Args[1:]...)
	cmd.Env = append(os.Environ(), "BUGSNAG_API_KEY="+testAPIKey, "BUGSNAG_ENDPOINT="+testEndpoint, "please_panic=please_panic")

	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}

	if err = cmd.Wait(); err.Error() != "exit status 2" {
		t.Fatal(err)
	}

	json, err := simplejson.NewJson(<-postedJSON)
	if err != nil {
		t.Fatal(err)
	}

	event := json.Get("events").GetIndex(0)

	if event.Get("severity").MustString() != "error" {
		t.Errorf("severity should be error")
	}
	exception := event.Get("exceptions").GetIndex(0)

	if exception.Get("message").MustString() != "ruh roh" {
		t.Errorf("caught wrong panic")
	}

	if exception.Get("errorClass").MustString() != "panic" {
		t.Errorf("caught wrong panic")
	}

	frame := exception.Get("stacktrace").GetIndex(1)

	// Yeah, we just caught a panic from the init() function below and sent it to the server running above (mindblown)
	if frame.Get("inProject").MustBool() != true ||
		frame.Get("file").MustString() != "panicwrap_test.go" ||
		frame.Get("method").MustString() != "panick" ||
		frame.Get("lineNumber").MustInt() == 0 {
		t.Errorf("stack trace seemed wrong")
	}
}

func init() {
	if os.Getenv("please_panic") != "" {
		Configure(Configuration{APIKey: os.Getenv("BUGSNAG_API_KEY"), Endpoint: os.Getenv("BUGSNAG_ENDPOINT"), ProjectPackages: []string{"github.com/bugsnag/bugsnag-go"}})
		go func() {
			panick()
		}()
		// Plenty of time to crash, it shouldn't need any of it.
		time.Sleep(1 * time.Second)
	}
}

func panick() {
	panic("ruh roh")
}
