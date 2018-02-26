package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/notary/tuf/data"
)

type key struct {
	algorithm string
	public    []byte
}

type ver struct {
	version      int
	data         []byte
	createupdate time.Time
}

// we want to keep these sorted by version so that it's in increasing version
// order
type verList []ver

func (k verList) Len() int      { return len(k) }
func (k verList) Swap(i, j int) { k[i], k[j] = k[j], k[i] }
func (k verList) Less(i, j int) bool {
	return k[i].version < k[j].version
}

// MemStorage is really just designed for dev and testing. It is very
// inefficient in many scenarios
type MemStorage struct {
	lock      sync.Mutex
	tufMeta   map[string]verList
	keys      map[string]map[string]*key
	checksums map[string]map[string]ver
	changes   []Change
}

// NewMemStorage instantiates a memStorage instance
func NewMemStorage() *MemStorage {
	return &MemStorage{
		tufMeta:   make(map[string]verList),
		keys:      make(map[string]map[string]*key),
		checksums: make(map[string]map[string]ver),
	}
}

// UpdateCurrent updates the meta data for a specific role
func (st *MemStorage) UpdateCurrent(gun data.GUN, update MetaUpdate) error {
	id := entryKey(gun, update.Role)
	st.lock.Lock()
	defer st.lock.Unlock()
	if space, ok := st.tufMeta[id]; ok {
		for _, v := range space {
			if v.version >= update.Version {
				return ErrOldVersion{}
			}
		}
	}
	version := ver{version: update.Version, data: update.Data, createupdate: time.Now()}
	st.tufMeta[id] = append(st.tufMeta[id], version)
	checksumBytes := sha256.Sum256(update.Data)
	checksum := hex.EncodeToString(checksumBytes[:])

	_, ok := st.checksums[gun.String()]
	if !ok {
		st.checksums[gun.String()] = make(map[string]ver)
	}
	st.checksums[gun.String()][checksum] = version
	if update.Role == data.CanonicalTimestampRole {
		st.writeChange(gun, update.Version, checksum)
	}
	return nil
}

// writeChange must only be called by a function already holding a lock on
// the MemStorage. Behaviour is undefined otherwise
func (st *MemStorage) writeChange(gun data.GUN, version int, checksum string) {
	c := Change{
		ID:        uint(len(st.changes) + 1),
		GUN:       gun.String(),
		Version:   version,
		SHA256:    checksum,
		CreatedAt: time.Now(),
		Category:  changeCategoryUpdate,
	}
	st.changes = append(st.changes, c)
}

// UpdateMany updates multiple TUF records
func (st *MemStorage) UpdateMany(gun data.GUN, updates []MetaUpdate) error {
	st.lock.Lock()
	defer st.lock.Unlock()

	versioner := make(map[string]map[int]struct{})
	constant := struct{}{}

	// ensure that we only update in one transaction
	for _, u := range updates {
		id := entryKey(gun, u.Role)

		// prevent duplicate versions of the same role
		if _, ok := versioner[u.Role.String()][u.Version]; ok {
			return ErrOldVersion{}
		}
		if _, ok := versioner[u.Role.String()]; !ok {
			versioner[u.Role.String()] = make(map[int]struct{})
		}
		versioner[u.Role.String()][u.Version] = constant

		if space, ok := st.tufMeta[id]; ok {
			for _, v := range space {
				if v.version >= u.Version {
					return ErrOldVersion{}
				}
			}
		}
	}

	for _, u := range updates {
		id := entryKey(gun, u.Role)

		version := ver{version: u.Version, data: u.Data, createupdate: time.Now()}
		st.tufMeta[id] = append(st.tufMeta[id], version)
		sort.Sort(st.tufMeta[id]) // ensure that it's sorted
		checksumBytes := sha256.Sum256(u.Data)
		checksum := hex.EncodeToString(checksumBytes[:])

		_, ok := st.checksums[gun.String()]
		if !ok {
			st.checksums[gun.String()] = make(map[string]ver)
		}
		st.checksums[gun.String()][checksum] = version
		if u.Role == data.CanonicalTimestampRole {
			st.writeChange(gun, u.Version, checksum)
		}
	}
	return nil
}

// GetCurrent returns the createupdate date metadata for a given role, under a GUN.
func (st *MemStorage) GetCurrent(gun data.GUN, role data.RoleName) (*time.Time, []byte, error) {
	id := entryKey(gun, role)
	st.lock.Lock()
	defer st.lock.Unlock()
	space, ok := st.tufMeta[id]
	if !ok || len(space) == 0 {
		return nil, nil, ErrNotFound{}
	}
	return &(space[len(space)-1].createupdate), space[len(space)-1].data, nil
}

