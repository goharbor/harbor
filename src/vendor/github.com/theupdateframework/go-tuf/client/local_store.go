package client

import (
	"encoding/json"
)

func MemoryLocalStore() LocalStore {
	return make(memoryLocalStore)
}

type memoryLocalStore map[string]json.RawMessage

func (m memoryLocalStore) GetMeta() (map[string]json.RawMessage, error) {
	return m, nil
}

func (m memoryLocalStore) SetMeta(name string, meta json.RawMessage) error {
	m[name] = meta
	return nil
}

func (m memoryLocalStore) DeleteMeta(name string) error {
	delete(m, name)
	return nil
}

func (m memoryLocalStore) Close() error {
	return nil
}
