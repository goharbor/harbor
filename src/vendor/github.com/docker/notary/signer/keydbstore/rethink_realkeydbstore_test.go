// +build rethinkdb

// Uses a real RethinkDB connection testing purposes

package keydbstore

import (
	"crypto/rand"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/tlsconfig"
	"github.com/docker/notary/storage/rethinkdb"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/dvsekhvalnov/jose2go"
	"github.com/stretchr/testify/require"
	"gopkg.in/dancannon/gorethink.v3"
)

var tlsOpts = tlsconfig.Options{InsecureSkipVerify: true}
var rdbNow = time.Date(2016, 12, 31, 1, 1, 1, 0, time.UTC)

func rethinkSessionSetup(t *testing.T) (*gorethink.Session, string) {
	// Get the Rethink connection string from an environment variable
	rethinkSource := os.Getenv("DBURL")
	require.NotEqual(t, "", rethinkSource)

	sess, err := rethinkdb.AdminConnection(tlsOpts, rethinkSource)
	require.NoError(t, err)

	return sess, rethinkSource
}

func rethinkDBSetup(t *testing.T, dbName string) (*RethinkDBKeyStore, func()) {
	session, _ := rethinkSessionSetup(t)
	var cleanup = func() { gorethink.DBDrop(dbName).Exec(session) }

	cleanup()

	err := rethinkdb.SetupDB(session, dbName, []rethinkdb.Table{PrivateKeysRethinkTable})
	require.NoError(t, err)

	dbStore := NewRethinkDBKeyStore(dbName, "", "", multiAliasRetriever, validAliases[0], session)
	require.Equal(t, "RethinkDB", dbStore.Name())

	dbStore.nowFunc = func() time.Time { return rdbNow }

	return dbStore, cleanup
}

func TestRethinkBootstrapSetsUsernamePassword(t *testing.T) {
	adminSession, source := rethinkSessionSetup(t)
	dbname, username, password := "signertestdb", "testuser", "testpassword"
	otherDB, otherUser, otherPass := "othersignertestdb", "otheruser", "otherpassword"

	// create a separate user with access to a different DB
	require.NoError(t, rethinkdb.SetupDB(adminSession, otherDB, nil))
	defer gorethink.DBDrop(otherDB).Exec(adminSession)
	require.NoError(t, rethinkdb.CreateAndGrantDBUser(adminSession, otherDB, otherUser, otherPass))

	// Bootstrap
	s := NewRethinkDBKeyStore(dbname, username, password, constRetriever, "ignored", adminSession)
	require.NoError(t, s.Bootstrap())
	defer gorethink.DBDrop(dbname).Exec(adminSession)

	// A user with an invalid password cannot connect to rethink DB at all
	_, err := rethinkdb.UserConnection(tlsOpts, source, username, "wrongpass")
	require.Error(t, err)

	// the other user cannot access rethink, causing health checks to fail
	userSession, err := rethinkdb.UserConnection(tlsOpts, source, otherUser, otherPass)
	require.NoError(t, err)
	s = NewRethinkDBKeyStore(dbname, otherUser, otherPass, constRetriever, "ignored", userSession)
	_, _, err = s.GetPrivateKey("nonexistent")
	require.Error(t, err)
	require.IsType(t, gorethink.RQLRuntimeError{}, err)
	key := s.GetKey("nonexistent")
	require.Nil(t, key)
	require.Error(t, s.CheckHealth())

	// our user can access the DB though
	userSession, err = rethinkdb.UserConnection(tlsOpts, source, username, password)
	require.NoError(t, err)
	s = NewRethinkDBKeyStore(dbname, username, password, constRetriever, "ignored", userSession)
	_, _, err = s.GetPrivateKey("nonexistent")
	require.Error(t, err)
	require.IsType(t, trustmanager.ErrKeyNotFound{}, err)
	require.NoError(t, s.CheckHealth())
}

// Checks that the DB contains the expected keys, and returns a map of the GormPrivateKey object by key ID
func requireExpectedRDBKeys(t *testing.T, dbStore *RethinkDBKeyStore, expectedKeys []data.PrivateKey) map[string]RDBPrivateKey {
	res, err := gorethink.DB(dbStore.dbName).Table(PrivateKeysRethinkTable.Name).Run(dbStore.sess)
	require.NoError(t, err)

	var rows []RDBPrivateKey
	require.NoError(t, res.All(&rows))

	require.Len(t, rows, len(expectedKeys))
	result := make(map[string]RDBPrivateKey)

	for _, rdbKey := range rows {
		result[rdbKey.KeyID] = rdbKey
	}

	for _, key := range expectedKeys {
		rdbKey, ok := result[key.ID()]
		require.True(t, ok)
		require.NotNil(t, rdbKey)
		require.Equal(t, key.Public(), rdbKey.Public)
		require.Equal(t, key.Algorithm(), rdbKey.Algorithm)

		// because we have to manually set the created and modified times
		require.True(t, rdbKey.CreatedAt.Equal(rdbNow))
		require.True(t, rdbKey.UpdatedAt.Equal(rdbNow))
		require.True(t, rdbKey.DeletedAt.Equal(time.Time{}))
	}

	return result
}

