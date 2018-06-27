package cassandra

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	nurl "net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/golang-migrate/migrate/database"
)

func init() {
	db := new(Cassandra)
	database.Register("cassandra", db)
}

var DefaultMigrationsTable = "schema_migrations"

var (
	ErrNilConfig     = errors.New("no config")
	ErrNoKeyspace    = errors.New("no keyspace provided")
	ErrDatabaseDirty = errors.New("database is dirty")
	ErrClosedSession = errors.New("session is closed")
)

type Config struct {
	MigrationsTable string
	KeyspaceName    string
}

type Cassandra struct {
	session  *gocql.Session
	isLocked bool

	// Open and WithInstance need to guarantee that config is never nil
	config *Config
}

func WithInstance(session *gocql.Session, config *Config) (database.Driver, error) {
	if config == nil {
		return nil, ErrNilConfig
	} else if len(config.KeyspaceName) == 0 {
		return nil, ErrNoKeyspace
	}

	if session.Closed() {
		return nil, ErrClosedSession
	}

	if len(config.MigrationsTable) == 0 {
		config.MigrationsTable = DefaultMigrationsTable
	}

	c := &Cassandra{
		session: session,
		config:  config,
	}

	if err := c.ensureVersionTable(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Cassandra) Open(url string) (database.Driver, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	// Check for missing mandatory attributes
	if len(u.Path) == 0 {
		return nil, ErrNoKeyspace
	}

	cluster := gocql.NewCluster(u.Host)
	cluster.Keyspace = strings.TrimPrefix(u.Path, "/")
	cluster.Consistency = gocql.All
	cluster.Timeout = 1 * time.Minute

	if len(u.Query().Get("username")) > 0 && len(u.Query().Get("password")) > 0 {
		authenticator := gocql.PasswordAuthenticator{
			Username: u.Query().Get("username"),
			Password: u.Query().Get("password"),
		}
		cluster.Authenticator = authenticator
	}

	// Retrieve query string configuration
	if len(u.Query().Get("consistency")) > 0 {
		var consistency gocql.Consistency
		consistency, err = parseConsistency(u.Query().Get("consistency"))
		if err != nil {
			return nil, err
		}

		cluster.Consistency = consistency
	}
	if len(u.Query().Get("protocol")) > 0 {
		var protoversion int
		protoversion, err = strconv.Atoi(u.Query().Get("protocol"))
		if err != nil {
			return nil, err
		}
		cluster.ProtoVersion = protoversion
	}
	if len(u.Query().Get("timeout")) > 0 {
		var timeout time.Duration
		timeout, err = time.ParseDuration(u.Query().Get("timeout"))
		if err != nil {
			return nil, err
		}
		cluster.Timeout = timeout
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return WithInstance(session, &Config{
		KeyspaceName:    strings.TrimPrefix(u.Path, "/"),
		MigrationsTable: u.Query().Get("x-migrations-table"),
	})
}

func (c *Cassandra) Close() error {
	c.session.Close()
	return nil
}

func (c *Cassandra) Lock() error {
	if c.isLocked {
		return database.ErrLocked
	}
	c.isLocked = true
	return nil
}

func (c *Cassandra) Unlock() error {
	c.isLocked = false
	return nil
}

func (c *Cassandra) Run(migration io.Reader) error {
	migr, err := ioutil.ReadAll(migration)
	if err != nil {
		return err
	}
	// run migration
	query := string(migr[:])
	if err := c.session.Query(query).Exec(); err != nil {
		// TODO: cast to Cassandra error and get line number
		return database.Error{OrigErr: err, Err: "migration failed", Query: migr}
	}

	return nil
}

func (c *Cassandra) SetVersion(version int, dirty bool) error {
	query := `TRUNCATE "` + c.config.MigrationsTable + `"`
	if err := c.session.Query(query).Exec(); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}
	if version >= 0 {
		query = `INSERT INTO "` + c.config.MigrationsTable + `" (version, dirty) VALUES (?, ?)`
		if err := c.session.Query(query, version, dirty).Exec(); err != nil {
			return &database.Error{OrigErr: err, Query: []byte(query)}
		}
	}

	return nil
}

// Return current keyspace version
func (c *Cassandra) Version() (version int, dirty bool, err error) {
	query := `SELECT version, dirty FROM "` + c.config.MigrationsTable + `" LIMIT 1`
	err = c.session.Query(query).Scan(&version, &dirty)
	switch {
	case err == gocql.ErrNotFound:
		return database.NilVersion, false, nil

	case err != nil:
		if _, ok := err.(*gocql.Error); ok {
			return database.NilVersion, false, nil
		}
		return 0, false, &database.Error{OrigErr: err, Query: []byte(query)}

	default:
		return version, dirty, nil
	}
}

func (c *Cassandra) Drop() error {
	// select all tables in current schema
	query := fmt.Sprintf(`SELECT table_name from system_schema.tables WHERE keyspace_name='%s'`, c.config.KeyspaceName)
	iter := c.session.Query(query).Iter()
	var tableName string
	for iter.Scan(&tableName) {
		err := c.session.Query(fmt.Sprintf(`DROP TABLE %s`, tableName)).Exec()
		if err != nil {
			return err
		}
	}
	// Re-create the version table
	return c.ensureVersionTable()
}

// Ensure version table exists
func (c *Cassandra) ensureVersionTable() error {
	err := c.session.Query(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (version bigint, dirty boolean, PRIMARY KEY(version))", c.config.MigrationsTable)).Exec()
	if err != nil {
		return err
	}
	if _, _, err = c.Version(); err != nil {
		return err
	}
	return nil
}

// ParseConsistency wraps gocql.ParseConsistency
// to return an error instead of a panicking.
func parseConsistency(consistencyStr string) (consistency gocql.Consistency, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("Failed to parse consistency \"%s\": %v", consistencyStr, r)
			}
		}
	}()
	consistency = gocql.ParseConsistency(consistencyStr)

	return consistency, nil
}
