package db

import (
	"sync"

	"github.com/goharbor/harbor/src/common/config/encrypt"

	"github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// ConfigureDriver - Retrieve configurations from database
type ConfigureDriver struct {
	config.ConfigureStore
}

// NewDBConfigureStore ...
func NewDBConfigureStore() *ConfigureDriver {
	return NewDBConfigureStoreFromArray(config.ConfigList)
}

// NewDBConfigureStoreFromArray ...
func NewDBConfigureStoreFromArray(items []config.Item) *ConfigureDriver {
	cd := &ConfigureDriver{*config.NewConfigureStore()}
	config.MetaData.InitMetaDataFromArray(items)
	cd.InitFromArray(items)
	return cd
}

var instance *ConfigureDriver
var once sync.Once

// GetConfigureDriverInstance - get instance of DB ConfigureDriver
func GetConfigureDriverInstance() *ConfigureDriver {
	once.Do(func() {
		instance = &ConfigureDriver{*config.NewConfigureStore()}
		config.MetaData.InitMetaDataFromArray(config.ConfigList)
		instance.InitFromArray(config.ConfigList)
	})
	return instance
}

// Load ...
// FIXME: refactor
func (cd *ConfigureDriver) Load() error {
	cfgs := map[string]string{}
	configEntries, err := dao.GetConfigEntries()
	if err != nil {
		return err
	}
	for _, item := range configEntries {
		// ignore item can be relead from env
		itemMetadata, err := config.MetaData.GetConfigMetaData(item.Key)
		if err != nil {
			log.Errorf("failed to GetConfigMetaData, key:%v, error:%v", item.Key, err)
			continue
		}
		if itemMetadata.Reloadable {
			continue
		}
		if itemMetadata.Type == config.PasswordType {
			if decryptPassword, err := encrypt.GetInstance().Decrypt(item.Value); err == nil {
				cfgs[item.Key] = decryptPassword
			} else {
				log.Errorf("Encrypt password failed, error %v", err)
			}
		} else {
			cfgs[item.Key] = item.Value
		}

	}
	cd.LoadFromMap(cfgs)
	return nil
}

// Save ...
func (cd *ConfigureDriver) Save() error {
	var configEntries []models.ConfigEntry
	configValues, err := cd.GetAllSettings()
	if err != nil {
		return err
	}
	for _, v := range configValues {
		var entry = new(models.ConfigEntry)
		entry.Key = v.GetKey()
		itemMetadata, err := config.MetaData.GetConfigMetaData(v.GetKey())
		if err != nil {
			log.Errorf("failed to GetConfigMetaData, key:%v, error:%v", v.GetKey(), err)
			continue
		}
		if itemMetadata.Type == config.PasswordType {
			if encryptPassword, err := encrypt.GetInstance().Encrypt(v.GetString()); err == nil {
				entry.Value = encryptPassword
			} else {
				log.Errorf("Encrypt password failed, error %v", err)
			}
		} else {
			entry.Value = v.GetString()
		}
		configEntries = append(configEntries, *entry)
	}
	return dao.SaveConfigEntries(configEntries)
}
