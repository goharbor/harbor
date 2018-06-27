// Package testing has the source tests.
// All source drivers must pass the Test function.
// This lives in it's own package so it stays a test dependency.
package testing

import (
	"os"
	"testing"

	"github.com/golang-migrate/migrate/source"
)

// Test runs tests against source implementations.
// It assumes that the driver tests has access to the following migrations:
//
// u = up migration, d = down migration, n = version
//  |  1  |  -  |  3  |  4  |  5  |  -  |  7  |
//  | u d |  -  | u   | u d |   d |  -  | u d |
//
// See source/stub/stub_test.go or source/file/file_test.go for an example.
func Test(t *testing.T, d source.Driver) {
	TestFirst(t, d)
	TestPrev(t, d)
	TestNext(t, d)
	TestReadUp(t, d)
	TestReadDown(t, d)
}

func TestFirst(t *testing.T, d source.Driver) {
	version, err := d.First()
	if err != nil {
		t.Fatalf("First: expected err to be nil, got %v", err)
	}
	if version != 1 {
		t.Errorf("First: expected 1, got %v", version)
	}
}

func TestPrev(t *testing.T, d source.Driver) {
	tt := []struct {
		version           uint
		expectErr         error
		expectPrevVersion uint
	}{
		{version: 0, expectErr: os.ErrNotExist},
		{version: 1, expectErr: os.ErrNotExist},
		{version: 2, expectErr: os.ErrNotExist},
		{version: 3, expectErr: nil, expectPrevVersion: 1},
		{version: 4, expectErr: nil, expectPrevVersion: 3},
		{version: 5, expectErr: nil, expectPrevVersion: 4},
		{version: 6, expectErr: os.ErrNotExist},
		{version: 7, expectErr: nil, expectPrevVersion: 5},
		{version: 8, expectErr: os.ErrNotExist},
		{version: 9, expectErr: os.ErrNotExist},
	}

	for i, v := range tt {
		pv, err := d.Prev(v.version)
		if (v.expectErr == os.ErrNotExist && !os.IsNotExist(err)) && v.expectErr != err {
			t.Errorf("Prev: expected %v, got %v, in %v", v.expectErr, err, i)
		}
		if err == nil && v.expectPrevVersion != pv {
			t.Errorf("Prev: expected %v, got %v, in %v", v.expectPrevVersion, pv, i)
		}
	}
}

func TestNext(t *testing.T, d source.Driver) {
	tt := []struct {
		version           uint
		expectErr         error
		expectNextVersion uint
	}{
		{version: 0, expectErr: os.ErrNotExist},
		{version: 1, expectErr: nil, expectNextVersion: 3},
		{version: 2, expectErr: os.ErrNotExist},
		{version: 3, expectErr: nil, expectNextVersion: 4},
		{version: 4, expectErr: nil, expectNextVersion: 5},
		{version: 5, expectErr: nil, expectNextVersion: 7},
		{version: 6, expectErr: os.ErrNotExist},
		{version: 7, expectErr: os.ErrNotExist},
		{version: 8, expectErr: os.ErrNotExist},
		{version: 9, expectErr: os.ErrNotExist},
	}

	for i, v := range tt {
		nv, err := d.Next(v.version)
		if (v.expectErr == os.ErrNotExist && !os.IsNotExist(err)) && v.expectErr != err {
			t.Errorf("Next: expected %v, got %v, in %v", v.expectErr, err, i)
		}
		if err == nil && v.expectNextVersion != nv {
			t.Errorf("Next: expected %v, got %v, in %v", v.expectNextVersion, nv, i)
		}
	}
}

func TestReadUp(t *testing.T, d source.Driver) {
	tt := []struct {
		version   uint
		expectErr error
		expectUp  bool
	}{
		{version: 0, expectErr: os.ErrNotExist},
		{version: 1, expectErr: nil, expectUp: true},
		{version: 2, expectErr: os.ErrNotExist},
		{version: 3, expectErr: nil, expectUp: true},
		{version: 4, expectErr: nil, expectUp: true},
		{version: 5, expectErr: os.ErrNotExist},
		{version: 6, expectErr: os.ErrNotExist},
		{version: 7, expectErr: nil, expectUp: true},
		{version: 8, expectErr: os.ErrNotExist},
	}

	for i, v := range tt {
		up, identifier, err := d.ReadUp(v.version)
		if (v.expectErr == os.ErrNotExist && !os.IsNotExist(err)) ||
			(v.expectErr != os.ErrNotExist && err != v.expectErr) {
			t.Errorf("expected %v, got %v, in %v", v.expectErr, err, i)

		} else if err == nil {
			if len(identifier) == 0 {
				t.Errorf("expected identifier not to be empty, in %v", i)
			}

			if v.expectUp == true && up == nil {
				t.Errorf("expected up not to be nil, in %v", i)
			} else if v.expectUp == false && up != nil {
				t.Errorf("expected up to be nil, got %v, in %v", up, i)
			}
		}
	}
}

func TestReadDown(t *testing.T, d source.Driver) {
	tt := []struct {
		version    uint
		expectErr  error
		expectDown bool
	}{
		{version: 0, expectErr: os.ErrNotExist},
		{version: 1, expectErr: nil, expectDown: true},
		{version: 2, expectErr: os.ErrNotExist},
		{version: 3, expectErr: os.ErrNotExist},
		{version: 4, expectErr: nil, expectDown: true},
		{version: 5, expectErr: nil, expectDown: true},
		{version: 6, expectErr: os.ErrNotExist},
		{version: 7, expectErr: nil, expectDown: true},
		{version: 8, expectErr: os.ErrNotExist},
	}

	for i, v := range tt {
		down, identifier, err := d.ReadDown(v.version)
		if (v.expectErr == os.ErrNotExist && !os.IsNotExist(err)) ||
			(v.expectErr != os.ErrNotExist && err != v.expectErr) {
			t.Errorf("expected %v, got %v, in %v", v.expectErr, err, i)

		} else if err == nil {
			if len(identifier) == 0 {
				t.Errorf("expected identifier not to be empty, in %v", i)
			}

			if v.expectDown == true && down == nil {
				t.Errorf("expected down not to be nil, in %v", i)
			} else if v.expectDown == false && down != nil {
				t.Errorf("expected down to be nil, got %v, in %v", down, i)
			}
		}
	}
}
