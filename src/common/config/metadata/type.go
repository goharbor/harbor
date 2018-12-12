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
	"encoding/json"
	"strconv"
)

// Type - Use this interface to define and encapsulate the behavior of validation and transformation
type Type interface {
	// validate the configure value
	validate(str string) error
	// get the real type of current value, if it is int, return int, if it is string return string etc.
	get(str string) (interface{}, error)
}

// StringType ...
type StringType struct {
}

func (t *StringType) validate(str string) error {
	return nil
}

func (t *StringType) get(str string) (interface{}, error) {
	return str, nil
}

// IntType ..
type IntType struct {
}

func (t *IntType) validate(str string) error {
	_, err := strconv.Atoi(str)
	return err
}

// GetInt ...
func (t *IntType) get(str string) (interface{}, error) {
	return strconv.Atoi(str)
}

// LdapScopeType - The LDAP scope is a int type, but its is limit to 0, 1, 2
type LdapScopeType struct {
	IntType
}

// Validate - Verify the range is limited
func (t *LdapScopeType) validate(str string) error {
	if str == "0" || str == "1" || str == "2" {
		return nil
	}
	return ErrInvalidData
}

// Int64Type ...
type Int64Type struct {
}

func (t *Int64Type) validate(str string) error {
	_, err := strconv.ParseInt(str, 10, 64)
	return err
}

// GetInt64 ...
func (t *Int64Type) get(str string) (interface{}, error) {
	return strconv.ParseInt(str, 10, 64)
}

// BoolType ...
type BoolType struct {
}

func (t *BoolType) validate(str string) error {
	_, err := strconv.ParseBool(str)
	return err
}

func (t *BoolType) get(str string) (interface{}, error) {
	return strconv.ParseBool(str)
}

// PasswordType ...
type PasswordType struct {
}

func (t *PasswordType) validate(str string) error {
	return nil
}

func (t *PasswordType) get(str string) (interface{}, error) {
	return str, nil
}

// MapType ...
type MapType struct {
}

func (t *MapType) validate(str string) error {
	result := map[string]interface{}{}
	err := json.Unmarshal([]byte(str), &result)
	return err
}

func (t *MapType) get(str string) (interface{}, error) {
	result := map[string]string{}
	err := json.Unmarshal([]byte(str), &result)
	return result, err
}
