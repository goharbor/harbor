// Package driver provide the implementation of config driver used in CfgManager
package driver

// Driver the interface to save/load config
type Driver interface {
	// Load - load config item from config driver
	Load() (map[string]interface{}, error)
	// Save - save config item into config driver
	Save(cfg map[string]interface{}) error
}
