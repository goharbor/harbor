package clickhouse

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"time"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
)

var DefaultMigrationsTable = "schema_migrations"

var ErrNilConfig = fmt.Errorf("no config")

type Config struct {
	DatabaseName    string
	MigrationsTable string
}

func init() {
	database.Register("clickhouse", &ClickHouse{})
}

func WithInstance(conn *sql.DB, config *Config) (database.Driver, error) {
	if config == nil {
		return nil, ErrNilConfig
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	ch := &ClickHouse{
		conn:   conn,
		config: config,
	}

	if err := ch.init(); err != nil {
		return nil, err
	}

	return ch, nil
}

type ClickHouse struct {
	conn   *sql.DB
	config *Config
}

func (ch *ClickHouse) Open(dsn string) (database.Driver, error) {
	purl, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	q := migrate.FilterCustomQuery(purl)
	q.Scheme = "tcp"
	conn, err := sql.Open("clickhouse", q.String())
	if err != nil {
		return nil, err
	}

	ch = &ClickHouse{
		conn: conn,
		config: &Config{
			MigrationsTable: purl.Query().Get("x-migrations-table"),
			DatabaseName:    purl.Query().Get("database"),
		},
	}

	if err := ch.init(); err != nil {
		return nil, err
	}

	return ch, nil
}

func (ch *ClickHouse) init() error {
	if len(ch.config.DatabaseName) == 0 {
		if err := ch.conn.QueryRow("SELECT currentDatabase()").Scan(&ch.config.DatabaseName); err != nil {
			return err
		}
	}

	if len(ch.config.MigrationsTable) == 0 {
		ch.config.MigrationsTable = DefaultMigrationsTable
	}

	return ch.ensureVersionTable()
}

func (ch *ClickHouse) Run(r io.Reader) error {
	migration, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	if _, err := ch.conn.Exec(string(migration)); err != nil {
		return database.Error{OrigErr: err, Err: "migration failed", Query: migration}
	}

	return nil
}
func (ch *ClickHouse) Version() (int, bool, error) {
	var (
		version int
		dirty   uint8
		query   = "SELECT version, dirty FROM `" + ch.config.MigrationsTable + "` ORDER BY sequence DESC LIMIT 1"
	)
	if err := ch.conn.QueryRow(query).Scan(&version, &dirty); err != nil {
		if err == sql.ErrNoRows {
			return database.NilVersion, false, nil
		}
		return 0, false, &database.Error{OrigErr: err, Query: []byte(query)}
	}
	return version, dirty == 1, nil
}

func (ch *ClickHouse) SetVersion(version int, dirty bool) error {
	var (
		bool = func(v bool) uint8 {
			if v {
				return 1
			}
			return 0
		}
		tx, err = ch.conn.Begin()
	)
	if err != nil {
		return err
	}

	query := "INSERT INTO " + ch.config.MigrationsTable + " (version, dirty, sequence) VALUES (?, ?, ?)"
	if _, err := tx.Exec(query, version, bool(dirty), time.Now().UnixNano()); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}

	return tx.Commit()
}

func (ch *ClickHouse) ensureVersionTable() error {
	var (
		table string
		query = "SHOW TABLES FROM " + ch.config.DatabaseName + " LIKE '" + ch.config.MigrationsTable + "'"
	)
	// check if migration table exists
	if err := ch.conn.QueryRow(query).Scan(&table); err != nil {
		if err != sql.ErrNoRows {
			return &database.Error{OrigErr: err, Query: []byte(query)}
		}
	} else {
		return nil
	}
	// if not, create the empty migration table
	query = `
		CREATE TABLE ` + ch.config.MigrationsTable + ` (
			version    UInt32, 
			dirty      UInt8,
			sequence   UInt64
		) Engine=TinyLog
	`
	if _, err := ch.conn.Exec(query); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}
	return nil
}

func (ch *ClickHouse) Drop() error {
	var (
		query       = "SHOW TABLES FROM " + ch.config.DatabaseName
		tables, err = ch.conn.Query(query)
	)
	if err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}
	defer tables.Close()
	for tables.Next() {
		var table string
		if err := tables.Scan(&table); err != nil {
			return err
		}

		query = "DROP TABLE IF EXISTS " + ch.config.DatabaseName + "." + table

		if _, err := ch.conn.Exec(query); err != nil {
			return &database.Error{OrigErr: err, Query: []byte(query)}
		}
	}
	return ch.ensureVersionTable()
}

func (ch *ClickHouse) Lock() error   { return nil }
func (ch *ClickHouse) Unlock() error { return nil }
func (ch *ClickHouse) Close() error  { return ch.conn.Close() }
