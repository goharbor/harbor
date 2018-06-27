package bindata

import (
	"testing"

	"github.com/golang-migrate/migrate/source/go_bindata/testdata"
	st "github.com/golang-migrate/migrate/source/testing"
)

func Test(t *testing.T) {
	// wrap assets into Resource first
	s := Resource(testdata.AssetNames(),
		func(name string) ([]byte, error) {
			return testdata.Asset(name)
		})

	d, err := WithInstance(s)
	if err != nil {
		t.Fatal(err)
	}
	st.Test(t, d)
}

func TestWithInstance(t *testing.T) {
	// wrap assets into Resource
	s := Resource(testdata.AssetNames(),
		func(name string) ([]byte, error) {
			return testdata.Asset(name)
		})

	_, err := WithInstance(s)
	if err != nil {
		t.Fatal(err)
	}
}

func TestOpen(t *testing.T) {
	b := &Bindata{}
	_, err := b.Open("")
	if err == nil {
		t.Fatal("expected err, because it's not implemented yet")
	}
}
