package config

import (
	"os"
	"sync"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// ConfigureStore - to manage all configurations
type ConfigureStore struct {
	sync.RWMutex
	// ConfigureValues to store all configure values
	configureValues sync.Map
}

// NewConfigureStore ...
func NewConfigureStore() *ConfigureStore {
	cs := new(ConfigureStore)
	cs.configureValues = sync.Map{}
	return cs
}

func (s *ConfigureStore) readMap(key string) (Value, error) {
	if value, ok := s.configureValues.Load(key); ok {
		if result, suc := value.(Value); suc {
			return result, nil
		}
		return nil, ErrTypeNotMatch
	}
	return nil, ErrValueNotSet

}

func (s *ConfigureStore) writeMap(key string, value Value) {
	s.configureValues.Store(key, value)
}

// StorageInterface ...
type StorageInterface interface {
	// Init - init configurations with default value
	Init() error
	// InitFromString - used for testing
	InitFromString(testingMetaDataArray []Item) error
	// Load from store
	Load() error
	// Save to store
	Save() error
	// LoadFromMap ...
	LoadFromMap(map[string]string)
	// Save all configuration to store
	UpdateAll() error
	// Reset configure to default value
	Reset()
}

// Init - int the store
func (s *ConfigureStore) Init() error {
	MetaData.InitMetaData()
	// Init Default Value
	itemArray := MetaData.GetAllConfigureItems()
	for _, item := range itemArray {
		if len(item.DefaultValue) > 0 {
			c := &ConfigureValue{item.Name, item.DefaultValue}
			err := c.Validate()
			if err == nil {
				s.writeMap(item.Name, c)
			} else {
				log.Errorf("Failed to init config item %+v, default err: %+v", c, err)
			}
		}
	}

	// Init System Value
	for _, item := range itemArray {
		if item.Scope == SystemScope {
			if len(item.EnvironmentKey) > 0 {
				if envValue, ok := os.LookupEnv(item.EnvironmentKey); ok {
					c := &ConfigureValue{item.Name, envValue}
					err := c.Validate()
					if err == nil {
						s.writeMap(item.Name, c)
					} else {
						log.Errorf("Failed to init system config item %+v,  err: %+v", c, err)
					}
				}
			}
		}
	}

	return nil
}

// LoadFromMap ...
func (s *ConfigureStore) LoadFromMap(cfgs map[string]string) {
	for k, v := range cfgs {
		c := &ConfigureValue{k, v}
		err := c.Validate()
		if err != nil {
			log.Errorf("Failed LoadFromMap, config item %+v, err: %+v", c, err)
			continue
		}
		s.writeMap(k, c)
	}
}

// InitFromArray ... Used for testing
func (s *ConfigureStore) InitFromArray(testingMetaDataArray []Item) error {
	MetaData.InitMetaDataFromArray(testingMetaDataArray)
	itemArray := MetaData.GetAllConfigureItems()
	// Init Default Value
	for _, item := range itemArray {
		if len(item.DefaultValue) > 0 {
			c := &ConfigureValue{item.Name, item.DefaultValue}
			err := c.Validate()
			if err == nil {
				s.writeMap(item.Name, c)
			} else {
				log.Errorf("Failed InitFromArray, config item %+v,  err: %+v", c, err)
			}
		}
	}

	// Init System Value
	for _, item := range itemArray {
		if item.Scope == SystemScope {
			if len(item.EnvironmentKey) > 0 {
				if envValue, ok := os.LookupEnv(item.EnvironmentKey); ok {
					c := &ConfigureValue{item.Name, envValue}
					err := c.Validate()
					if err == nil {
						s.writeMap(item.Name, c)
					} else {
						log.Errorf("Failed InitFromArray, config item %+v,  err: %+v", c, err)
					}
				}
			}
		}
	}

	return nil
}

// Load ...
func (s *ConfigureStore) Load() error {
	panic("Load not implemented")
}

// Save ...
func (s *ConfigureStore) Save() error {
	panic("Save not implemented")
}

// UpdateAll ...
func (s *ConfigureStore) UpdateAll() error {
	log.Info("UpdateAll not implemented")
	return nil
}

// Reset ...
func (s *ConfigureStore) Reset() {
	s.Lock()
	defer s.Unlock()
	err := s.Init()
	if err != nil {
		log.Errorf("Error occurred when Init: %v", err)
		return
	}
	err = s.UpdateAll()
	if err != nil {
		log.Errorf("Error occurred when UpdateAll: %v", err)
	}

}

// GetAllSettings ...
func (s *ConfigureStore) GetAllSettings() ([]Value, error) {
	resultValues := make([]Value, 0)
	s.configureValues.Range(func(key, value interface{}) bool {
		if result, suc := value.(Value); suc {
			resultValues = append(resultValues, result)
		}
		return true
	})
	return resultValues, nil
}

// GetSettingByGroup ...
func (s *ConfigureStore) GetSettingByGroup(groupName string) ([]Value, error) {
	resultValues := make([]Value, 0)
	s.configureValues.Range(func(key, value interface{}) bool {
		if keyString, suc := key.(string); suc {
			itemMataData, err := MetaData.GetConfigMetaData(keyString)
			if err == nil {
				if actValue, ok := value.(Value); ok && itemMataData.Group == groupName {
					resultValues = append(resultValues, actValue)
				}
			}
		}
		return true
	})

	return resultValues, nil
}

// GetSettingByScope ...
func (s *ConfigureStore) GetSettingByScope(scope string) ([]Value, error) {
	resultValues := make([]Value, 0)
	s.configureValues.Range(func(key, value interface{}) bool {
		if keyString, suc := key.(string); suc {
			itemMataData, err := MetaData.GetConfigMetaData(keyString)
			if err == nil {
				if actValue, ok := value.(Value); ok && itemMataData.Scope == scope {
					resultValues = append(resultValues, actValue)
				}
			}
		}
		return true
	})
	return resultValues, nil
}

// GetSetting ...
func (s *ConfigureStore) GetSetting(keyName string) (Value, error) {
	_, err := MetaData.GetConfigMetaData(keyName)
	if err == nil {
		return s.readMap(keyName)
	}
	return nil, ErrNotDefined
}

// GetInt ...
func (s *ConfigureStore) GetInt(keyName string) int {
	itemMetadata, err := MetaData.GetConfigMetaData(keyName)
	if err == nil {
		value, err := s.readMap(keyName)
		if err != nil {
			log.Errorf("Error while getting %v, error: %+v", keyName, err)
			return 0
		}
		if itemMetadata.Type == IntType {
			return value.GetInt()
		}
		return 0
	}
	return 0
}

// GetString ...
func (s *ConfigureStore) GetString(keyName string) string {
	_, err := MetaData.GetConfigMetaData(keyName)
	if err == nil {
		value, err := s.readMap(keyName)
		if err != nil {
			log.Errorf("Error while getting %v, error: %+v", keyName, err)
			return ""
		}
		return value.GetString()
	}
	return ""
}

// GetPassword ...
func (s *ConfigureStore) GetPassword(keyName string) string {
	itemMetadata, err := MetaData.GetConfigMetaData(keyName)
	if err == nil {
		value, err := s.readMap(keyName)
		if err != nil {
			log.Errorf("Error while geting %v, error: %+v", keyName, err)
			return ""
		}
		if itemMetadata.Type == PasswordType {
			return value.GetPassword()
		}
	}
	return ""
}

// GetBool ...
func (s *ConfigureStore) GetBool(keyName string) bool {
	itemMetadata, err := MetaData.GetConfigMetaData(keyName)
	if err == nil {
		value, err := s.readMap(keyName)
		if err != nil {
			log.Errorf("Error while geting %v, error: %+v", keyName, err)
			return false
		}
		if itemMetadata.Type == BoolType {
			return value.GetBool()
		}
	}
	return false
}

// UpdateConfig ...
func (s *ConfigureStore) UpdateConfig(cfg map[string]string) error {
	for key, value := range cfg {
		err := s.UpdateConfigValue(key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateConfigValue ...
func (s *ConfigureStore) UpdateConfigValue(keyName string, value string) error {
	itemMetadata, err := MetaData.GetConfigMetaData(keyName)
	c := &ConfigureValue{Key: keyName, Value: value}
	if err == nil {
		if itemMetadata.Type == PasswordType {
			err := c.Set(keyName, value)
			if err != nil {
				return err
			}
		} else {
			err := c.Validate()
			if err != nil {
				return err
			}
		}
		s.writeMap(keyName, c)
	}
	return err
}

// GetDatabaseCfg ...
func (s *ConfigureStore) GetDatabaseCfg() *models.Database {
	return &models.Database{
		Type: s.GetString(common.DatabaseType),
		PostGreSQL: &models.PostGreSQL{
			Host:     s.GetString(common.PostGreSQLHOST),
			Port:     s.GetInt(common.PostGreSQLPort),
			Username: s.GetString(common.PostGreSQLUsername),
			Password: s.GetPassword(common.PostGreSQLPassword),
			Database: s.GetString(common.PostGreSQLDatabase),
			SSLMode:  s.GetString(common.PostGreSQLSSLMode),
		},
	}
}

// GetCfgs ...
func (s *ConfigureStore) GetCfgs() map[string]interface{} {
	resultMap := map[string]interface{}{}
	s.configureValues.Range(func(key, value interface{}) bool {
		if result, suc := value.(Value); suc {
			keyName := result.GetKey()
			itemMetadata, err := MetaData.GetConfigMetaData(keyName)
			if err != nil {
				log.Errorf("Failed to get metadata for key %v", keyName)
			} else if itemMetadata.Type == BoolType {
				resultMap[keyName] = result.GetBool()
			} else if itemMetadata.Type == IntType {
				resultMap[keyName] = result.GetInt()
			} else if itemMetadata.Type == Int64Type {
				resultMap[keyName] = result.GetInt64()
			} else if itemMetadata.Type == PasswordType {
				resultMap[keyName] = result.GetString()
			} else if itemMetadata.Type == StringType {
				resultMap[keyName] = result.GetString()
			}
		}
		return true
	})
	return resultMap

}
