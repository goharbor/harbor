package testutils

import (
	store "github.com/docker/notary/storage"
	"github.com/docker/notary/tuf/data"
)

// CorruptingMemoryStore corrupts all data returned by GetMeta
type CorruptingMemoryStore struct {
	store.MemoryStore
}

// NewCorruptingMemoryStore returns a new instance of memory store that
// corrupts all data requested from it.
func NewCorruptingMemoryStore(meta map[data.RoleName][]byte) *CorruptingMemoryStore {
	s := store.NewMemoryStore(meta)
	return &CorruptingMemoryStore{MemoryStore: *s}
}

// GetSized returns up to size bytes of meta identified by string. It will
// always be corrupted by setting the first character to }
func (cm CorruptingMemoryStore) GetSized(name string, size int64) ([]byte, error) {
	d, err := cm.MemoryStore.GetSized(name, size)
	if err != nil {
		return nil, err
	}
	d[0] = '}' // all our content is JSON so must start with {
	return d, err
}

// LongMemoryStore corrupts all data returned by GetMeta
type LongMemoryStore struct {
	store.MemoryStore
}

// NewLongMemoryStore returns a new instance of memory store that
// returns one byte too much data on any request to GetMeta
func NewLongMemoryStore(meta map[data.RoleName][]byte) *LongMemoryStore {
	s := store.NewMemoryStore(meta)
	return &LongMemoryStore{MemoryStore: *s}
}

// GetSized returns one byte too much
func (lm LongMemoryStore) GetSized(name string, size int64) ([]byte, error) {
	d, err := lm.MemoryStore.GetSized(name, size)
	if err != nil {
		return nil, err
	}
	d = append(d, ' ')
	return d, err
}

// ShortMemoryStore corrupts all data returned by GetMeta
type ShortMemoryStore struct {
	store.MemoryStore
}

// NewShortMemoryStore returns a new instance of memory store that
// returns one byte too little data on any request to GetMeta
func NewShortMemoryStore(meta map[data.RoleName][]byte) *ShortMemoryStore {
	s := store.NewMemoryStore(meta)
	return &ShortMemoryStore{MemoryStore: *s}
}

// GetSized returns one byte too few
func (sm ShortMemoryStore) GetSized(name string, size int64) ([]byte, error) {
	d, err := sm.MemoryStore.GetSized(name, size)
	if err != nil {
		return nil, err
	}
	return d[1:], err
}
