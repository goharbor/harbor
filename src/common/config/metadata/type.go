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

// Package metadata define config related metadata
package metadata

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/common"
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

// NonEmptyStringType ...
type NonEmptyStringType struct {
	StringType
}

func (t *NonEmptyStringType) validate(str string) error {
	if len(strings.TrimSpace(str)) == 0 {
		return ErrStringValueIsEmpty
	}
	return nil
}

// AuthModeType ...
type AuthModeType struct {
	StringType
}

func (t *AuthModeType) validate(str string) error {
	if str == common.LDAPAuth || str == common.DBAuth || str == common.UAAAuth || str == common.HTTPAuth || str == common.OIDCAuth {
		return nil
	}
	return fmt.Errorf("invalid %s, shoud be one of %s, %s, %s, %s, %s",
		common.AUTHMode, common.DBAuth, common.LDAPAuth, common.UAAAuth, common.HTTPAuth, common.OIDCAuth)
}

// ProjectCreationRestrictionType ...
type ProjectCreationRestrictionType struct {
	StringType
}

func (t *ProjectCreationRestrictionType) validate(str string) error {
	if !(str == common.ProCrtRestrAdmOnly || str == common.ProCrtRestrEveryone) {
		return fmt.Errorf("invalid %s, should be %s or %s",
			common.ProjectCreationRestriction,
			common.ProCrtRestrAdmOnly,
			common.ProCrtRestrEveryone)
	}
	return nil
}

// IntType ..
type IntType struct {
}

func (t *IntType) validate(str string) error {
	_, err := strconv.Atoi(str)
	return err
}

func (t *IntType) get(str string) (interface{}, error) {
	return strconv.Atoi(str)
}

// PortType ...
type PortType struct {
	IntType
}

func (t *PortType) validate(str string) error {
	val, err := strconv.Atoi(str)
	if err != nil {
		return err
	}
	if val < 0 {
		return fmt.Errorf("network port should be greater than 0")
	}

	if val > 65535 {
		return fmt.Errorf("network port should be less than 65535")
	}

	return err
}

// LdapScopeType - The LDAP scope is a int type, but its is limit to 0, 1, 2
type LdapScopeType struct {
	IntType
}

// validate - Verify the range is limited
func (t *LdapScopeType) validate(str string) error {
	if str == "0" || str == "1" || str == "2" {
		return nil
	}
	return fmt.Errorf("invalid scope, should be %d, %d or %d",
		common.LDAPScopeBase,
		common.LDAPScopeOnelevel,
		common.LDAPScopeSubtree)
}

// Int64Type ...
type Int64Type struct {
}

func (t *Int64Type) validate(str string) error {
	_, err := parseInt64(str)
	return err
}

func (t *Int64Type) get(str string) (interface{}, error) {
	return parseInt64(str)
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
	result := map[string]interface{}{}
	err := json.Unmarshal([]byte(str), &result)
	return result, err
}

// QuotaType ...
type QuotaType struct {
	Int64Type
}

func (t *QuotaType) validate(str string) error {
	val, err := parseInt64(str)
	if err != nil {
		return err
	}

	if val <= 0 && val != -1 {
		return fmt.Errorf("quota value should be -1 or great than zero")
	}

	return nil
}

// parseInt64 returns int64 from string which support scientific notation
func parseInt64(str string) (int64, error) {
	val, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		return val, nil
	}

	fval, err := strconv.ParseFloat(str, 64)
	if err == nil && fval == math.Trunc(fval) {
		return int64(fval), nil
	}

	return 0, fmt.Errorf("invalid int64 string: %s", str)
}
