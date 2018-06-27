package ql

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/cznic/ql/driver"
	"github.com/golang-migrate/migrate"
	dt "github.com/golang-migrate/migrate/database/testing"
	_ "github.com/golang-migrate/migrate/source/file"
)

func Test(t *testing.T) {
	dir, err := ioutil.TempDir("", "ql-driver-test")
	if err != nil {
		return
	}
	defer func() {
		os.RemoveAll(dir)
	}()
	fmt.Printf("DB path : %s\n", filepath.Join(dir, "ql.db"))
	p := &Ql{}
	addr := fmt.Sprintf("ql://%s", filepath.Join(dir, "ql.db"))
	d, err := p.Open(addr)
	if err != nil {
		t.Fatalf("%v", err)
	}

	db, err := sql.Open("ql", filepath.Join(dir, "ql.db"))
	if err != nil {
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			return
		}
	}()
	dt.Test(t, d, []byte("CREATE TABLE t (Qty int, Name string);"))
	driver, err := WithInstance(db, &Config{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if err := d.Drop(); err != nil {
		t.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migration",
		"ql", driver)
	if err != nil {
		t.Fatalf("%v", err)
	}
	fmt.Println("UP")
	err = m.Up()
	if err != nil {
		t.Fatalf("%v", err)
	}
}
