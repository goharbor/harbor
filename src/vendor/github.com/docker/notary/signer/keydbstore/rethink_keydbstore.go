package keydbstore

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/docker/notary"
	"github.com/docker/notary/storage/rethinkdb"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	jose "github.com/dvsekhvalnov/jose2go"
	"gopkg.in/dancannon/gorethink.v3"
)

// RethinkDBKeyStore persists and manages private keys on a RethinkDB database
type RethinkDBKeyStore struct {
	sess             *gorethink.Session
	dbName           string
	defaultPassAlias string
	retriever        notary.PassRetriever
	user             string
	password         string
	nowFunc          func() time.Time
}

// RDBPrivateKey represents a PrivateKey in the rethink database
type RDBPrivateKey struct {
	rethinkdb.Timing
	KeyID           string        `gorethink:"key_id"`
	EncryptionAlg   string        `gorethink:"encryption_alg"`
	KeywrapAlg      string        `gorethink:"keywrap_alg"`
	Algorithm       string        `gorethink:"algorithm"`
	PassphraseAlias string        `gorethink:"passphrase_alias"`
	Gun             data.GUN      `gorethink:"gun"`
	Role            data.RoleName `gorethink:"role"`

	// gorethink specifically supports binary types, and says to pass it in as
	// a byteslice.  Currently our encryption method for the private key bytes
	// produces a base64-encoded string, but for future compatibility in case
	// we change how we encrypt, use a byteslace for the encrypted private key
	// too
	Public  []byte `gorethink:"public"`
	Private []byte `gorethink:"private"`

	// whether this key is active or not
	LastUsed time.Time `gorethink:"last_used"`
}

// gorethink can't handle an UnmarshalJSON function (see https://github.com/gorethink/gorethink/issues/201),
// so do this here in an anonymous struct
func rdbPrivateKeyFromJSON(jsonData []byte) (interface{}, error) {
	a := struct {
		CreatedAt       time.Time     `json:"created_at"`
		UpdatedAt       time.Time     `json:"updated_at"`
		DeletedAt       time.Time     `json:"deleted_at"`
		KeyID           string        `json:"key_id"`
		EncryptionAlg   string        `json:"encryption_alg"`
		KeywrapAlg      string        `json:"keywrap_alg"`
		Algorithm       string        `json:"algorithm"`
		PassphraseAlias string        `json:"passphrase_alias"`
		Gun             data.GUN      `json:"gun"`
		Role            data.RoleName `json:"role"`
		Public          []byte        `json:"public"`
		Private         []byte        `json:"private"`
		LastUsed        time.Time     `json:"last_used"`
	}{}
	if err := json.Unmarshal(jsonData, &a); err != nil {
		return RDBPrivateKey{}, err
	}
	return RDBPrivateKey{
		Timing: rethinkdb.Timing{
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
			DeletedAt: a.DeletedAt,
		},
		KeyID:           a.KeyID,
		EncryptionAlg:   a.EncryptionAlg,
		KeywrapAlg:      a.KeywrapAlg,
		Algorithm:       a.Algorithm,
		PassphraseAlias: a.PassphraseAlias,
		Gun:             a.Gun,
		Role:            a.Role,
		Public:          a.Public,
		Private:         a.Private,
		LastUsed:        a.LastUsed,
	}, nil

}

// PrivateKeysRethinkTable is the table definition for notary signer's key information
var PrivateKeysRethinkTable = rethinkdb.Table{
	Name:             RDBPrivateKey{}.TableName(),
	PrimaryKey:       "key_id",
	JSONUnmarshaller: rdbPrivateKeyFromJSON,
}

// TableName sets a specific table name for our RDBPrivateKey
func (g RDBPrivateKey) TableName() string {
	return "private_keys"
}

// NewRethinkDBKeyStore returns a new RethinkDBKeyStore backed by a RethinkDB database
func NewRethinkDBKeyStore(dbName, username, password string, passphraseRetriever notary.PassRetriever, defaultPassAlias string, rethinkSession *gorethink.Session) *RethinkDBKeyStore {
	return &RethinkDBKeyStore{
		sess:             rethinkSession,
		defaultPassAlias: defaultPassAlias,
		dbName:           dbName,
		retriever:        passphraseRetriever,
		user:             username,
		password:         password,
		nowFunc:          time.Now,
	}
}

