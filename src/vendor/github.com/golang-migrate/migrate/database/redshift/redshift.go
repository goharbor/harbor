package redshift

import (
	"net/url"

	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/postgres"
)

// init registers the driver under the name 'redshift'
func init() {
	db := new(Redshift)
	db.Driver = new(postgres.Postgres)

	database.Register("redshift", db)
}

// Redshift is a wrapper around the PostgreSQL driver which implements Redshift-specific behavior.
//
// Currently, the only different behaviour is the lack of locking in Redshift.  The (Un)Lock() method(s) have been overridden from the PostgreSQL adapter to simply return nil.
type Redshift struct {
	// The wrapped PostgreSQL driver.
	database.Driver
}

// Open implements the database.Driver interface by parsing the URL, switching the scheme from "redshift" to "postgres", and delegating to the underlying PostgreSQL driver.
func (driver *Redshift) Open(dsn string) (database.Driver, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	parsed.Scheme = "postgres"
	psql, err := driver.Driver.Open(parsed.String())
	if err != nil {
		return nil, err
	}

	return &Redshift{Driver: psql}, nil
}

// Lock implements the database.Driver interface by not locking and returning nil.
func (driver *Redshift) Lock() error { return nil }

// Unlock implements the database.Driver interface by not unlocking and returning nil.
func (driver *Redshift) Unlock() error { return nil }
