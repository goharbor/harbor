[![Build Status](https://img.shields.io/travis/golang-migrate/migrate/master.svg)](https://travis-ci.org/golang-migrate/migrate)
[![GoDoc](https://godoc.org/github.com/golang-migrate/migrate?status.svg)](https://godoc.org/github.com/golang-migrate/migrate)
[![Coverage Status](https://img.shields.io/coveralls/github/golang-migrate/migrate/master.svg)](https://coveralls.io/github/golang-migrate/migrate?branch=master)
[![packagecloud.io](https://img.shields.io/badge/deb-packagecloud.io-844fec.svg)](https://packagecloud.io/golang-migrate/migrate?filter=debs)
[![GitHub Release](https://img.shields.io/github/release/golang-migrate/migrate.svg)](https://github.com/golang-migrate/migrate/releases)


# migrate

__Database migrations written in Go. Use as [CLI](#cli-usage) or import as [library](#use-in-your-go-project).__

 * Migrate reads migrations from [sources](#migration-sources)
   and applies them in correct order to a [database](#databases).
 * Drivers are "dumb", migrate glues everything together and makes sure the logic is bulletproof.
   (Keeps the drivers lightweight, too.)
 * Database drivers don't assume things or try to correct user input. When in doubt, fail.


Looking for [v1](https://github.com/golang-migrate/migrate/tree/v1)?


## Databases

Database drivers run migrations. [Add a new database?](database/driver.go)

  * [PostgreSQL](database/postgres)
  * [Redshift](database/redshift)
  * [Ql](database/ql)
  * [Cassandra](database/cassandra)
  * [SQLite](database/sqlite3) ([todo #165](https://github.com/mattes/migrate/issues/165))
  * [MySQL/ MariaDB](database/mysql)
  * [Neo4j](database/neo4j) ([todo #167](https://github.com/mattes/migrate/issues/167))
  * [MongoDB](database/mongodb) ([todo #169](https://github.com/mattes/migrate/issues/169))
  * [CrateDB](database/crate) ([todo #170](https://github.com/mattes/migrate/issues/170))
  * [Shell](database/shell) ([todo #171](https://github.com/mattes/migrate/issues/171))
  * [Google Cloud Spanner](database/spanner)
  * [CockroachDB](database/cockroachdb)
  * [ClickHouse](database/clickhouse)


## Migration Sources

Source drivers read migrations from local or remote sources. [Add a new source?](source/driver.go)

  * [Filesystem](source/file) - read from fileystem
  * [Go-Bindata](source/go_bindata) - read from embedded binary data ([jteeuwen/go-bindata](https://github.com/jteeuwen/go-bindata))
  * [Github](source/github) - read from remote Github repositories
  * [AWS S3](source/aws_s3) - read from Amazon Web Services S3
  * [Google Cloud Storage](source/google_cloud_storage) - read from Google Cloud Platform Storage



## CLI usage

  * Simple wrapper around this library.
  * Handles ctrl+c (SIGINT) gracefully.
  * No config search paths, no config files, no magic ENV var injections.

__[CLI Documentation](cli)__

### Basic usage:

```
$ migrate -database postgres://localhost:5432/database up 2
```

### Docker usage

```
$ docker run -v {{ migration dir }}:/migrations --network host migrate/migrate
    -path=/migrations/ -database postgres://localhost:5432/database up 2
```

## Use in your Go project

 * API is stable and frozen for this release (v3.x).
 * Package migrate has no external dependencies.
 * Only import the drivers you need.
   (check [dependency_tree.txt](https://github.com/golang-migrate/migrate/releases) for each driver)
 * To help prevent database corruptions, it supports graceful stops via `GracefulStop chan bool`.
 * Bring your own logger.
 * Uses `io.Reader` streams internally for low memory overhead.
 * Thread-safe and no goroutine leaks.

__[Go Documentation](https://godoc.org/github.com/golang-migrate/migrate)__

```go
import (
    "github.com/golang-migrate/migrate"
    _ "github.com/golang-migrate/migrate/database/postgres"
    _ "github.com/golang-migrate/migrate/source/github"
)

func main() {
    m, err := migrate.New(
        "github://mattes:personal-access-token@mattes/migrate_test",
        "postgres://localhost:5432/database?sslmode=enable")
    m.Steps(2)
}
```

Want to use an existing database client?

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
    "github.com/golang-migrate/migrate"
    "github.com/golang-migrate/migrate/database/postgres"
    _ "github.com/golang-migrate/migrate/source/file"
)

func main() {
    db, err := sql.Open("postgres", "postgres://localhost:5432/database?sslmode=enable")
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    m, err := migrate.NewWithDatabaseInstance(
        "file:///migrations",
        "postgres", driver)
    m.Steps(2)
}
```

## Migration files

Each migration has an up and down migration. [Why?](FAQ.md#why-two-separate-files-up-and-down-for-a-migration)

```
1481574547_create_users_table.up.sql
1481574547_create_users_table.down.sql
```

[Best practices: How to write migrations.](MIGRATIONS.md)



## Development and Contributing

Yes, please! [`Makefile`](Makefile) is your friend,
read the [development guide](CONTRIBUTING.md).

Also have a look at the [FAQ](FAQ.md).



---

Looking for alternatives? [https://awesome-go.com/#database](https://awesome-go.com/#database).
