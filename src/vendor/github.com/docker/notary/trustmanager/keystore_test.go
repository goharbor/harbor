package trustmanager

import (
	"crypto/rand"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/notary"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/utils"
	"github.com/stretchr/testify/require"
)

const cannedPassphrase = "passphrase"

var passphraseRetriever = func(keyID string, alias string, createNew bool, numAttempts int) (string, bool, error) {
	if numAttempts > 5 {
		giveup := true
		return "", giveup, errors.New("passPhraseRetriever failed after too many requests")
	}
	return cannedPassphrase, false, nil
}

func TestAddKey(t *testing.T) {
	testAddKeyWithRole(t, data.CanonicalRootRole)
	testAddKeyWithRole(t, data.CanonicalTargetsRole)
	testAddKeyWithRole(t, data.CanonicalSnapshotRole)
	testAddKeyWithRole(t, "targets/a/b/c")
	testAddKeyWithRole(t, "invalidRole")
}

func testAddKeyWithRole(t *testing.T, role data.RoleName) {
	var gun data.GUN = "docker.com/notary"
	testExt := "key"

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)
	// Create our store
	store, err := NewKeyFileStore(tempBaseDir, passphraseRetriever)
	require.NoError(t, err, "failed to create new key filestore")

	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, "could not generate private key")

	// Since we're generating this manually we need to add the extension '.'
	expectedFilePath := filepath.Join(tempBaseDir, notary.PrivDir, privKey.ID()+"."+testExt)

	// Call the AddKey function
	err = store.AddKey(KeyInfo{Role: role, Gun: gun}, privKey)
	require.NoError(t, err, "failed to add key to store")

	// Check to see if file exists
	b, err := ioutil.ReadFile(expectedFilePath)
	require.NoError(t, err, "expected file not found")
	require.Contains(t, string(b), "-----BEGIN ENCRYPTED PRIVATE KEY-----")

	// Check that we have the role and gun info for this key's ID
	keyInfo, ok := store.keyInfoMap[privKey.ID()]
	require.True(t, ok)
	require.Equal(t, role, keyInfo.Role)
	if role == data.CanonicalRootRole || data.IsDelegation(role) || !data.ValidRole(role) {
		require.Empty(t, keyInfo.Gun.String())
	} else {
		require.EqualValues(t, gun, keyInfo.Gun.String())
	}
}

