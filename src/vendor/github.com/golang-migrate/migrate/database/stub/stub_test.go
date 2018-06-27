package stub

import (
	"testing"

	dt "github.com/golang-migrate/migrate/database/testing"
)

func Test(t *testing.T) {
	s := &Stub{}
	d, err := s.Open("")
	if err != nil {
		t.Fatal(err)
	}
	dt.Test(t, d, []byte("/* foobar migration */"))
}
