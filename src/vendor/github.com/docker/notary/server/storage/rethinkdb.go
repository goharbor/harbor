package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/notary/storage/rethinkdb"
	"github.com/docker/notary/tuf/data"
	"gopkg.in/dancannon/gorethink.v3"
)

// RDBTUFFile is a TUF file record
type RDBTUFFile struct {
	rethinkdb.Timing
	GunRoleVersion []interface{} `gorethink:"gun_role_version"`
	Gun            string        `gorethink:"gun"`
	Role           string        `gorethink:"role"`
	Version        int           `gorethink:"version"`
	SHA256         string        `gorethink:"sha256"`
	Data           []byte        `gorethink:"data"`
	TSchecksum     string        `gorethink:"timestamp_checksum"`
}

// TableName returns the table name for the record type
func (r RDBTUFFile) TableName() string {
	return "tuf_files"
}

// gorethink can't handle an UnmarshalJSON function (see https://github.com/gorethink/gorethink/issues/201),
// so do this here in an anonymous struct
func rdbTUFFileFromJSON(data []byte) (interface{}, error) {
	a := struct {
		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
		DeletedAt  time.Time `json:"deleted_at"`
		Gun        string    `json:"gun"`
		Role       string    `json:"role"`
		Version    int       `json:"version"`
		SHA256     string    `json:"sha256"`
		Data       []byte    `json:"data"`
		TSchecksum string    `json:"timestamp_checksum"`
	}{}
	if err := json.Unmarshal(data, &a); err != nil {
		return RDBTUFFile{}, err
	}
	return RDBTUFFile{
		Timing: rethinkdb.Timing{
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
			DeletedAt: a.DeletedAt,
		},
		GunRoleVersion: []interface{}{a.Gun, a.Role, a.Version},
		Gun:            a.Gun,
		Role:           a.Role,
		Version:        a.Version,
		SHA256:         a.SHA256,
		Data:           a.Data,
		TSchecksum:     a.TSchecksum,
	}, nil
}

// RethinkDB implements a MetaStore against the Rethink Database
type RethinkDB struct {
	dbName   string
	sess     *gorethink.Session
	user     string
	password string
}

// NewRethinkDBStorage initializes a RethinkDB object
func NewRethinkDBStorage(dbName, user, password string, sess *gorethink.Session) RethinkDB {
	return RethinkDB{
		dbName:   dbName,
		sess:     sess,
		user:     user,
		password: password,
	}
}

// UpdateCurrent adds new metadata version for the given GUN if and only
// if it's a new role, or the version is greater than the current version
// for the role. Otherwise an error is returned.
func (rdb RethinkDB) UpdateCurrent(gun data.GUN, update MetaUpdate) error {
	now := time.Now()
	checksum := sha256.Sum256(update.Data)
	file := RDBTUFFile{
		Timing: rethinkdb.Timing{
			CreatedAt: now,
			UpdatedAt: now,
		},
		GunRoleVersion: []interface{}{gun, update.Role, update.Version},
		Gun:            gun.String(),
		Role:           update.Role.String(),
		Version:        update.Version,
		SHA256:         hex.EncodeToString(checksum[:]),
		Data:           update.Data,
	}
	_, err := gorethink.DB(rdb.dbName).Table(file.TableName()).Insert(
		file,
		gorethink.InsertOpts{
			Conflict: "error", // default but explicit for clarity of intent
		},
	).RunWrite(rdb.sess)
	if err != nil && gorethink.IsConflictErr(err) {
		return ErrOldVersion{}
	}
	return err
}

// UpdateCurrentWithTSChecksum adds new metadata version for the given GUN with an associated
// checksum for the timestamp it belongs to, to afford us transaction-like functionality
func (rdb RethinkDB) UpdateCurrentWithTSChecksum(gun, tsChecksum string, update MetaUpdate) error {
	now := time.Now()
	checksum := sha256.Sum256(update.Data)
	file := RDBTUFFile{
		Timing: rethinkdb.Timing{
			CreatedAt: now,
			UpdatedAt: now,
		},
		GunRoleVersion: []interface{}{gun, update.Role, update.Version},
		Gun:            gun,
		Role:           update.Role.String(),
		Version:        update.Version,
		SHA256:         hex.EncodeToString(checksum[:]),
		TSchecksum:     tsChecksum,
		Data:           update.Data,
	}
	_, err := gorethink.DB(rdb.dbName).Table(file.TableName()).Insert(
		file,
		gorethink.InsertOpts{
			Conflict: "error", // default but explicit for clarity of intent
		},
	).RunWrite(rdb.sess)
	if err != nil && gorethink.IsConflictErr(err) {
		return ErrOldVersion{}
	}
	return err
}

// Used for sorting updates alphabetically by role name, such that timestamp is always last:
// Ordering: root, snapshot, targets, targets/* (delegations), timestamp
type updateSorter []MetaUpdate

func (u updateSorter) Len() int      { return len(u) }
func (u updateSorter) Swap(i, j int) { u[i], u[j] = u[j], u[i] }
func (u updateSorter) Less(i, j int) bool {
	return u[i].Role < u[j].Role
}