func TestKeyStoreInternalState(t *testing.T) {
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	var gun data.GUN = "docker.com/notary"

	// Mimic a notary repo setup, and test that bringing up a keyfilestore creates the correct keyInfoMap
	roles := []data.RoleName{data.CanonicalRootRole, data.CanonicalTargetsRole, data.CanonicalSnapshotRole, data.RoleName("targets/delegation")}
	// Keep track of the key IDs for each role, so we can validate later against the keystore state
	roleToID := make(map[string]string)
	for _, role := range roles {
		// generate a key for the role
		privKey, err := utils.GenerateECDSAKey(rand.Reader)
		require.NoError(t, err, "could not generate private key")

		var privKeyPEM []byte
		// generate the correct PEM role header
		if role == data.CanonicalRootRole || data.IsDelegation(role) || !data.ValidRole(role) {
			privKeyPEM, err = utils.ConvertPrivateKeyToPKCS8(privKey, role, "", "")
		} else {
			privKeyPEM, err = utils.ConvertPrivateKeyToPKCS8(privKey, role, gun, "")
		}

		require.NoError(t, err, "could not generate PEM")

		// write the key file to the correct location
		keyPath := filepath.Join(tempBaseDir, notary.PrivDir)
		keyPath = filepath.Join(keyPath, privKey.ID())
		require.NoError(t, os.MkdirAll(filepath.Dir(keyPath), 0755))
		require.NoError(t, ioutil.WriteFile(keyPath+".key", privKeyPEM, 0755))

		roleToID[role.String()] = privKey.ID()
	}

	store, err := NewKeyFileStore(tempBaseDir, passphraseRetriever)
	require.NoError(t, err)
	require.Len(t, store.keyInfoMap, 4)
	for _, role := range roles {
		keyID, _ := roleToID[role.String()]
		// make sure this keyID is the right length
		require.Len(t, keyID, notary.SHA256HexSize)
		require.Equal(t, role, store.keyInfoMap[keyID].Role)
		// targets and snapshot keys should have a gun set, root and delegation keys should not
		if role == data.CanonicalTargetsRole || role == data.CanonicalSnapshotRole {
			require.EqualValues(t, gun, store.keyInfoMap[keyID].Gun.String())
		} else {
			require.Empty(t, store.keyInfoMap[keyID].Gun.String())
		}
	}

	// Try removing the targets key only by ID (no gun provided)
	require.NoError(t, store.RemoveKey(roleToID[data.CanonicalTargetsRole.String()]))
	// The key file itself should have been removed
	_, err = os.Stat(filepath.Join(tempBaseDir, notary.PrivDir, roleToID[data.CanonicalTargetsRole.String()]+".key"))
	require.Error(t, err)
	// The keyInfoMap should have also updated by deleting the key
	_, ok := store.keyInfoMap[roleToID[data.CanonicalTargetsRole.String()]]
	require.False(t, ok)

	// Try removing the delegation key only by ID (no gun provided)
	require.NoError(t, store.RemoveKey(roleToID["targets/delegation"]))
	// The key file itself should have been removed
	_, err = os.Stat(filepath.Join(tempBaseDir, notary.PrivDir, roleToID["targets/delegation"]+".key"))
	require.Error(t, err)
	// The keyInfoMap should have also updated
	_, ok = store.keyInfoMap[roleToID["targets/delegation"]]
	require.False(t, ok)

	// Try removing the root key only by ID (no gun provided)
	require.NoError(t, store.RemoveKey(roleToID[data.CanonicalRootRole.String()]))
	// The key file itself should have been removed
	_, err = os.Stat(filepath.Join(tempBaseDir, notary.PrivDir, roleToID[data.CanonicalRootRole.String()]+".key"))
	require.Error(t, err)
	// The keyInfoMap should have also updated_
	_, ok = store.keyInfoMap[roleToID[data.CanonicalRootRole.String()]]
	require.False(t, ok)

	// Generate a new targets key and add it with its gun, check that the map gets updated back
	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, "could not generate private key")
	require.NoError(t, store.AddKey(KeyInfo{Role: data.CanonicalTargetsRole, Gun: gun}, privKey))
	require.Equal(t, gun, store.keyInfoMap[privKey.ID()].Gun)
	require.Equal(t, data.CanonicalTargetsRole, store.keyInfoMap[privKey.ID()].Role)
}

func TestGet(t *testing.T) {
	nonRootRolesToTest := []data.RoleName{
		data.CanonicalTargetsRole,
		data.CanonicalSnapshotRole,
		"targets/a/b/c",
		"invalidRole",
	}

	var gun data.GUN = "docker.io/notary"

	testGetKeyWithRole(t, "", data.CanonicalRootRole)
	for _, role := range nonRootRolesToTest {
		testGetKeyWithRole(t, data.GUN(""), role)
		testGetKeyWithRole(t, gun, role)
	}
}