// Name returns a user friendly name for the storage location
func (rdb *RethinkDBKeyStore) Name() string {
	return "RethinkDB"
}

// AddKey stores the contents of a private key. Both role and gun are ignored,
// we always use Key IDs as name, and don't support aliases
func (rdb *RethinkDBKeyStore) AddKey(role data.RoleName, gun data.GUN, privKey data.PrivateKey) error {
	passphrase, _, err := rdb.retriever(privKey.ID(), rdb.defaultPassAlias, false, 1)
	if err != nil {
		return err
	}

	encryptedKey, err := jose.Encrypt(string(privKey.Private()), KeywrapAlg, EncryptionAlg, passphrase)
	if err != nil {
		return err
	}

	now := rdb.nowFunc()
	rethinkPrivKey := RDBPrivateKey{
		Timing: rethinkdb.Timing{
			CreatedAt: now,
			UpdatedAt: now,
		},
		KeyID:           privKey.ID(),
		EncryptionAlg:   EncryptionAlg,
		KeywrapAlg:      KeywrapAlg,
		PassphraseAlias: rdb.defaultPassAlias,
		Algorithm:       privKey.Algorithm(),
		Gun:             gun,
		Role:            role,
		Public:          privKey.Public(),
		Private:         []byte(encryptedKey),
	}

	// Add encrypted private key to the database
	_, err = gorethink.DB(rdb.dbName).Table(rethinkPrivKey.TableName()).Insert(rethinkPrivKey).RunWrite(rdb.sess)
	if err != nil {
		return fmt.Errorf("failed to add private key %s to database: %s", privKey.ID(), err.Error())
	}

	return nil
}

// getKeyBytes returns the RDBPrivateKey given a KeyID, as well as the decrypted private bytes
func (rdb *RethinkDBKeyStore) getKey(keyID string) (*RDBPrivateKey, string, error) {
	// Retrieve the RethinkDB private key from the database
	dbPrivateKey := RDBPrivateKey{}
	res, err := gorethink.DB(rdb.dbName).Table(dbPrivateKey.TableName()).Filter(gorethink.Row.Field("key_id").Eq(keyID)).Run(rdb.sess)
	if err != nil {
		return nil, "", err
	}
	defer res.Close()

	err = res.One(&dbPrivateKey)
	if err != nil {
		return nil, "", trustmanager.ErrKeyNotFound{}
	}

	// Get the passphrase to use for this key
	passphrase, _, err := rdb.retriever(dbPrivateKey.KeyID, dbPrivateKey.PassphraseAlias, false, 1)
	if err != nil {
		return nil, "", err
	}

	// Decrypt private bytes from the gorm key
	decryptedPrivKey, _, err := jose.Decode(string(dbPrivateKey.Private), passphrase)
	if err != nil {
		return nil, "", err
	}

	return &dbPrivateKey, decryptedPrivKey, nil
}

// GetPrivateKey returns the PrivateKey given a KeyID
func (rdb *RethinkDBKeyStore) GetPrivateKey(keyID string) (data.PrivateKey, data.RoleName, error) {
	dbPrivateKey, decryptedPrivKey, err := rdb.getKey(keyID)
	if err != nil {
		return nil, "", err
	}

	pubKey := data.NewPublicKey(dbPrivateKey.Algorithm, dbPrivateKey.Public)

	// Create a new PrivateKey with unencrypted bytes
	privKey, err := data.NewPrivateKey(pubKey, []byte(decryptedPrivKey))
	if err != nil {
		return nil, "", err
	}

	return activatingPrivateKey{PrivateKey: privKey, activationFunc: rdb.markActive}, dbPrivateKey.Role, nil
}

// GetKey returns the PublicKey given a KeyID, and does not activate the key
func (rdb *RethinkDBKeyStore) GetKey(keyID string) data.PublicKey {
	dbPrivateKey, _, err := rdb.getKey(keyID)
	if err != nil {
		return nil
	}

	return data.NewPublicKey(dbPrivateKey.Algorithm, dbPrivateKey.Public)
}

// ListKeys always returns nil. This method is here to satisfy the CryptoService interface
func (rdb RethinkDBKeyStore) ListKeys(role data.RoleName) []string {
	return nil
}

// ListAllKeys always returns nil. This method is here to satisfy the CryptoService interface
func (rdb RethinkDBKeyStore) ListAllKeys() map[string]data.RoleName {
	return nil
}