// UpdateMany adds multiple new metadata for the given GUN. RethinkDB does
// not support transactions, therefore we will attempt to insert the timestamp
// last as this represents a published version of the repo.  However, we will
// insert all other role data in alphabetical order first, and also include the
// associated timestamp checksum so that we can easily roll back this pseudotransaction
func (rdb RethinkDB) UpdateMany(gun data.GUN, updates []MetaUpdate) error {
	// find the timestamp first and save its checksum
	// then apply the updates in alphabetic role order with the timestamp last
	// if there are any failures, we roll back in the same alphabetic order
	var tsChecksum string
	for _, up := range updates {
		if up.Role == data.CanonicalTimestampRole {
			tsChecksumBytes := sha256.Sum256(up.Data)
			tsChecksum = hex.EncodeToString(tsChecksumBytes[:])
			break
		}
	}

	// alphabetize the updates by Role name
	sort.Stable(updateSorter(updates))

	for _, up := range updates {
		if err := rdb.UpdateCurrentWithTSChecksum(gun.String(), tsChecksum, up); err != nil {
			// roll back with best-effort deletion, and then error out
			rollbackErr := rdb.deleteByTSChecksum(tsChecksum)
			if rollbackErr != nil {
				logrus.Errorf("Unable to rollback DB conflict - items with timestamp_checksum %s: %v",
					tsChecksum, rollbackErr)
			}
			return err
		}
	}
	return nil
}

// GetCurrent returns the modification date and data part of the metadata for
// the latest version of the given GUN and role.  If there is no data for
// the given GUN and role, an error is returned.
func (rdb RethinkDB) GetCurrent(gun data.GUN, role data.RoleName) (created *time.Time, data []byte, err error) {
	file := RDBTUFFile{}
	res, err := gorethink.DB(rdb.dbName).Table(file.TableName(), gorethink.TableOpts{ReadMode: "majority"}).GetAllByIndex(
		rdbGunRoleIdx, []string{gun.String(), role.String()},
	).OrderBy(gorethink.Desc("version")).Run(rdb.sess)
	if err != nil {
		return nil, nil, err
	}
	defer res.Close()
	if res.IsNil() {
		return nil, nil, ErrNotFound{}
	}
	err = res.One(&file)
	if err == gorethink.ErrEmptyResult {
		return nil, nil, ErrNotFound{}
	}
	return &file.CreatedAt, file.Data, err
}

// GetChecksum returns the given TUF role file and creation date for the
// GUN with the provided checksum. If the given (gun, role, checksum) are
// not found, it returns storage.ErrNotFound
func (rdb RethinkDB) GetChecksum(gun data.GUN, role data.RoleName, checksum string) (created *time.Time, data []byte, err error) {
	var file RDBTUFFile
	res, err := gorethink.DB(rdb.dbName).Table(file.TableName(), gorethink.TableOpts{ReadMode: "majority"}).GetAllByIndex(
		rdbGunRoleSHA256Idx, []string{gun.String(), role.String(), checksum},
	).Run(rdb.sess)
	if err != nil {
		return nil, nil, err
	}
	defer res.Close()
	if res.IsNil() {
		return nil, nil, ErrNotFound{}
	}
	err = res.One(&file)
	if err == gorethink.ErrEmptyResult {
		return nil, nil, ErrNotFound{}
	}
	return &file.CreatedAt, file.Data, err
}

// GetVersion gets a specific TUF record by its version
func (rdb RethinkDB) GetVersion(gun data.GUN, role data.RoleName, version int) (*time.Time, []byte, error) {
	var file RDBTUFFile
	res, err := gorethink.DB(rdb.dbName).Table(file.TableName(), gorethink.TableOpts{ReadMode: "majority"}).Get([]interface{}{gun.String(), role.String(), version}).Run(rdb.sess)
	if err != nil {
		return nil, nil, err
	}
	defer res.Close()
	if res.IsNil() {
		return nil, nil, ErrNotFound{}
	}
	err = res.One(&file)
	if err == gorethink.ErrEmptyResult {
		return nil, nil, ErrNotFound{}
	}
	return &file.CreatedAt, file.Data, err
}

// Delete removes all metadata for a given GUN.  It does not return an
// error if no metadata exists for the given GUN.
func (rdb RethinkDB) Delete(gun data.GUN) error {
	_, err := gorethink.DB(rdb.dbName).Table(RDBTUFFile{}.TableName()).GetAllByIndex(
		"gun", gun.String(),
	).Delete().RunWrite(rdb.sess)
	if err != nil {
		return fmt.Errorf("unable to delete %s from database: %s", gun.String(), err.Error())
	}
	return nil
}

// deleteByTSChecksum removes all metadata by a timestamp checksum, used for rolling back a "transaction"
// from a call to rethinkdb's UpdateMany
func (rdb RethinkDB) deleteByTSChecksum(tsChecksum string) error {
	_, err := gorethink.DB(rdb.dbName).Table(RDBTUFFile{}.TableName()).GetAllByIndex(
		"timestamp_checksum", tsChecksum,
	).Delete().RunWrite(rdb.sess)
	if err != nil {
		return fmt.Errorf("unable to delete timestamp checksum data: %s from database: %s", tsChecksum, err.Error())
	}
	return nil
}

// Bootstrap sets up the database and tables, also creating the notary server user with appropriate db permission
func (rdb RethinkDB) Bootstrap() error {
	if err := rethinkdb.SetupDB(rdb.sess, rdb.dbName, []rethinkdb.Table{
		TUFFilesRethinkTable,
	}); err != nil {
		return err
	}
	return rethinkdb.CreateAndGrantDBUser(rdb.sess, rdb.dbName, rdb.user, rdb.password)
}

// CheckHealth checks that all tables and databases exist and are query-able
func (rdb RethinkDB) CheckHealth() error {
	res, err := gorethink.DB(rdb.dbName).Table(TUFFilesRethinkTable.Name).Info().Run(rdb.sess)
	if err != nil {
		return fmt.Errorf("%s is unavailable, or missing one or more tables, or permissions are incorrectly set", rdb.dbName)
	}
	defer res.Close()
	return nil
}

// GetChanges is not implemented for RethinkDB
func (rdb RethinkDB) GetChanges(changeID string, pageSize int, filterName string) ([]Change, error) {
	return nil, errors.New("Not Implemented")
}
