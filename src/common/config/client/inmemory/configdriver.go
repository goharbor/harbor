package inmemory

import "github.com/goharbor/harbor/src/common/config"

// ConfigInMemory - used in testing only
type ConfigInMemory struct {
	config.ConfigureStore
}

// NewConfigInMemory ...
func NewConfigInMemory() *ConfigInMemory {
	return NewConfigInMemoryFromArray(config.ConfigList)
}

// NewConfigInMemoryFromArray ...
func NewConfigInMemoryFromArray(items []config.Item) *ConfigInMemory {
	cim := &ConfigInMemory{*config.NewConfigureStore()}
	config.MetaData.InitMetaDataFromArray(items)
	cim.InitFromArray(items)
	return cim
}