func testGetKeyWithRole(t *testing.T, gun data.GUN, role data.RoleName) {
	var testPEM []byte
	testPEM = []byte(`-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC2cL8WamG24ihl
JSVG8ZVel05lPqYD0S8ol1L+zzwsHkim2DS+a5BLX5+QJtCfZrR+Pzo+4pCrjU+N
R/71aYNm/M95h/JSJxdEoTgYCCHNJD8IYpTc6lXyy49lSQh7svLpZ2dQwHoGB5VC
tpsh8xvLLbXfk/G7ihEeZqG7/Tnoe+uotkiODOTjxiTGvQQjoAc4hQgzGH4sjC7U
8E8zB0j1BQWM/fhRX/ww3V/SRB2T1u0aAurF1BnUdDazZMBxWQ7DxmY3FNbeNXqf
KKeQMN1Rodu8hJw0gxL1hbOWmcYksmGZfPDzYXiHBdscCFr/wimOl9BO/o2xbV5+
phbph9cFAgMBAAECggEBAIAcA9L1uM/3V25O+zIqCj11+jLWHzWm+nqCaGFNnG9O
hK3EPKVKWvTSnPVYjD6inDPaqkfmSLhubmJDICGsif0ToY0xjVNq58flfcJCU5n9
zdVRhD7svpXTo0n4UuCp9DE5zy7BOe5p/MHwAFeCow21d3UcKi8K8KJsZz3ev38j
9Y8ASd24NcyZfE4mnjDjA/MuzlPoQYMwAh4f3mrEKu5v9dCT+m70lJTzSNAc4gD0
93mMkGRsUKjvZyCu/IlXncBczaSVovX5IGdiGPa7Qk+CP9r+PGQUasb+e5o7VMzh
xyjIrCV1u48vRyJsc7xrZ+PUkVk74u9mQ3wxQXNzi7ECgYEA5BftyMlzv2oqAzQg
isS0f616qX5YmRK/riC/4+HRaXEsA/LiI8tuW04vdgcelUqxo1TFpv+J4z16ItF5
kscb6ev9wsFa0VInsvI3hqZ8e4AuqlvU8Rii1anxkbwE5mstRgeR9p410+0T2GiW
JaWVy8mxsneVI0sdR5ooJ+ZBQpcCgYEAzMLtV52aQvnCLPejPI+fBnOjoLXTVaaB
xqZWfOzuozjYVlqSUsKbKbMVtIy+rPIJt26/qw8i6V8Dx2HlUcySU5fAumpWigK4
Dh64eZ+yJrQeqgRJoLoZhTbgxe4fv7+f649WcipwD0ptEaqjD11Wdr0973tw0wdc
Pqn9SlPoksMCgYBqUKj5xMRZvQ82DQ75/3Oua1rYM9byCmYjsIogmrn0Ltb4RDaZ
vpGCp2/B0NG1fmpMGhBCpatMqvQJ1J+ZBYuCPgg6xcsh8+wjIXk2HtW47udRappX
gkcr1hmN9xhFmkEw+ghT7ixiyodMgHszsvmeUjWsXMa7+5/7JuR+rHlQowKBgE0T
Lr3lMDT3yJSeno5kTWrjSntrFeLOq1j4MeQSV32PHzfaHewTHs7if1AYDooRDYFD
qdgc+Xo47rY1blmNFKNsovpInsySW2/NNolpiGizMjuzI3fhtUuErbUzfjXyTqMf
sF2HBelrjYSx43EcJDjL4S1tHLoCskFQQWyiCxB7AoGBANSohPiPmJLvCEmZTdHm
KcRNz9jE0wO5atCZADIfuOrYHYTQk3YTI5V3GviUNLdmbw4TQChwAgAYVNth1rpL
5jSqfF3RtNBePZixG2WzxYd2ZwvJxvKa33i1E8UfM+yEZH4Gc5ukDt28m0fyFBmi
QvS5quTEllrvrVuWfhpsjl/l
-----END PRIVATE KEY-----
`)
	testBlock, _ := pem.Decode(testPEM)
	require.NotEmpty(t, testBlock, "could not decode pem")

	testPrivKey, err := utils.ParsePKCS8ToTufKey(testBlock.Bytes, nil)
	require.NoError(t, err, "could not parse pkcs8 key")

	testData, err := utils.ConvertPrivateKeyToPKCS8(testPrivKey, role, gun, "")
	require.NoError(t, err, "could not wrap pkcs8 key")

	testName := "keyID"
	testExt := "key"
	perms := os.FileMode(0755)

	emptyPassphraseRetriever := func(string, string, bool, int) (string, bool, error) { return "", false, nil }

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	// Since we're generating this manually we need to add the extension '.'
	filePath := filepath.Join(tempBaseDir, notary.PrivDir, testName+"."+testExt)
	os.MkdirAll(filepath.Dir(filePath), perms)
	err = ioutil.WriteFile(filePath, testData, perms)
	require.NoError(t, err, "failed to write test file")

	// Create our store
	store, err := NewKeyFileStore(tempBaseDir, emptyPassphraseRetriever)
	require.NoError(t, err, "failed to create new key filestore")

	// Call the GetKey function
	privKey, _, err := store.GetKey(testName)
	require.NoError(t, err, "failed to get %s key from store (it's in %s)", role, filepath.Join(tempBaseDir, notary.PrivDir))

	pemPrivKey, err := utils.ConvertPrivateKeyToPKCS8(privKey, role, gun, "")
	require.NoError(t, err, "failed to convert key to PEM")
	require.Equal(t, testData, pemPrivKey)
}

