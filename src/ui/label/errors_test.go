package label

import (
	"fmt"
	"testing"
)

// Test cases for kinds of error definitions.
func TestErrorFormats(t *testing.T) {
	br := NewErrLabelBadRequest("bad requests")
	if !checkErrorFormat(br, "bad requests") {
		t.Fatalf("expect an error with ErrLabelBadRequest kind but got incorrect format '%v'", br)
	}

	cf := NewErrLabelConflict(1, "c", "repo1/mychart:1.0.0")
	if !checkErrorFormat(cf, fmt.Sprintf("conflict: %s '%v' is already marked with label '%d'", "c", "repo1/mychart:1.0.0", 1)) {
		t.Fatalf("expect an error with ErrLabelConflict kind but got incorrect format '%v'", cf)
	}

	nf := NewErrLabelNotFound(1, "c", "repo1/mychart:1.0.0")
	if !checkErrorFormat(nf, fmt.Sprintf("not found: label '%d' on %s '%v'", 1, "c", "repo1/mychart:1.0.0")) {
		t.Fatalf("expect an error with ErrLabelNotFound kind but got incorrect format '%v'", nf)
	}

	nf2 := NewErrLabelNotFound(1, "", "")
	if !checkErrorFormat(nf2, fmt.Sprintf("not found: label '%d'", 1)) {
		t.Fatalf("expect an error with ErrLabelNotFound kind but got incorrect format %v", nf2)
	}
}

func checkErrorFormat(err error, msg string) bool {
	return err.Error() == msg
}
