package postgres

// error codes https://github.com/lib/pq/blob/master/error.go

import (
	"bytes"
	"database/sql"
	sqldriver "database/sql/driver"
	"fmt"
	"io"
	"testing"

	"context"
	dt "github.com/golang-migrate/migrate/database/testing"
	mt "github.com/golang-migrate/migrate/testing"
	// "github.com/lib/pq"
)

var versions = []mt.Version{
	{Image: "postgres:9.6"},
	{Image: "postgres:9.5"},
	{Image: "postgres:9.4"},
	{Image: "postgres:9.3"},
	{Image: "postgres:9.2"},
}

func isReady(i mt.Instance) bool {
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://postgres@%v:%v/postgres?sslmode=disable", i.Host(), i.Port()))
	if err != nil {
		return false
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		switch err {
		case sqldriver.ErrBadConn, io.EOF:
			return false
		default:
			fmt.Println(err)
		}
		return false
	}

	return true
}

func Test(t *testing.T) {
	mt.ParallelTest(t, versions, isReady,
		func(t *testing.T, i mt.Instance) {
			p := &Postgres{}
			addr := fmt.Sprintf("postgres://postgres@%v:%v/postgres?sslmode=disable", i.Host(), i.Port())
			d, err := p.Open(addr)
			if err != nil {
				t.Fatalf("%v", err)
			}
			defer d.Close()
			dt.Test(t, d, []byte("SELECT 1"))
		})
}

func TestMultiStatement(t *testing.T) {
	mt.ParallelTest(t, versions, isReady,
		func(t *testing.T, i mt.Instance) {
			p := &Postgres{}
			addr := fmt.Sprintf("postgres://postgres@%v:%v/postgres?sslmode=disable", i.Host(), i.Port())
			d, err := p.Open(addr)
			if err != nil {
				t.Fatalf("%v", err)
			}
			defer d.Close()
			if err := d.Run(bytes.NewReader([]byte("CREATE TABLE foo (foo text); CREATE TABLE bar (bar text);"))); err != nil {
				t.Fatalf("expected err to be nil, got %v", err)
			}

			// make sure second table exists
			var exists bool
			if err := d.(*Postgres).conn.QueryRowContext(context.Background(), "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'bar' AND table_schema = (SELECT current_schema()))").Scan(&exists); err != nil {
				t.Fatal(err)
			}
			if !exists {
				t.Fatalf("expected table bar to exist")
			}
		})
}

func TestFilterCustomQuery(t *testing.T) {
	mt.ParallelTest(t, versions, isReady,
		func(t *testing.T, i mt.Instance) {
			p := &Postgres{}
			addr := fmt.Sprintf("postgres://postgres@%v:%v/postgres?sslmode=disable&x-custom=foobar", i.Host(), i.Port())
			d, err := p.Open(addr)
			if err != nil {
				t.Fatalf("%v", err)
			}
			defer d.Close()
		})
}

func TestWithSchema(t *testing.T) {
	mt.ParallelTest(t, versions, isReady,
		func(t *testing.T, i mt.Instance) {
			p := &Postgres{}
			addr := fmt.Sprintf("postgres://postgres@%v:%v/postgres?sslmode=disable", i.Host(), i.Port())
			d, err := p.Open(addr)
			if err != nil {
				t.Fatalf("%v", err)
			}
			defer d.Close()

			// create foobar schema
			if err := d.Run(bytes.NewReader([]byte("CREATE SCHEMA foobar AUTHORIZATION postgres"))); err != nil {
				t.Fatal(err)
			}
			if err := d.SetVersion(1, false); err != nil {
				t.Fatal(err)
			}

			// re-connect using that schema
			d2, err := p.Open(fmt.Sprintf("postgres://postgres@%v:%v/postgres?sslmode=disable&search_path=foobar", i.Host(), i.Port()))
			if err != nil {
				t.Fatalf("%v", err)
			}
			defer d2.Close()

			version, _, err := d2.Version()
			if err != nil {
				t.Fatal(err)
			}
			if version != -1 {
				t.Fatal("expected NilVersion")
			}

			// now update version and compare
			if err := d2.SetVersion(2, false); err != nil {
				t.Fatal(err)
			}
			version, _, err = d2.Version()
			if err != nil {
				t.Fatal(err)
			}
			if version != 2 {
				t.Fatal("expected version 2")
			}

			// meanwhile, the public schema still has the other version
			version, _, err = d.Version()
			if err != nil {
				t.Fatal(err)
			}
			if version != 1 {
				t.Fatal("expected version 2")
			}
		})
}

func TestWithInstance(t *testing.T) {

}

func TestPostgres_Lock(t *testing.T) {
	mt.ParallelTest(t, versions, isReady,
		func(t *testing.T, i mt.Instance) {
			p := &Postgres{}
			addr := fmt.Sprintf("postgres://postgres@%v:%v/postgres?sslmode=disable", i.Host(), i.Port())
			d, err := p.Open(addr)
			if err != nil {
				t.Fatalf("%v", err)
			}

			dt.Test(t, d, []byte("SELECT 1"))

			ps := d.(*Postgres)

			err = ps.Lock()
			if err != nil {
				t.Fatal(err)
			}

			err = ps.Unlock()
			if err != nil {
				t.Fatal(err)
			}

			err = ps.Lock()
			if err != nil {
				t.Fatal(err)
			}

			err = ps.Unlock()
			if err != nil {
				t.Fatal(err)
			}
		})
}