// TestGetLegacyKey ensures we can still load keys where the role
// is stored as part of the filename (i.e. <hexID>_<role>.key
func TestGetLegacyKey(t *testing.T) {
	if notary.FIPSEnabled() {
		t.Skip("skip backward compatibility test in FIPS mode")
	}
	testData := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAyUIXjsrWRrvPa4Bzp3VJ6uOUGPay2fUpSV8XzNxZxIG/Opdr
+k3EQi1im6WOqF3Y5AS1UjYRxNuRN+cAZeo3uS1pOTuoSupBXuchVw8s4hZJ5vXn
TRmGb+xY7tZ1ZVgPfAZDib9sRSUsL/gC+aSyprAjG/YBdbF06qKbfOfsoCEYW1OQ
82JqHzQH514RFYPTnEGpvfxWaqmFQLmv0uMxV/cAYvqtrGkXuP0+a8PknlD2obw5
0rHE56Su1c3Q42S7L51K38tpbgWOSRcTfDUWEj5v9wokkNQvyKBwbS996s4EJaZd
7r6M0h1pHnuRxcSaZLYRwgOe1VNGg2VfWzgd5QIDAQABAoIBAF9LGwpygmj1jm3R
YXGd+ITugvYbAW5wRb9G9mb6wspnwNsGTYsz/UR0ZudZyaVw4jx8+jnV/i3e5PC6
QRcAgqf8l4EQ/UuThaZg/AlT1yWp9g4UyxNXja87EpTsGKQGwTYxZRM4/xPyWOzR
mt8Hm8uPROB9aA2JG9npaoQG8KSUj25G2Qot3ukw/IOtqwN/Sx1EqF0EfCH1K4KU
a5TrqlYDFmHbqT1zTRec/BTtVXNsg8xmF94U1HpWf3Lpg0BPYT7JiN2DPoLelRDy
a/A+a3ZMRNISL5wbq/jyALLOOyOkIqa+KEOeW3USuePd6RhDMzMm/0ocp5FCwYfo
k4DDeaECgYEA0eSMD1dPGo+u8UTD8i7ZsZCS5lmXLNuuAg5f5B/FGghD8ymPROIb
dnJL5QSbUpmBsYJ+nnO8RiLrICGBe7BehOitCKi/iiZKJO6edrfNKzhf4XlU0HFl
jAOMa975pHjeCoZ1cXJOEO9oW4SWTCyBDBSqH3/ZMgIOiIEk896lSmkCgYEA9Xf5
Jqv3HtQVvjugV/axAh9aI8LMjlfFr9SK7iXpY53UdcylOSWKrrDok3UnrSEykjm7
UL3eCU5jwtkVnEXesNn6DdYo3r43E6iAiph7IBkB5dh0yv3vhIXPgYqyTnpdz4pg
3yPGBHMPnJUBThg1qM7k6a2BKHWySxEgC1DTMB0CgYAGvdmF0J8Y0k6jLzs/9yNE
4cjmHzCM3016gW2xDRgumt9b2xTf+Ic7SbaIV5qJj6arxe49NqhwdESrFohrKaIP
kM2l/o2QaWRuRT/Pvl2Xqsrhmh0QSOQjGCYVfOb10nAHVIRHLY22W4o1jk+piLBo
a+1+74NRaOGAnu1J6/fRKQKBgAF180+dmlzemjqFlFCxsR/4G8s2r4zxTMXdF+6O
3zKuj8MbsqgCZy7e8qNeARxwpCJmoYy7dITNqJ5SOGSzrb2Trn9ClP+uVhmR2SH6
AlGQlIhPn3JNzI0XVsLIloMNC13ezvDE/7qrDJ677EQQtNEKWiZh1/DrsmHr+irX
EkqpAoGAJWe8PC0XK2RE9VkbSPg9Ehr939mOLWiHGYTVWPttUcum/rTKu73/X/mj
WxnPWGtzM1pHWypSokW90SP4/xedMxludvBvmz+CTYkNJcBGCrJumy11qJhii9xp
EMl3eFOJXjIch/wIesRSN+2dGOsl7neercjMh1i9RvpCwHDx/E0=
-----END RSA PRIVATE KEY-----
`)
	testName := "docker.com/notary/root"
	testExt := "key"
	testAlias := "root"
	perms := os.FileMode(0755)

	emptyPassphraseRetriever := func(string, string, bool, int) (string, bool, error) { return "", false, nil }

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	// Since we're generating this manually we need to add the extension '.'
	filePath := filepath.Join(tempBaseDir, notary.PrivDir, notary.RootKeysSubdir, testName+"_"+testAlias+"."+testExt)

	os.MkdirAll(filepath.Dir(filePath), perms)
	err = ioutil.WriteFile(filePath, testData, perms)
	require.NoError(t, err, "failed to write test file")

	// Create our store
	store, err := NewKeyFileStore(tempBaseDir, emptyPassphraseRetriever)
	require.NoError(t, err, "failed to create new key filestore")

	// Call the GetKey function
	_, role, err := store.GetKey(testAlias)
	require.NoError(t, err, "failed to get key from store")
	require.EqualValues(t, testAlias, role)
}

func TestListKeys(t *testing.T) {
	testName := "docker.com/notary/root"
	perms := os.FileMode(0755)

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	// Create our store
	store, err := NewKeyFileStore(tempBaseDir, passphraseRetriever)
	require.NoError(t, err, "failed to create new key filestore")

	roles := append(data.BaseRoles, "targets/a", "invalidRoleName")

	for i, role := range roles {
		// Make a new key for each role
		privKey, err := utils.GenerateECDSAKey(rand.Reader)
		require.NoError(t, err, "could not generate private key")

		// Call the AddKey function
		gun := data.GUN(filepath.Dir(testName))
		err = store.AddKey(KeyInfo{Role: role, Gun: gun}, privKey)
		require.NoError(t, err, "failed to add key to store")

		// Check to see if the keystore lists this key
		keyMap := store.ListKeys()

		// Expect to see exactly one key in the map
		require.Len(t, keyMap, i+1)
		// Expect to see privKeyID inside of the map
		listedInfo, ok := keyMap[privKey.ID()]
		require.True(t, ok)
		require.Equal(t, role, listedInfo.Role)
	}

	// Write an invalid filename to the directory
	filePath := filepath.Join(tempBaseDir, notary.PrivDir, "fakekeyname.key")
	err = ioutil.WriteFile(filePath, []byte("data"), perms)
	require.NoError(t, err, "failed to write test file")

	// Check to see if the keystore still lists two keys
	keyMap := store.ListKeys()
	require.Len(t, keyMap, len(roles))

	// Check that ListKeys() returns a copy of the state
	// so modifying its returned information does not change the underlying store's keyInfo
	for keyID := range keyMap {
		delete(keyMap, keyID)
		_, err = store.GetKeyInfo(keyID)
		require.NoError(t, err)
	}
}

func TestAddGetKeyMemStore(t *testing.T) {
	testAlias := data.CanonicalRootRole

	// Create our store
	store := NewKeyMemoryStore(passphraseRetriever)

	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, "could not generate private key")

	// Call the AddKey function
	err = store.AddKey(KeyInfo{Role: testAlias, Gun: ""}, privKey)
	require.NoError(t, err, "failed to add key to store")

	// Check to see if file exists
	retrievedKey, retrievedAlias, err := store.GetKey(privKey.ID())
	require.NoError(t, err, "failed to get key from store")

	require.Equal(t, retrievedAlias, testAlias)
	require.Equal(t, retrievedKey.Public(), privKey.Public())
	require.Equal(t, retrievedKey.Private(), privKey.Private())
}

func TestAddGetKeyInfoMemStore(t *testing.T) {
	var gun data.GUN = "docker.com/notary"

	// Create our store
	store := NewKeyMemoryStore(passphraseRetriever)

	rootKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, "could not generate private key")

	// Call the AddKey function
	err = store.AddKey(KeyInfo{Role: data.CanonicalRootRole, Gun: ""}, rootKey)
	require.NoError(t, err, "failed to add key to store")

	// Get and validate key info
	rootInfo, err := store.GetKeyInfo(rootKey.ID())
	require.NoError(t, err)
	require.Equal(t, data.CanonicalRootRole, rootInfo.Role)
	require.EqualValues(t, "", rootInfo.Gun)

	targetsKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, "could not generate private key")

	// Call the AddKey function
	err = store.AddKey(KeyInfo{Role: data.CanonicalTargetsRole, Gun: gun}, targetsKey)
	require.NoError(t, err, "failed to add key to store")

	// Get and validate key info
	targetsInfo, err := store.GetKeyInfo(targetsKey.ID())
	require.NoError(t, err)
	require.Equal(t, data.CanonicalTargetsRole, targetsInfo.Role)
	require.Equal(t, gun, targetsInfo.Gun)

	delgKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, "could not generate private key")

	// Call the AddKey function
	err = store.AddKey(KeyInfo{Role: "targets/delegation", Gun: gun}, delgKey)
	require.NoError(t, err, "failed to add key to store")

	// Get and validate key info
	delgInfo, err := store.GetKeyInfo(delgKey.ID())
	require.NoError(t, err)
	require.EqualValues(t, "targets/delegation", delgInfo.Role)
	require.EqualValues(t, "", delgInfo.Gun)
}

func TestGetDecryptedWithTamperedCipherText(t *testing.T) {
	testExt := "key"
	testAlias := data.CanonicalRootRole

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	// Create our FileStore
	store, err := NewKeyFileStore(tempBaseDir, passphraseRetriever)
	require.NoError(t, err, "failed to create new key filestore")

	// Generate a new Private Key
	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, "could not generate private key")

	// Call the AddEncryptedKey function
	err = store.AddKey(KeyInfo{Role: testAlias, Gun: ""}, privKey)
	require.NoError(t, err, "failed to add key to store")

	// Since we're generating this manually we need to add the extension '.'
	expectedFilePath := filepath.Join(tempBaseDir, notary.PrivDir, privKey.ID()+"."+testExt)

	// Get file description, open file
	fp, err := os.OpenFile(expectedFilePath, os.O_WRONLY, 0600)
	require.NoError(t, err, "expected file not found")

	// Tamper the file
	fp.WriteAt([]byte("a"), int64(1))

	// Recreate the KeyFileStore to avoid caching
	store, err = NewKeyFileStore(tempBaseDir, passphraseRetriever)
	require.NoError(t, err, "failed to create new key filestore")

	// Try to decrypt the file
	_, _, err = store.GetKey(privKey.ID())
	require.Error(t, err, "expected error while decrypting the content due to invalid cipher text")
}

func TestGetDecryptedWithInvalidPassphrase(t *testing.T) {

	// Make a passphraseRetriever that always returns a different passphrase in order to test
	// decryption failure
	a := "a"
	var invalidPassphraseRetriever = func(keyId string, alias string, createNew bool, numAttempts int) (string, bool, error) {
		if numAttempts > 5 {
			giveup := true
			return "", giveup, nil
		}
		a = a + a
		return a, false, nil
	}

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	// Test with KeyFileStore
	fileStore, err := NewKeyFileStore(tempBaseDir, invalidPassphraseRetriever)
	require.NoError(t, err, "failed to create new key filestore")

	newFileStore, err := NewKeyFileStore(tempBaseDir, invalidPassphraseRetriever)
	require.NoError(t, err, "failed to create new key filestore")

	testGetDecryptedWithInvalidPassphrase(t, fileStore, newFileStore, ErrPasswordInvalid{})

	// Can't test with KeyMemoryStore because we cache the decrypted version of
	// the key forever
}

func TestGetDecryptedWithConsistentlyInvalidPassphrase(t *testing.T) {
	// Make a passphraseRetriever that always returns a different passphrase in order to test
	// decryption failure
	a := "aaaaaaaaaaaaa"
	var consistentlyInvalidPassphraseRetriever = func(keyID string, alias string, createNew bool, numAttempts int) (string, bool, error) {
		a = a + "a"
		return a, false, nil
	}

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	// Test with KeyFileStore
	fileStore, err := NewKeyFileStore(tempBaseDir, consistentlyInvalidPassphraseRetriever)
	require.NoError(t, err, "failed to create new key filestore")

	newFileStore, err := NewKeyFileStore(tempBaseDir, consistentlyInvalidPassphraseRetriever)
	require.NoError(t, err, "failed to create new key filestore")

	testGetDecryptedWithInvalidPassphrase(t, fileStore, newFileStore, ErrAttemptsExceeded{})

	// Can't test with KeyMemoryStore because we cache the decrypted version of
	// the key forever
}

// testGetDecryptedWithInvalidPassphrase takes two keystores so it can add to
// one and get from the other (to work around caching)
func testGetDecryptedWithInvalidPassphrase(t *testing.T, store KeyStore, newStore KeyStore, expectedFailureType interface{}) {
	testAlias := data.CanonicalRootRole

	// Generate a new random RSA Key
	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, "could not generate private key")

	// Call the AddKey function
	err = store.AddKey(KeyInfo{Role: testAlias, Gun: data.GUN("")}, privKey)
	require.NoError(t, err, "failed to add key to store")

	// Try to decrypt the file with an invalid passphrase
	_, _, err = newStore.GetKey(privKey.ID())
	require.Error(t, err, "expected error while decrypting the content due to invalid passphrase")
	require.IsType(t, err, expectedFailureType)
}

func TestRemoveKey(t *testing.T) {
	testRemoveKeyWithRole(t, data.CanonicalRootRole)
	testRemoveKeyWithRole(t, data.CanonicalTargetsRole)
	testRemoveKeyWithRole(t, data.CanonicalSnapshotRole)
	testRemoveKeyWithRole(t, "targets/a/b/c")
	testRemoveKeyWithRole(t, "invalidRole")
}

func testRemoveKeyWithRole(t *testing.T, role data.RoleName) {
	var gun data.GUN = "docker.com/notary"
	testExt := "key"

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	// Create our store
	store, err := NewKeyFileStore(tempBaseDir, passphraseRetriever)
	require.NoError(t, err, "failed to create new key filestore")

	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, "could not generate private key")

	// Since we're generating this manually we need to add the extension '.'
	expectedFilePath := filepath.Join(tempBaseDir, notary.PrivDir, privKey.ID()+"."+testExt)

	err = store.AddKey(KeyInfo{Role: role, Gun: gun}, privKey)
	require.NoError(t, err, "failed to add key to store")

	// Check to see if file exists
	_, err = ioutil.ReadFile(expectedFilePath)
	require.NoError(t, err, "expected file not found")

	// Call remove key
	err = store.RemoveKey(privKey.ID())
	require.NoError(t, err, "unable to remove key")

	// Check to see if file still exists
	_, err = ioutil.ReadFile(expectedFilePath)
	require.Error(t, err, "file should not exist")
}

func TestKeysAreCached(t *testing.T) {
	var (
		gun       data.GUN      = "docker.com/notary"
		testAlias data.RoleName = "alias"
	)

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	var countingPassphraseRetriever notary.PassRetriever

	numTimesCalled := 0
	countingPassphraseRetriever = func(keyId, alias string, createNew bool, attempts int) (passphrase string, giveup bool, err error) {
		numTimesCalled++
		return "password", false, nil
	}

	// Create our store
	store, err := NewKeyFileStore(tempBaseDir, countingPassphraseRetriever)
	require.NoError(t, err, "failed to create new key filestore")

	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, "could not generate private key")

	// Call the AddKey function
	err = store.AddKey(KeyInfo{Role: testAlias, Gun: gun}, privKey)
	require.NoError(t, err, "failed to add key to store")

	require.Equal(t, 1, numTimesCalled, "numTimesCalled should have been 1")

	// Call the AddKey function
	privKey2, _, err := store.GetKey(privKey.ID())
	require.NoError(t, err, "failed to add key to store")

	require.Equal(t, privKey.Public(), privKey2.Public(), "cachedPrivKey should be the same as the added privKey")
	require.Equal(t, privKey.Private(), privKey2.Private(), "cachedPrivKey should be the same as the added privKey")
	require.Equal(t, 1, numTimesCalled, "numTimesCalled should be 1 -- no additional call to passphraseRetriever")

	// Create a new store
	store2, err := NewKeyFileStore(tempBaseDir, countingPassphraseRetriever)
	require.NoError(t, err, "failed to create new key filestore")

	// Call the GetKey function
	privKey3, _, err := store2.GetKey(privKey.ID())
	require.NoError(t, err, "failed to get key from store")

	require.Equal(t, privKey2.Private(), privKey3.Private(), "privkey from store1 should be the same as privkey from store2")
	require.Equal(t, privKey2.Public(), privKey3.Public(), "privkey from store1 should be the same as privkey from store2")
	require.Equal(t, 2, numTimesCalled, "numTimesCalled should be 2 -- one additional call to passphraseRetriever")

	// Call the GetKey function a bunch of times
	for i := 0; i < 10; i++ {
		_, _, err := store2.GetKey(privKey.ID())
		require.NoError(t, err, "failed to get key from store")
	}
	require.Equal(t, 2, numTimesCalled, "numTimesCalled should be 2 -- no additional call to passphraseRetriever")
}
