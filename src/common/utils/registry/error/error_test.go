package error

import (
	"testing"
)

func TestError(t *testing.T) {
	err := &Error{
		StatusCode: 404,
		Detail:     "not found",
	}

	if err.Error() != "404 not found" {
		t.Fatalf("unexpected content: %s != %s",
			err.Error(), "404 not found")
	}
}
