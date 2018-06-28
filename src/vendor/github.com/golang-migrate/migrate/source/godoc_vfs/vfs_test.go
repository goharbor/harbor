package godoc_vfs_test

import (
	"testing"

	"github.com/golang-migrate/migrate/source/godoc_vfs"
	st "github.com/golang-migrate/migrate/source/testing"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

func TestVFS(t *testing.T) {
	fs := mapfs.New(map[string]string{
		"1_foobar.up.sql":   "1 up",
		"1_foobar.down.sql": "1 down",
		"3_foobar.up.sql":   "3 up",
		"4_foobar.up.sql":   "4 up",
		"4_foobar.down.sql": "4 down",
		"5_foobar.down.sql": "5 down",
		"7_foobar.up.sql":   "7 up",
		"7_foobar.down.sql": "7 down",
	})

	d, err := godoc_vfs.WithInstance(fs, "")
	if err != nil {
		t.Fatal(err)
	}
	st.Test(t, d)
}

func TestOpen(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Open to panic")
		}
	}()
	b := &godoc_vfs.VFS{}
	b.Open("")
}
