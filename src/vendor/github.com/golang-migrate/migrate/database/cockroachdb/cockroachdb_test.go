package cockroachdb

// error codes https://github.com/lib/pq/blob/master/error.go

import (
	//"bytes"
	"database/sql"
	"fmt"
	"io"
	"testing"

	"github.com/lib/pq"
	dt "github.com/golang-migrate/migrate/database/testing"
	mt "github.com/golang-migrate/migrate/testing"
	"bytes"
)

var versions = []mt.Version{
	{Image: "cockroachdb/cockroach:v1.0.2", Cmd: []string{"start", "--insecure"}},
}

func isReady(i mt.Instance) bool {
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://root@%v:%v?sslmode=disable", i.Host(), i.PortFor(26257)))
	if err != nil {
		return false
	}
	defer db.Close()
	err = db.Ping()
	if err == io.EOF {
		_, err = db.Exec("CREATE DATABASE migrate")
		return err == nil;
	} else if e, ok := err.(*pq.Error); ok {
		if e.Code.Name() == "cannot_connect_now" {
			return false
		}
	}

	_, err = db.Exec("CREATE DATABASE migrate")
	return err == nil;

	return true
}

func Test(t *testing.T) {
	mt.ParallelTest(t, versions, isReady,
		func(t *testing.T, i mt.Instance) {
			c := &CockroachDb{}
			addr := fmt.Sprintf("cockroach://root@%v:%v/migrate?sslmode=disable", i.Host(), i.PortFor(26257))
			d, err := c.Open(addr)
			if err != nil {
				t.Fatalf("%v", err)
			}
			dt.Test(t, d, []byte("SELECT 1"))
		})
}

func TestMultiStatement(t *testing.T) {
	mt.ParallelTest(t, versions, isReady,
		func(t *testing.T, i mt.Instance) {
			c := &CockroachDb{}
			addr := fmt.Sprintf("cockroach://root@%v:%v/migrate?sslmode=disable", i.Host(), i.Port())
			d, err := c.Open(addr)
			if err != nil {
				t.Fatalf("%v", err)
			}
			if err := d.Run(bytes.NewReader([]byte("CREATE TABLE foo (foo text); CREATE TABLE bar (bar text);"))); err != nil {
				t.Fatalf("expected err to be nil, got %v", err)
			}

			// make sure second table exists
			var exists bool
			if err := d.(*CockroachDb).db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'bar' AND table_schema = (SELECT current_schema()))").Scan(&exists); err != nil {
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
			c := &CockroachDb{}
			addr := fmt.Sprintf("cockroach://root@%v:%v/migrate?sslmode=disable&x-custom=foobar", i.Host(), i.PortFor(26257))
			_, err := c.Open(addr)
			if err != nil {
				t.Fatalf("%v", err)
			}
		})
}