// RemoveKey removes the key from the table
func (rdb RethinkDBKeyStore) RemoveKey(keyID string) error {
	// Delete the key from the database
	dbPrivateKey := RDBPrivateKey{KeyID: keyID}
	_, err := gorethink.DB(rdb.dbName).Table(dbPrivateKey.TableName()).Filter(gorethink.Row.Field("key_id").Eq(keyID)).Delete().RunWrite(rdb.sess)
	if err != nil {
		return fmt.Errorf("unable to delete private key %s from database: %s", keyID, err.Error())
	}

	return nil
}

// RotateKeyPassphrase rotates the key-encryption-key
func (rdb RethinkDBKeyStore) RotateKeyPassphrase(keyID, newPassphraseAlias string) error {
	dbPrivateKey, decryptedPrivKey, err := rdb.getKey(keyID)
	if err != nil {
		return err
	}

	// Get the new passphrase to use for this key
	newPassphrase, _, err := rdb.retriever(dbPrivateKey.KeyID, newPassphraseAlias, false, 1)
	if err != nil {
		return err
	}

	// Re-encrypt the private bytes with the new passphrase
	newEncryptedKey, err := jose.Encrypt(decryptedPrivKey, KeywrapAlg, EncryptionAlg, newPassphrase)
	if err != nil {
		return err
	}

	// Update the database object
	dbPrivateKey.Private = []byte(newEncryptedKey)
	dbPrivateKey.PassphraseAlias = newPassphraseAlias
	if _, err := gorethink.DB(rdb.dbName).Table(dbPrivateKey.TableName()).Get(keyID).Update(dbPrivateKey).RunWrite(rdb.sess); err != nil {
		return err
	}

	return nil
}

// markActive marks a particular key as active
func (rdb RethinkDBKeyStore) markActive(keyID string) error {
	_, err := gorethink.DB(rdb.dbName).Table(PrivateKeysRethinkTable.Name).Get(keyID).Update(map[string]interface{}{
		"last_used": rdb.nowFunc(),
	}).RunWrite(rdb.sess)
	return err
}

// Create will attempt to first re-use an inactive key for the same role, gun, and algorithm.
// If one isn't found, it will create a private key and add it to the DB as an inactive key
func (rdb RethinkDBKeyStore) Create(role data.RoleName, gun data.GUN, algorithm string) (data.PublicKey, error) {
	dbPrivateKey := RDBPrivateKey{}
	res, err := gorethink.DB(rdb.dbName).Table(dbPrivateKey.TableName()).
		Filter(gorethink.Row.Field("gun").Eq(gun.String())).
		Filter(gorethink.Row.Field("role").Eq(role.String())).
		Filter(gorethink.Row.Field("algorithm").Eq(algorithm)).
		Filter(gorethink.Row.Field("last_used").Eq(time.Time{})).
		OrderBy(gorethink.Row.Field("key_id")).
		Run(rdb.sess)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	err = res.One(&dbPrivateKey)
	if err == nil {
		return data.NewPublicKey(dbPrivateKey.Algorithm, dbPrivateKey.Public), nil
	}

	privKey, err := generatePrivateKey(algorithm)
	if err != nil {
		return nil, err
	}
	if err = rdb.AddKey(role, gun, privKey); err != nil {
		return nil, fmt.Errorf("failed to store key: %v", err)
	}

	return privKey, nil
}

// Bootstrap sets up the database and tables, also creating the notary signer user with appropriate db permission
func (rdb RethinkDBKeyStore) Bootstrap() error {
	if err := rethinkdb.SetupDB(rdb.sess, rdb.dbName, []rethinkdb.Table{
		PrivateKeysRethinkTable,
	}); err != nil {
		return err
	}
	return rethinkdb.CreateAndGrantDBUser(rdb.sess, rdb.dbName, rdb.user, rdb.password)
}

// CheckHealth verifies that DB exists and is query-able
func (rdb RethinkDBKeyStore) CheckHealth() error {
	res, err := gorethink.DB(rdb.dbName).Table(PrivateKeysRethinkTable.Name).Info().Run(rdb.sess)
	if err != nil {
		return fmt.Errorf("%s is unavailable, or missing one or more tables, or permissions are incorrectly set", rdb.dbName)
	}
	defer res.Close()
	return nil
}