func TestRethinkKeyCanOnlyBeAddedOnce(t *testing.T) {
	dbStore, cleanup := rethinkDBSetup(t, "signerAddTests")
	defer cleanup()

	expectedKeys := testKeyCanOnlyBeAddedOnce(t, dbStore)

	rdbKeys := requireExpectedRDBKeys(t, dbStore, expectedKeys)

	// none of these keys are active, since they have not been activated
	for _, rdbKey := range rdbKeys {
		require.True(t, rdbKey.LastUsed.Equal(time.Time{}))
	}
}

func TestRethinkCreateDelete(t *testing.T) {
	dbStore, cleanup := rethinkDBSetup(t, "signerDeleteTests")
	defer cleanup()
	expectedKeys := testCreateDelete(t, dbStore)

	rdbKeys := requireExpectedRDBKeys(t, dbStore, expectedKeys)

	// none of these keys are active, since they have not been activated
	for _, rdbKey := range rdbKeys {
		require.True(t, rdbKey.LastUsed.Equal(time.Time{}))
	}
}

func TestRethinkKeyRotation(t *testing.T) {
	dbStore, cleanup := rethinkDBSetup(t, "signerRotationTests")
	defer cleanup()

	rotatedKey, nonRotatedKey := testKeyRotation(t, dbStore, validAliases[1])

	rdbKeys := requireExpectedRDBKeys(t, dbStore, []data.PrivateKey{rotatedKey, nonRotatedKey})

	// none of these keys are active, since they have not been activated
	for _, rdbKey := range rdbKeys {
		require.True(t, rdbKey.LastUsed.Equal(time.Time{}))
	}

	// require that the rotated key is encrypted with the new passphrase
	rotatedRDBKey := rdbKeys[rotatedKey.ID()]
	require.Equal(t, validAliases[1], rotatedRDBKey.PassphraseAlias)
	decryptedKey, _, err := jose.Decode(string(rotatedRDBKey.Private), validAliasesAndPasswds[validAliases[1]])
	require.NoError(t, err)
	require.Equal(t, string(rotatedKey.Private()), decryptedKey)

	// require that the nonrotated key is encrypted with the old passphrase
	nonRotatedRDBKey := rdbKeys[nonRotatedKey.ID()]
	require.Equal(t, validAliases[0], nonRotatedRDBKey.PassphraseAlias)
	decryptedKey, _, err = jose.Decode(string(nonRotatedRDBKey.Private), validAliasesAndPasswds[validAliases[0]])
	require.NoError(t, err)
	require.Equal(t, string(nonRotatedKey.Private()), decryptedKey)
}

func TestRethinkCheckHealth(t *testing.T) {
	dbStore, cleanup := rethinkDBSetup(t, "signerHealthcheckTests")
	defer cleanup()

	// sanity check - all tables present - health check passes
	require.NoError(t, dbStore.CheckHealth())

	// if the DB is unreachable, health check fails
	require.NoError(t, dbStore.sess.Close())
	require.Error(t, dbStore.CheckHealth())

	// if the connection is reopened, health check succeeds
	require.NoError(t, dbStore.sess.Reconnect())
	require.NoError(t, dbStore.CheckHealth())

	// No tables, health check fails
	require.NoError(t, gorethink.DB(dbStore.dbName).TableDrop(PrivateKeysRethinkTable.Name).Exec(dbStore.sess))
	require.Error(t, dbStore.CheckHealth())

	// No DB, health check fails
	cleanup()
	require.Error(t, dbStore.CheckHealth())
}

func TestRethinkSigningMarksKeyActive(t *testing.T) {
	dbStore, cleanup := rethinkDBSetup(t, "signerActivationTests")
	defer cleanup()

	activeKey, nonActiveKey := testSigningWithKeyMarksAsActive(t, dbStore)

	rdbKeys := requireExpectedRDBKeys(t, dbStore, []data.PrivateKey{activeKey, nonActiveKey})

	// check that activation updates the activated key but not the unactivated key
	require.True(t, rdbKeys[activeKey.ID()].LastUsed.Equal(rdbNow))
	require.True(t, rdbKeys[nonActiveKey.ID()].LastUsed.Equal(time.Time{}))

	// check that signing succeeds even if the DB connection is closed and hence
	// mark as active errors
	dbStore.sess.Close()
	msg := []byte("successful, db closed")
	sig, err := nonActiveKey.Sign(rand.Reader, msg, nil)
	require.NoError(t, err)
	require.NoError(t, signed.Verifiers[data.ECDSASignature].Verify(
		data.PublicKeyFromPrivate(nonActiveKey), sig, msg))
}

func TestRethinkCreateKey(t *testing.T) {
	dbStore, cleanup := rethinkDBSetup(t, "signerCreationTests")
	defer cleanup()

	activeED25519Key, pendingED25519Key, pendingECDSAKey := testCreateKey(t, dbStore)

	rdbKeys := requireExpectedRDBKeys(t, dbStore, []data.PrivateKey{activeED25519Key, pendingED25519Key, pendingECDSAKey})

	// check that activation updates the activated key but not the unactivated keys
	require.True(t, rdbKeys[activeED25519Key.ID()].LastUsed.Equal(rdbNow))
	require.True(t, rdbKeys[pendingED25519Key.ID()].LastUsed.Equal(time.Time{}))
	require.True(t, rdbKeys[pendingECDSAKey.ID()].LastUsed.Equal(time.Time{}))
}

func TestRethinkUnimplementedInterfaceBehavior(t *testing.T) {
	dbStore, cleanup := rethinkDBSetup(t, "signerInterfaceTests")
	defer cleanup()
	testUnimplementedInterfaceMethods(t, dbStore)
}
