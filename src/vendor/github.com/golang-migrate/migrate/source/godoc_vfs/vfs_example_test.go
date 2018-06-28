package godoc_vfs_test

import (
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/source/godoc_vfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

func Example_mapfs() {
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
		panic("bad migrations found!")
	}
	m, err := migrate.NewWithSourceInstance("godoc-vfs", d, "database://foobar")
	if err != nil {
		panic("error creating the migrations")
	}
	m.Up()
}
