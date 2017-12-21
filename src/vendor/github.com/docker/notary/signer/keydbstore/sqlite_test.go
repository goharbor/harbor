// +build !mysqldb

// Initializes an SQLlite DBs for testing purposes

package keydbstore

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func sqlite3Setup(t *testing.T) (*SQLKeyDBStore, func()) {
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err)

	dbStore := SetupSQLDB(t, "sqlite3", filepath.Join(tempBaseDir, "test_db"))
	var cleanup = func() {
		dbStore.db.Close()
		os.RemoveAll(tempBaseDir)
	}

	require.Equal(t, "sqlite3", dbStore.Name())
	return dbStore, cleanup
}

func init() {
	sqldbSetup = sqlite3Setup
}
