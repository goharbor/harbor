// !build rethinkdb

package keydbstore

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/dvsekhvalnov/jose2go"
	"github.com/stretchr/testify/require"
)

// not to the nanosecond scale because mysql timestamps ignore nanoseconds
var gormActiveTime = time.Date(2016, 12, 31, 1, 1, 1, 0, time.UTC)

func SetupSQLDB(t *testing.T, dbtype, dburl string) *SQLKeyDBStore {
	dbStore, err := NewSQLKeyDBStore(multiAliasRetriever, validAliases[0], dbtype, dburl)
	require.NoError(t, err)
	dbStore.nowFunc = func() time.Time { return gormActiveTime }

	// Create the DB tables if they don't exist
	dbStore.db.CreateTable(&GormPrivateKey{})

	// verify that the table is empty
	var count int
	query := dbStore.db.Model(&GormPrivateKey{}).Count(&count)
	require.NoError(t, query.Error)
	require.Equal(t, 0, count)

	return dbStore
}

type sqldbSetupFunc func(*testing.T) (*SQLKeyDBStore, func())

var sqldbSetup sqldbSetupFunc

// Creating a new KeyDBStore propagates any db opening error
func TestNewSQLKeyDBStorePropagatesDBError(t *testing.T) {
	dbStore, err := NewSQLKeyDBStore(constRetriever, "ignoredalias", "nodb", "somestring")
	require.Error(t, err)
	require.Nil(t, dbStore)
}

func TestSQLDBHealthCheckMissingTable(t *testing.T) {
	dbStore, cleanup := sqldbSetup(t)
	defer cleanup()

	// health check passes because the table exists
	require.NoError(t, dbStore.HealthCheck())

	// delete the table - health check fails
	require.NoError(t, dbStore.db.DropTableIfExists(&GormPrivateKey{}).Error)
	require.Error(t, dbStore.HealthCheck())
}

func TestSQLDBHealthCheckNoConnection(t *testing.T) {
	dbStore, cleanup := sqldbSetup(t)
	defer cleanup()

	// health check passes because the table exists and connection is open
	require.NoError(t, dbStore.HealthCheck())

	// Close the connection - health check fails
	require.NoError(t, dbStore.db.Close())
	require.Error(t, dbStore.HealthCheck())
}

// Checks that the DB contains the expected keys, and returns a map of the GormPrivateKey object by key ID
func requireExpectedGORMKeys(t *testing.T, dbStore *SQLKeyDBStore, expectedKeys []data.PrivateKey) map[string]GormPrivateKey {
	var rows []GormPrivateKey
	query := dbStore.db.Find(&rows)
	require.NoError(t, query.Error)

	require.Len(t, rows, len(expectedKeys))
	result := make(map[string]GormPrivateKey)

	for _, gormKey := range rows {
		result[gormKey.KeyID] = gormKey
	}

	for _, key := range expectedKeys {
		gormKey, ok := result[key.ID()]
		require.True(t, ok)
		require.NotNil(t, gormKey)
		require.Equal(t, string(key.Public()), gormKey.Public)
		require.Equal(t, key.Algorithm(), gormKey.Algorithm)
	}

	return result
}

func TestSQLKeyCanOnlyBeAddedOnce(t *testing.T) {
	dbStore, cleanup := sqldbSetup(t)
	defer cleanup()

	expectedKeys := testKeyCanOnlyBeAddedOnce(t, dbStore)

	gormKeys := requireExpectedGORMKeys(t, dbStore, expectedKeys)

	// none of these keys are active, since they have not been activated
	for _, gormKey := range gormKeys {
		require.True(t, gormKey.LastUsed.Equal(time.Time{}))
	}
}

func TestSQLCreateDelete(t *testing.T) {
	dbStore, cleanup := sqldbSetup(t)
	defer cleanup()
	expectedKeys := testCreateDelete(t, dbStore)

	gormKeys := requireExpectedGORMKeys(t, dbStore, expectedKeys)

	// none of these keys are active, since they have not been activated
	for _, gormKey := range gormKeys {
		require.True(t, gormKey.LastUsed.Equal(time.Time{}))
	}
}

func TestSQLKeyRotation(t *testing.T) {
	dbStore, cleanup := sqldbSetup(t)
	defer cleanup()

	rotatedKey, nonRotatedKey := testKeyRotation(t, dbStore, validAliases[1])

	gormKeys := requireExpectedGORMKeys(t, dbStore, []data.PrivateKey{rotatedKey, nonRotatedKey})

	// none of these keys are active, since they have not been activated
	for _, gormKey := range gormKeys {
		require.True(t, gormKey.LastUsed.Equal(time.Time{}))
	}

	// require that the rotated key is encrypted with the new passphrase
	rotatedGormKey := gormKeys[rotatedKey.ID()]
	require.Equal(t, validAliases[1], rotatedGormKey.PassphraseAlias)
	decryptedKey, _, err := jose.Decode(string(rotatedGormKey.Private), validAliasesAndPasswds[validAliases[1]])
	require.NoError(t, err)
	require.Equal(t, string(rotatedKey.Private()), decryptedKey)

	// require that the nonrotated key is encrypted with the old passphrase
	nonRotatedGormKey := gormKeys[nonRotatedKey.ID()]
	require.Equal(t, validAliases[0], nonRotatedGormKey.PassphraseAlias)
	decryptedKey, _, err = jose.Decode(string(nonRotatedGormKey.Private), validAliasesAndPasswds[validAliases[0]])
	require.NoError(t, err)
	require.Equal(t, string(nonRotatedKey.Private()), decryptedKey)
}

func TestSQLSigningMarksKeyActive(t *testing.T) {
	dbStore, cleanup := sqldbSetup(t)
	defer cleanup()

	activeKey, nonActiveKey := testSigningWithKeyMarksAsActive(t, dbStore)

	gormKeys := requireExpectedGORMKeys(t, dbStore, []data.PrivateKey{activeKey, nonActiveKey})

	// check that activation updates the activated key but not the unactivated key
	require.True(t, gormKeys[activeKey.ID()].LastUsed.Equal(gormActiveTime))
	require.True(t, gormKeys[nonActiveKey.ID()].LastUsed.Equal(time.Time{}))

	// check that signing succeeds even if the DB connection is closed and hence
	// mark as active errors
	dbStore.db.Close()
	msg := []byte("successful, db closed")
	sig, err := nonActiveKey.Sign(rand.Reader, msg, nil)
	require.NoError(t, err)
	require.NoError(t, signed.Verifiers[data.ECDSASignature].Verify(
		data.PublicKeyFromPrivate(nonActiveKey), sig, msg))
}

func TestSQLCreateKey(t *testing.T) {
	dbStore, cleanup := sqldbSetup(t)
	defer cleanup()

	activeED25519Key, pendingED25519Key, pendingECDSAKey := testCreateKey(t, dbStore)

	gormKeys := requireExpectedGORMKeys(t, dbStore, []data.PrivateKey{activeED25519Key, pendingED25519Key, pendingECDSAKey})

	// check that activation updates the activated key but not the pending key
	require.True(t, gormKeys[activeED25519Key.ID()].LastUsed.Equal(gormActiveTime))
	require.True(t, gormKeys[pendingED25519Key.ID()].LastUsed.Equal(time.Time{}))
	require.True(t, gormKeys[pendingECDSAKey.ID()].LastUsed.Equal(time.Time{}))
}

func TestSQLUnimplementedInterfaceBehavior(t *testing.T) {
	dbStore, cleanup := sqldbSetup(t)
	defer cleanup()
	testUnimplementedInterfaceMethods(t, dbStore)
}
