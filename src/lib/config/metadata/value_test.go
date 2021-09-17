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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testingMetaDataArray = []Item{
	{Name: "ldap_search_scope", ItemType: &LdapScopeType{}, Scope: "system", Group: "ldapbasic"},
	{Name: "ldap_search_dn", ItemType: &StringType{}, Scope: "user", Group: "ldapbasic"},
	{Name: "ulimit", ItemType: &Int64Type{}, Scope: "user", Group: "ldapbasic"},
	{Name: "ldap_verify_cert", ItemType: &BoolType{}, Scope: "user", Group: "ldapbasic"},
	{Name: "sample_map_setting", ItemType: &MapType{}, Scope: "user", Group: "ldapbasic"},
	{Name: "scan_all_policy", ItemType: &MapType{}, Scope: "user", Group: "basic"},
	{Name: "sample_rate", ItemType: &Float64Type{}, Scope: "system", Group: "basic"},
}

// createCfgValue ... Create a ConfigureValue object, only used in test
func createCfgValue(name, value string) *ConfigureValue {
	result := &ConfigureValue{}
	err := result.Set(name, value)
	if err != nil {
		fmt.Printf("failed to create ConfigureValue name:%v, value:%v, error %v\n", name, value, err)
		result.Name = name // Keep name to trace error
	}
	return result
}

func TestConfigureValue_GetBool(t *testing.T) {
	assert.Equal(t, createCfgValue("ldap_verify_cert", "true").GetBool(), true)
	assert.Equal(t, createCfgValue("unknown", "false").GetBool(), false)
}

func TestConfigureValue_GetString(t *testing.T) {
	assert.Equal(t, createCfgValue("ldap_url", "ldaps://ldap.vmware.com").GetString(), "ldaps://ldap.vmware.com")
}

func TestConfigureValue_GetStringToStringMap(t *testing.T) {
	Instance().initFromArray(testingMetaDataArray)
	val, err := createCfgValue("sample_map_setting", `{"sample":"abc"}`).GetAnyType()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, val, map[string]interface{}{"sample": "abc"})
	Instance().init()
}
func TestConfigureValue_GetInt(t *testing.T) {
	assert.Equal(t, createCfgValue("ldap_timeout", "5").GetInt(), 5)
}

func TestConfigureValue_GetInt64(t *testing.T) {
	Instance().initFromArray(testingMetaDataArray)
	assert.Equal(t, createCfgValue("ulimit", "99999").GetInt64(), int64(99999))
}

func TestConfigureValue_GetFloat64(t *testing.T) {
	Instance().initFromArray(testingMetaDataArray)
	assert.Equal(t, createCfgValue("sample_rate", "0.5").GetFloat64(), float64(0.5))
}

func TestNewScanAllPolicy(t *testing.T) {
	Instance().initFromArray(testingMetaDataArray)
	value, err := NewCfgValue("scan_all_policy", `{"parameter":{"daily_time":0},"type":"daily"}`)
	if err != nil {
		t.Errorf("Can not create scan all policy err: %v", err)
	}
	fmt.Printf("value %v\n", value.GetString())
}
