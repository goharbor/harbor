package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/notary/storage/rethinkdb"
	"github.com/docker/notary/tuf/data"
	"gopkg.in/dancannon/gorethink.v3"
)

// RethinkDB has eventual consistency. This represents a 60 second blackout
// period of the most recent changes in the changefeed which will not be
// returned while the eventual consistency works itself out.
// It's a var not a const so that the tests can turn it down to zero rather
// than have to include a sleep.
var blackoutTime = 60

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
	return TUFFileTableName
}

// Change defines the the fields required for an object in the changefeed
type Change struct {
	ID        string    `gorethink:"id,omitempty" gorm:"primary_key" sql:"not null"`
	CreatedAt time.Time `gorethink:"created_at"`
	GUN       string    `gorethink:"gun" gorm:"column:gun" sql:"type:varchar(255);not null"`
	Version   int       `gorethink:"version" sql:"not null"`
	SHA256    string    `gorethink:"sha256" gorm:"column:sha256" sql:"type:varchar(64);"`
	Category  string    `gorethink:"category" sql:"type:varchar(20);not null;"`
}

// TableName sets a specific table name for Changefeed
func (rdb Change) TableName() string {
	return ChangefeedTableName
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

func rdbChangeFromJSON(data []byte) (interface{}, error) {
	res := Change{}
	if err := json.Unmarshal(data, &res); err != nil {
		return Change{}, err
	}
	return res, nil
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
	// empty string is the zero value for tsChecksum in the RDBTUFFile struct.
	// Therefore we can just call through to updateCurrentWithTSChecksum passing
	// "" for the tsChecksum value.
	if err := rdb.updateCurrentWithTSChecksum(gun.String(), "", update); err != nil {
		return err
	}
	if update.Role == data.CanonicalTimestampRole {
		tsChecksumBytes := sha256.Sum256(update.Data)
		return rdb.writeChange(
			gun.String(),
			update.Version,
			hex.EncodeToString(tsChecksumBytes[:]),
			changeCategoryUpdate,
		)
	}
	return nil
}

// updateCurrentWithTSChecksum adds new metadata version for the given GUN with an associated
// checksum for the timestamp it belongs to, to afford us transaction-like functionality
func (rdb RethinkDB) updateCurrentWithTSChecksum(gun, tsChecksum string, update MetaUpdate) error {
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
	var (
		tsChecksum string
		tsVersion  int
	)
	for _, up := range updates {
		if up.Role == data.CanonicalTimestampRole {
			tsChecksumBytes := sha256.Sum256(up.Data)
			tsChecksum = hex.EncodeToString(tsChecksumBytes[:])
			tsVersion = up.Version
			break
		}
	}

	// alphabetize the updates by Role name
	sort.Stable(updateSorter(updates))

	for _, up := range updates {
		if err := rdb.updateCurrentWithTSChecksum(gun.String(), tsChecksum, up); err != nil {
			// roll back with best-effort deletion, and then error out
			rollbackErr := rdb.deleteByTSChecksum(tsChecksum)
			if rollbackErr != nil {
				logrus.Errorf("Unable to rollback DB conflict - items with timestamp_checksum %s: %v",
					tsChecksum, rollbackErr)
			}
			return err
		}
	}

	// if the update included a timestamp, write a change object
	if tsChecksum != "" {
		return rdb.writeChange(gun.String(), tsVersion, tsChecksum, changeCategoryUpdate)
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
	resp, err := gorethink.DB(rdb.dbName).Table(RDBTUFFile{}.TableName()).GetAllByIndex(
		"gun", gun.String(),
	).Delete().RunWrite(rdb.sess)
	if err != nil {
		return fmt.Errorf("unable to delete %s from database: %s", gun.String(), err.Error())
	}
	if resp.Deleted > 0 {
		return rdb.writeChange(gun.String(), 0, "", changeCategoryDeletion)
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
	// DO NOT WRITE CHANGE! THIS IS USED _ONLY_ TO ROLLBACK A FAILED INSERT
	return nil
}

// Bootstrap sets up the database and tables, also creating the notary server user with appropriate db permission
func (rdb RethinkDB) Bootstrap() error {
	if err := rethinkdb.SetupDB(rdb.sess, rdb.dbName, []rethinkdb.Table{
		TUFFilesRethinkTable,
		ChangeRethinkTable,
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

func (rdb RethinkDB) writeChange(gun string, version int, sha256, category string) error {
	now := time.Now()
	ch := Change{
		CreatedAt: now,
		GUN:       gun,
		Version:   version,
		SHA256:    sha256,
		Category:  category,
	}
	_, err := gorethink.DB(rdb.dbName).Table(ch.TableName()).Insert(
		ch,
		gorethink.InsertOpts{
			Conflict: "error", // default but explicit for clarity of intent
		},
	).RunWrite(rdb.sess)
	return err
}

// GetChanges returns up to pageSize changes starting from changeID. It uses the
// blackout to account for RethinkDB's eventual consistency model
func (rdb RethinkDB) GetChanges(changeID string, pageSize int, filterName string) ([]Change, error) {
	var (
		lower, upper, bound []interface{}
		idx                 = "rdb_created_at_id"
		max                 = []interface{}{gorethink.Now().Sub(blackoutTime), gorethink.MaxVal}
		min                 = []interface{}{gorethink.MinVal, gorethink.MinVal}
		order               gorethink.OrderByOpts
		reversed            bool
	)
	if filterName != "" {
		idx = "rdb_gun_created_at_id"
		max = append([]interface{}{filterName}, max...)
		min = append([]interface{}{filterName}, min...)
	}

	switch changeID {
	case "0", "-1":
		lower = min
		upper = max
	default:
		bound, idx = rdb.bound(changeID, filterName)
		if pageSize < 0 {
			lower = min
			upper = bound
		} else {
			lower = bound
			upper = max
		}
	}

	if changeID == "-1" || pageSize < 0 {
		reversed = true
		order = gorethink.OrderByOpts{Index: gorethink.Desc(idx)}
	} else {
		order = gorethink.OrderByOpts{Index: gorethink.Asc(idx)}
	}

	if pageSize < 0 {
		pageSize = pageSize * -1
	}

	changes := make([]Change, 0, pageSize)

	// Between returns a slice of results from the rethinkdb table.
	// The results are ordered using BetweenOpts.Index, which will
	// default to the index of the immediately preceding OrderBy.
	// The lower and upper are the start and end points for the slice
	// and the Left/RightBound values determine whether the lower and
	// upper values are included in the result per normal set semantics
	// of "open" and "closed"
	res, err := gorethink.DB(rdb.dbName).
		Table(Change{}.TableName(), gorethink.TableOpts{ReadMode: "majority"}).
		OrderBy(order).
		Between(
			lower,
			upper,
			gorethink.BetweenOpts{
				LeftBound:  "open",
				RightBound: "open",
			},
		).Limit(pageSize).Run(rdb.sess)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	defer func() {
		if reversed {
			// results are currently newest to oldest, should be oldest to newest
			for i, j := 0, len(changes)-1; i < j; i, j = i+1, j-1 {
				changes[i], changes[j] = changes[j], changes[i]
			}
		}
	}()

	return changes, res.All(&changes)
}

// bound creates the correct boundary based in the index that should be used for
// querying the changefeed.
func (rdb RethinkDB) bound(changeID, filterName string) ([]interface{}, string) {
	createdAtTerm := gorethink.DB(rdb.dbName).Table(Change{}.TableName()).Get(changeID).Field("created_at")
	if filterName != "" {
		return []interface{}{filterName, createdAtTerm, changeID}, "rdb_gun_created_at_id"
	}
	return []interface{}{createdAtTerm, changeID}, "rdb_created_at_id"
}
