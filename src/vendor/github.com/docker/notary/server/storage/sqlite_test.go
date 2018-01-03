// +build !mysqldb,!rethinkdb

// Initializes an SQLlite DBs for testing purposes

package storage

import (
	"io/ioutil"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func sqlite3Setup(t *testing.T) (*SQLStorage, func()) {
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err)

	dbStore := SetupSQLDB(t, "sqlite3", tempBaseDir+"test_db")
	var cleanup = func() {
		dbStore.DB.Close()
		os.RemoveAll(tempBaseDir)
	}
	return dbStore, cleanup
}

func init() {
	sqldbSetup = sqlite3Setup
}
