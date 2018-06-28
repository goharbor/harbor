// +build go1.9

package mysql

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	nurl "net/url"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
)

func init() {
	database.Register("mysql", &Mysql{})
}

var DefaultMigrationsTable = "schema_migrations"

var (
	ErrDatabaseDirty  = fmt.Errorf("database is dirty")
	ErrNilConfig      = fmt.Errorf("no config")
	ErrNoDatabaseName = fmt.Errorf("no database name")
	ErrAppendPEM      = fmt.Errorf("failed to append PEM")
)

type Config struct {
	MigrationsTable string
	DatabaseName    string
}

type Mysql struct {
	// mysql RELEASE_LOCK must be called from the same conn, so
	// just do everything over a single conn anyway.
	conn     *sql.Conn
	isLocked bool

	config *Config
}

// instance must have `multiStatements` set to true
func WithInstance(instance *sql.DB, config *Config) (database.Driver, error) {
	if config == nil {
		return nil, ErrNilConfig
	}

	if err := instance.Ping(); err != nil {
		return nil, err
	}

	query := `SELECT DATABASE()`
	var databaseName sql.NullString
	if err := instance.QueryRow(query).Scan(&databaseName); err != nil {
		return nil, &database.Error{OrigErr: err, Query: []byte(query)}
	}

	if len(databaseName.String) == 0 {
		return nil, ErrNoDatabaseName
	}

	config.DatabaseName = databaseName.String

	if len(config.MigrationsTable) == 0 {
		config.MigrationsTable = DefaultMigrationsTable
	}

	conn, err := instance.Conn(context.Background())
	if err != nil {
		return nil, err
	}

	mx := &Mysql{
		conn:   conn,
		config: config,
	}

	if err := mx.ensureVersionTable(); err != nil {
		return nil, err
	}

	return mx, nil
}

func (m *Mysql) Open(url string) (database.Driver, error) {
	purl, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	q := purl.Query()
	q.Set("multiStatements", "true")
	purl.RawQuery = q.Encode()

	db, err := sql.Open("mysql", strings.Replace(
		migrate.FilterCustomQuery(purl).String(), "mysql://", "", 1))
	if err != nil {
		return nil, err
	}

	migrationsTable := purl.Query().Get("x-migrations-table")
	if len(migrationsTable) == 0 {
		migrationsTable = DefaultMigrationsTable
	}

	// use custom TLS?
	ctls := purl.Query().Get("tls")
	if len(ctls) > 0 {
		if _, isBool := readBool(ctls); !isBool && strings.ToLower(ctls) != "skip-verify" {
			rootCertPool := x509.NewCertPool()
			pem, err := ioutil.ReadFile(purl.Query().Get("x-tls-ca"))
			if err != nil {
				return nil, err
			}

			if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
				return nil, ErrAppendPEM
			}

			certs, err := tls.LoadX509KeyPair(purl.Query().Get("x-tls-cert"), purl.Query().Get("x-tls-key"))
			if err != nil {
				return nil, err
			}

			insecureSkipVerify := false
			if len(purl.Query().Get("x-tls-insecure-skip-verify")) > 0 {
				x, err := strconv.ParseBool(purl.Query().Get("x-tls-insecure-skip-verify"))
				if err != nil {
					return nil, err
				}
				insecureSkipVerify = x
			}

			mysql.RegisterTLSConfig(ctls, &tls.Config{
				RootCAs:            rootCertPool,
				Certificates:       []tls.Certificate{certs},
				InsecureSkipVerify: insecureSkipVerify,
			})
		}
	}

	mx, err := WithInstance(db, &Config{
		DatabaseName:    purl.Path,
		MigrationsTable: migrationsTable,
	})
	if err != nil {
		return nil, err
	}

	return mx, nil
}

func (m *Mysql) Close() error {
	return m.conn.Close()
}

func (m *Mysql) Lock() error {
	if m.isLocked {
		return database.ErrLocked
	}

	aid, err := database.GenerateAdvisoryLockId(
		fmt.Sprintf("%s:%s", m.config.DatabaseName, m.config.MigrationsTable))
	if err != nil {
		return err
	}

	query := "SELECT GET_LOCK(?, 10)"
	var success bool
	if err := m.conn.QueryRowContext(context.Background(), query, aid).Scan(&success); err != nil {
		return &database.Error{OrigErr: err, Err: "try lock failed", Query: []byte(query)}
	}

	if success {
		m.isLocked = true
		return nil
	}

	return database.ErrLocked
}