// GetChecksum returns the createupdate date and metadata for a given role, under a GUN.
func (st *MemStorage) GetChecksum(gun data.GUN, role data.RoleName, checksum string) (*time.Time, []byte, error) {
	st.lock.Lock()
	defer st.lock.Unlock()
	space, ok := st.checksums[gun.String()][checksum]
	if !ok || len(space.data) == 0 {
		return nil, nil, ErrNotFound{}
	}
	return &(space.createupdate), space.data, nil
}

// GetVersion gets a specific TUF record by its version
func (st *MemStorage) GetVersion(gun data.GUN, role data.RoleName, version int) (*time.Time, []byte, error) {
	st.lock.Lock()
	defer st.lock.Unlock()

	id := entryKey(gun, role)
	for _, ver := range st.tufMeta[id] {
		if ver.version == version {
			return &(ver.createupdate), ver.data, nil
		}
	}

	return nil, nil, ErrNotFound{}
}

// Delete deletes all the metadata for a given GUN
func (st *MemStorage) Delete(gun data.GUN) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	l := len(st.tufMeta)
	for k := range st.tufMeta {
		if strings.HasPrefix(k, gun.String()) {
			delete(st.tufMeta, k)
		}
	}
	if l == len(st.tufMeta) {
		// we didn't delete anything, don't write change.
		return nil
	}
	delete(st.checksums, gun.String())
	c := Change{
		ID:        uint(len(st.changes) + 1),
		GUN:       gun.String(),
		Category:  changeCategoryDeletion,
		CreatedAt: time.Now(),
	}
	st.changes = append(st.changes, c)
	return nil
}

// GetChanges returns a []Change starting from but excluding the record
// identified by changeID. In the context of the memory store, changeID
// is simply an index into st.changes. The ID of a change is its
// index+1, both to match the SQL implementations, and so that the first
// change can be retrieved by providing ID 0.
func (st *MemStorage) GetChanges(changeID string, records int, filterName string) ([]Change, error) {
	var (
		id  int64
		err error
	)
	if changeID == "" {
		id = 0
	} else {
		id, err = strconv.ParseInt(changeID, 10, 32)
		if err != nil {
			return nil, err
		}
	}
	var (
		start     = int(id)
		toInspect []Change
	)
	if err != nil {
		return nil, err
	}

	reversed := id < 0
	if records < 0 {
		reversed = true
		records = -records
	}

	if len(st.changes) <= int(id) && !reversed {
		// no records to return as we're essentially trying to retrieve
		// changes that haven't happened yet.
		return nil, nil
	}

	// technically only -1 is a valid negative input, but we're going to be
	// broad in what we accept here to reduce the need to error and instead
	// act in a "do what I mean not what I say" fashion. Same logic for
	// requesting changeID < 0 but not asking for reversed, we're just going
	// to force it to be reversed.
	if start < 0 {
		// need to add one so we don't later slice off the last element
		// when calculating toInspect.
		start = len(st.changes) + 1
	}
	// reduce to only look at changes we're interested in
	if reversed {
		if start > len(st.changes) {
			toInspect = st.changes
		} else {
			toInspect = st.changes[:start-1]
		}
	} else {
		toInspect = st.changes[start:]
	}

	// if we're not doing any filtering
	if filterName == "" {
		// if the pageSize is larger than the total records
		// that could be returned, return them all
		if records >= len(toInspect) {
			return toInspect, nil
		}
		// if we're going backwards, return the last pageSize records
		if reversed {
			return toInspect[len(toInspect)-records:], nil
		}
		// otherwise return pageSize records from front
		return toInspect[:records], nil
	}

	return getFilteredChanges(toInspect, filterName, records, reversed), nil
}

func getFilteredChanges(toInspect []Change, filterName string, records int, reversed bool) []Change {
	res := make([]Change, 0, records)
	if reversed {
		for i := len(toInspect) - 1; i >= 0; i-- {
			if toInspect[i].GUN == filterName {
				res = append(res, toInspect[i])
			}
			if len(res) == records {
				break
			}
		}
		// results are currently newest to oldest, should be oldest to newest
		for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
			res[i], res[j] = res[j], res[i]
		}
	} else {
		for _, c := range toInspect {
			if c.GUN == filterName {
				res = append(res, c)
			}
			if len(res) == records {
				break
			}
		}
	}
	return res
}

func entryKey(gun data.GUN, role data.RoleName) string {
	return fmt.Sprintf("%s.%s", gun, role)
}
