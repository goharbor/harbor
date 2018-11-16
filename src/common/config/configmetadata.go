package config

import (
	"sync"
)

// Constant for configure item
const (
	// Scope
	UserScope   = "user"
	SystemScope = "system"
	// Group
	LdapBasicGroup = "ldapbasic"
	LdapGroupGroup = "ldapgroup"
	EmailGroup     = "email"
	UAAGroup       = "uaa"
	DatabaseGroup  = "database"
	BasicGroup     = "basic"
	ClairGroup     = "clair"

	// Type
	IntType      = "int"
	Int64Type    = "int64"
	StringType   = "string"
	BoolType     = "bool"
	PasswordType = "password"
	MapType      = "map"
)

// ValidateFunc - function to validate configure items
type ValidateFunc func(key, value string) error

// ConfigureMetaData ...
type ConfigureMetaData struct {
	sync.RWMutex
	metaMap map[string]Item
}

// MetaData ...
var MetaData = NewConfigureMetaData()

// NewConfigureMetaData ...
func NewConfigureMetaData() *ConfigureMetaData {
	cm := new(ConfigureMetaData)
	cm.metaMap = make(map[string]Item)
	return cm
}

func (cm *ConfigureMetaData) readMap(key string) (Item, error) {
	cm.RLock()
	defer cm.RUnlock()
	if item, ok := cm.metaMap[key]; ok {
		return item, nil
	}
	return Item{}, ErrNotDefined

}
func (cm *ConfigureMetaData) writeMap(key string, item Item) {
	cm.Lock()
	defer cm.Unlock()
	cm.metaMap[key] = item
}

// InitMetaData ...
func (cm *ConfigureMetaData) InitMetaData() {
	for _, item := range ConfigList {
		cm.writeMap(item.Name, item)
	}
}

// InitMetaDataFromArray - used for testing
func (cm *ConfigureMetaData) InitMetaDataFromArray(items []Item) {
	for _, item := range items {
		cm.writeMap(item.Name, item)
	}
}

// GetAllConfigureItems - Get All Configuration Items
func (cm *ConfigureMetaData) GetAllConfigureItems() (items []Item) {
	cm.RLock()
	defer cm.RUnlock()
	result := make([]Item, 0)
	for _, item := range cm.metaMap {
		result = append(result, item)
	}
	return result
}

// GetConfigMetaData - Get single configuration item
func (cm *ConfigureMetaData) GetConfigMetaData(key string) (Item, error) {
	return cm.readMap(key)
}
