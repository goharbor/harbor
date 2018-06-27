package mysql

import (
	"database/sql"
	sqldriver "database/sql/driver"
	"fmt"
	// "io/ioutil"
	// "log"
	"testing"

	"github.com/go-sql-driver/mysql"
	dt "github.com/golang-migrate/migrate/database/testing"
	mt "github.com/golang-migrate/migrate/testing"
)

var versions = []mt.Version{
	{Image: "mysql:8", ENV: []string{"MYSQL_ROOT_PASSWORD=root", "MYSQL_DATABASE=public"}},
	{Image: "mysql:5.7", ENV: []string{"MYSQL_ROOT_PASSWORD=root", "MYSQL_DATABASE=public"}},
	{Image: "mysql:5.6", ENV: []string{"MYSQL_ROOT_PASSWORD=root", "MYSQL_DATABASE=public"}},
	{Image: "mysql:5.5", ENV: []string{"MYSQL_ROOT_PASSWORD=root", "MYSQL_DATABASE=public"}},
}

func isReady(i mt.Instance) bool {
	db, err := sql.Open("mysql", fmt.Sprintf("root:root@tcp(%v:%v)/public", i.Host(), i.Port()))
	if err != nil {
		return false
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		switch err {
		case sqldriver.ErrBadConn, mysql.ErrInvalidConn:
			return false
		default:
			fmt.Println(err)
		}
		return false
	}

	return true
}

func Test(t *testing.T) {
	// mysql.SetLogger(mysql.Logger(log.New(ioutil.Discard, "", log.Ltime)))

	mt.ParallelTest(t, versions, isReady,
		func(t *testing.T, i mt.Instance) {
			p := &Mysql{}
			addr := fmt.Sprintf("mysql://root:root@tcp(%v:%v)/public", i.Host(), i.Port())
			d, err := p.Open(addr)
			if err != nil {
				t.Fatalf("%v", err)
			}
			defer d.Close()
			dt.Test(t, d, []byte("SELECT 1"))

			// check ensureVersionTable
			if err := d.(*Mysql).ensureVersionTable(); err != nil {
				t.Fatal(err)
			}
			// check again
			if err := d.(*Mysql).ensureVersionTable(); err != nil {
				t.Fatal(err)
			}
		})
}

func TestLockWorks(t *testing.T) {
	mt.ParallelTest(t, versions, isReady,
		func(t *testing.T, i mt.Instance) {
			p := &Mysql{}
			addr := fmt.Sprintf("mysql://root:root@tcp(%v:%v)/public", i.Host(), i.Port())
			d, err := p.Open(addr)
			if err != nil {
				t.Fatalf("%v", err)
			}
			dt.Test(t, d, []byte("SELECT 1"))

			ms := d.(*Mysql)

			err = ms.Lock()
			if err != nil {
				t.Fatal(err)
			}
			err = ms.Unlock()
			if err != nil {
				t.Fatal(err)
			}

			// make sure the 2nd lock works (RELEASE_LOCK is very finicky)
			err = ms.Lock()
			if err != nil {
				t.Fatal(err)
			}
			err = ms.Unlock()
			if err != nil {
				t.Fatal(err)
			}
		})
}