func (m *Mysql) Unlock() error {
	if !m.isLocked {
		return nil
	}

	aid, err := database.GenerateAdvisoryLockId(
		fmt.Sprintf("%s:%s", m.config.DatabaseName, m.config.MigrationsTable))
	if err != nil {
		return err
	}

	query := `SELECT RELEASE_LOCK(?)`
	if _, err := m.conn.ExecContext(context.Background(), query, aid); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}

	// NOTE: RELEASE_LOCK could return NULL or (or 0 if the code is changed),
	// in which case isLocked should be true until the timeout expires -- synchronizing
	// these states is likely not worth trying to do; reconsider the necessity of isLocked.

	m.isLocked = false
	return nil
}

func (m *Mysql) Run(migration io.Reader) error {
	migr, err := ioutil.ReadAll(migration)
	if err != nil {
		return err
	}

	query := string(migr[:])
	if _, err := m.conn.ExecContext(context.Background(), query); err != nil {
		return database.Error{OrigErr: err, Err: "migration failed", Query: migr}
	}

	return nil
}

func (m *Mysql) SetVersion(version int, dirty bool) error {
	tx, err := m.conn.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return &database.Error{OrigErr: err, Err: "transaction start failed"}
	}

	query := "TRUNCATE `" + m.config.MigrationsTable + "`"
	if _, err := tx.ExecContext(context.Background(), query); err != nil {
		tx.Rollback()
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}

	if version >= 0 {
		query := "INSERT INTO `" + m.config.MigrationsTable + "` (version, dirty) VALUES (?, ?)"
		if _, err := tx.ExecContext(context.Background(), query, version, dirty); err != nil {
			tx.Rollback()
			return &database.Error{OrigErr: err, Query: []byte(query)}
		}
	}

	if err := tx.Commit(); err != nil {
		return &database.Error{OrigErr: err, Err: "transaction commit failed"}
	}

	return nil
}

func (m *Mysql) Version() (version int, dirty bool, err error) {
	query := "SELECT version, dirty FROM `" + m.config.MigrationsTable + "` LIMIT 1"
	err = m.conn.QueryRowContext(context.Background(), query).Scan(&version, &dirty)
	switch {
	case err == sql.ErrNoRows:
		return database.NilVersion, false, nil

	case err != nil:
		if e, ok := err.(*mysql.MySQLError); ok {
			if e.Number == 0 {
				return database.NilVersion, false, nil
			}
		}
		return 0, false, &database.Error{OrigErr: err, Query: []byte(query)}

	default:
		return version, dirty, nil
	}
}

func (m *Mysql) Drop() error {
	// select all tables
	query := `SHOW TABLES LIKE '%'`
	tables, err := m.conn.QueryContext(context.Background(), query)
	if err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}
	defer tables.Close()

	// delete one table after another
	tableNames := make([]string, 0)
	for tables.Next() {
		var tableName string
		if err := tables.Scan(&tableName); err != nil {
			return err
		}
		if len(tableName) > 0 {
			tableNames = append(tableNames, tableName)
		}
	}

	if len(tableNames) > 0 {
		// delete one by one ...
		for _, t := range tableNames {
			query = "DROP TABLE IF EXISTS `" + t + "` CASCADE"
			if _, err := m.conn.ExecContext(context.Background(), query); err != nil {
				return &database.Error{OrigErr: err, Query: []byte(query)}
			}
		}
		if err := m.ensureVersionTable(); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mysql) ensureVersionTable() error {
	// check if migration table exists
	var result string
	query := `SHOW TABLES LIKE "` + m.config.MigrationsTable + `"`
	if err := m.conn.QueryRowContext(context.Background(), query).Scan(&result); err != nil {
		if err != sql.ErrNoRows {
			return &database.Error{OrigErr: err, Query: []byte(query)}
		}
	} else {
		return nil
	}

	// if not, create the empty migration table
	query = "CREATE TABLE `" + m.config.MigrationsTable + "` (version bigint not null primary key, dirty boolean not null)"
	if _, err := m.conn.ExecContext(context.Background(), query); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}
	return nil
}

// Returns the bool value of the input.
// The 2nd return value indicates if the input was a valid bool value
// See https://github.com/go-sql-driver/mysql/blob/a059889267dc7170331388008528b3b44479bffb/utils.go#L71
func readBool(input string) (value bool, valid bool) {
	switch input {
	case "1", "true", "TRUE", "True":
		return true, true
	case "0", "false", "FALSE", "False":
		return false, true
	}

	// Not a valid bool value
	return
}
