// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metadata

import (
	"errors"

	"github.com/goharbor/harbor/src/lib/log"
)

var (
	// ErrNotDefined ...
	ErrNotDefined = errors.New("configure item is not defined in metadata")
	// ErrTypeNotMatch ...
	ErrTypeNotMatch = errors.New("the required value doesn't matched with metadata defined")
	// ErrInvalidData ...
	ErrInvalidData = errors.New("the data provided is invalid")
	// ErrValueNotSet ...
	ErrValueNotSet = errors.New("the configure value is not set")
	// ErrStringValueIsEmpty ...
	ErrStringValueIsEmpty = errors.New("the configure value can not be empty")
)

// ConfigureValue - struct to hold a actual value, also include the name of config metadata.
type ConfigureValue struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// NewCfgValue ... Create checked config value
func NewCfgValue(name, value string) (*ConfigureValue, error) {
	result := &ConfigureValue{}
	err := result.Set(name, value)
	if err != nil {
		result.Name = name // Keep name to trace error
	}
	return result, err
}

// GetString - Get the string value of current configure
func (c *ConfigureValue) GetString() string {
	// Any type has the string value
	if _, ok := Instance().GetByName(c.Name); ok {
		return c.Value
	}
	return ""
}

// GetName ...
func (c *ConfigureValue) GetName() string {
	return c.Name
}

// GetInt - return the int value of current value
func (c *ConfigureValue) GetInt() int {
	if item, ok := Instance().GetByName(c.Name); ok {
		val, err := item.ItemType.get(c.Value)
		if err != nil {
			log.Errorf("GetInt failed, error: %+v", err)
			return 0
		}
		if intValue, suc := val.(int); suc {
			return intValue
		}
	}
	log.Errorf("GetInt failed, the current value's metadata is not defined, %+v", c)
	return 0
}

// GetInt64 - return the int64 value of current value
func (c *ConfigureValue) GetInt64() int64 {
	if item, ok := Instance().GetByName(c.Name); ok {
		val, err := item.ItemType.get(c.Value)
		if err != nil {
			log.Errorf("GetInt64 failed, error: %+v", err)
			return 0
		}
		if int64Value, suc := val.(int64); suc {
			return int64Value
		}
	}
	log.Errorf("GetInt64 failed, the current value's metadata is not defined, %+v", c)
	return 0
}

// GetFloat64 - return the float64 value of current value
func (c *ConfigureValue) GetFloat64() float64 {
	if item, ok := Instance().GetByName(c.Name); ok {
		val, err := item.ItemType.get(c.Value)
		if err != nil {
			log.Errorf("GetFloat64 failed, error: %+v", err)
			return 0
		}
		if float64Value, suc := val.(float64); suc {
			return float64Value
		}
	}
	log.Errorf("GetFloat64 failed, the current value's metadata is not defined, %+v", c)
	return 0
}

// GetBool - return the bool value of current setting
func (c *ConfigureValue) GetBool() bool {
	if item, ok := Instance().GetByName(c.Name); ok {
		val, err := item.ItemType.get(c.Value)
		if err != nil {
			log.Errorf("GetBool failed, error: %+v", err)
			return false
		}
		if boolValue, suc := val.(bool); suc {
			return boolValue
		}
	}
	log.Errorf("GetBool failed, the current value's metadata is not defined, %+v", c)
	return false
}

// GetStringToStringMap - return the string to string map of current value
func (c *ConfigureValue) GetStringToStringMap() map[string]string {
	result := map[string]string{}
	if item, ok := Instance().GetByName(c.Name); ok {
		val, err := item.ItemType.get(c.Value)
		if err != nil {
			log.Errorf("The GetStringToStringMap failed, error: %+v", err)
			return result
		}
		if mapValue, suc := val.(map[string]string); suc {
			return mapValue
		}
	}
	log.Errorf("GetStringToStringMap failed, current value's metadata is not defined, %+v", c)
	return result
}

// GetAnyType get the interface{} of current value
func (c *ConfigureValue) GetAnyType() (interface{}, error) {
	if item, ok := Instance().GetByName(c.Name); ok {
		return item.ItemType.get(c.Value)
	}
	return nil, ErrNotDefined
}

// Validate - to validate configure items, if passed, return nil, else return error
func (c *ConfigureValue) Validate() error {
	if item, ok := Instance().GetByName(c.Name); ok {
		return item.ItemType.validate(c.Value)
	}
	return ErrNotDefined
}

// GetPassword ...
func (c *ConfigureValue) GetPassword() string {
	if _, ok := Instance().GetByName(c.Name); ok {
		return c.Value
	}
	log.Errorf("GetPassword failed, metadata not defined: %v", c.Name)
	return ""
}

// Set - set this configure item to configure store
func (c *ConfigureValue) Set(name, value string) error {
	if item, ok := Instance().GetByName(name); ok {
		err := item.ItemType.validate(value)
		if err == nil {
			c.Name = name
			c.Value = value
			return nil
		}
		return err
	}
	return ErrNotDefined
}
