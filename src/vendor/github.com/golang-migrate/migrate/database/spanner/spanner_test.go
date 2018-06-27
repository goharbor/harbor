package spanner

import (
	"fmt"
	"os"
	"testing"

	dt "github.com/golang-migrate/migrate/database/testing"
)

func Test(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	db, ok := os.LookupEnv("SPANNER_DATABASE")
	if !ok {
		t.Skip("SPANNER_DATABASE not set, skipping test.")
	}

	s := &Spanner{}
	addr := fmt.Sprintf("spanner://%v", db)
	d, err := s.Open(addr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	dt.Test(t, d, []byte("SELECT 1"))
}
