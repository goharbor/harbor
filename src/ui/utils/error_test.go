package utils

import (
	"encoding/json"
	"errors"
	"testing"
)

// Test case for error wrapping function.
func TestWrapError(t *testing.T) {
	if WrapError(nil) != nil {
		t.Fatal("expect nil error but got a non-nil one")
	}

	err := errors.New("mock error")
	formatedErr := WrapError(err)
	if formatedErr == nil {
		t.Fatal("expect non-nil error but got nil")
	}

	jsonErr := formatedErr.Error()
	structuredErr := make(map[string]string, 1)
	if e := json.Unmarshal([]byte(jsonErr), &structuredErr); e != nil {
		t.Fatal("expect nil error but got a non-nil one when doing error converting")
	}
	if msg, ok := structuredErr["error"]; !ok {
		t.Fatal("expect an 'error' filed but missing")
	} else {
		if msg != "mock error" {
			t.Fatalf("expect error message '%s' but got '%s'", "mock error", msg)
		}
	}
}
