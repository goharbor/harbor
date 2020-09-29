package chartserver

import (
	"fmt"
	"testing"
)

// Test controller
func TestController(t *testing.T) {
	s, c, err := createMockObjects()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	prefix := c.APIPrefix("fake")
	expected := fmt.Sprintf("%s/api/%s/charts", s.URL, "fake")
	if prefix != expected {
		t.Fatalf("expect '%s' but got '%s'", expected, prefix)
	}
}
