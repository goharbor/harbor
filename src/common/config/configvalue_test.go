package config

import (
	"errors"
	"reflect"
	"testing"
)

var testingMetaDataArray = []Item{
	{Name: "ldap_search_scope", Type: "int", Scope: "system", Group: "ldapbasic"},
	{Name: "ldap_search_dn", Type: "string", Scope: "user", Group: "ldapbasic"},
	{Name: "ulimit", Type: "int64", Scope: "user", Group: "ldapbasic"},
	{Name: "ldap_verify_cert", Type: "bool", Scope: "user", Group: "ldapbasic"},
	{Name: "sample_map_setting", Type: "map", Scope: "user", Group: "ldapbasic"},
}

func TestConfigureValue_GetString(t *testing.T) {

	type fields struct {
		Key   string
		Value string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{"normal", fields{"ldap_search_dn", "cn=admin,dc=example,dc=com"}, "cn=admin,dc=example,dc=com", false},
	}

	MetaData.InitMetaDataFromArray(testingMetaDataArray)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConfigureValue{
				Key:   tt.fields.Key,
				Value: tt.fields.Value,
			}
			got := c.GetString()
			if got != tt.want {
				t.Errorf("ConfigureValue.GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigureValue_GetInt64(t *testing.T) {
	type fields struct {
		Key   string
		Value string
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		{"normal", fields{"ulimit", "255534223"}, 255534223, false},
	}
	MetaData.InitMetaDataFromArray(testingMetaDataArray)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConfigureValue{
				Key:   tt.fields.Key,
				Value: tt.fields.Value,
			}
			got := c.GetInt64()
			if got != tt.want {
				t.Errorf("ConfigureValue.GetInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigureValue_GetBool(t *testing.T) {
	type fields struct {
		Key   string
		Value string
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{"normal", fields{"ldap_verify_cert", "true"}, true, false},
	}
	MetaData.InitMetaDataFromArray(testingMetaDataArray)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConfigureValue{
				Key:   tt.fields.Key,
				Value: tt.fields.Value,
			}
			got := c.GetBool()
			if got != tt.want {
				t.Errorf("ConfigureValue.GetBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigureValue_GetStringToStringMap(t *testing.T) {
	type fields struct {
		Key   string
		Value string
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]string
		wantErr bool
	}{
		{"normal", fields{"sample_map_setting", `{ "value1":"abc","value2":"def" }`}, map[string]string{"value1": "abc", "value2": "def"}, false},
	}
	MetaData.InitMetaDataFromArray(testingMetaDataArray)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConfigureValue{
				Key:   tt.fields.Key,
				Value: tt.fields.Value,
			}
			got := c.GetStringToStringMap()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConfigureValue.GetStringToStringMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigureValue_GetMap(t *testing.T) {
	type fields struct {
		Key   string
		Value string
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]interface{}
		wantErr bool
	}{
		{"normal", fields{"sample_map_setting", `{ "value1":"abc","value2":"def" }`}, map[string]interface{}{"value1": "abc", "value2": "def"}, false},
	}
	MetaData.InitMetaDataFromArray(testingMetaDataArray)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConfigureValue{
				Key:   tt.fields.Key,
				Value: tt.fields.Value,
			}
			got := c.GetMap()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConfigureValue.GetMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func LDAPScopeValidateFunc(key, value string) error {
	if value == "1" || value == "2" || value == "3" {
		return nil
	}
	return errors.New("The value should between 1, 2, 3")
}

func TestConfigureValue_Validate(t *testing.T) {
	type fields struct {
		Key   string
		Value string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"out of scope", fields{"ldap_search_scope", "4"}, true},
		{"normal", fields{"ldap_search_scope", "3"}, false},
	}

	MetaData.InitMetaDataFromArray(testingMetaDataArray)
	item, err := MetaData.GetConfigMetaData("ldap_search_scope")
	if err != nil {
		t.Errorf("Error occurred when GetConfigMetaData: %v", err)
	}
	item.Validator = LDAPScopeValidateFunc
	MetaData.writeMap("ldap_search_scope", item)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConfigureValue{
				Key:   tt.fields.Key,
				Value: tt.fields.Value,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ConfigureValue.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigureValue_Set(t *testing.T) {
	type fields struct {
		Key   string
		Value string
	}
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"normal", fields{"", ""}, args{"ldap_search_scope", "4"}, true},
		{"normal", fields{"", ""}, args{"ldap_search_scope", "3"}, false},
	}
	MetaData.InitMetaDataFromArray(testingMetaDataArray)

	item, err := MetaData.GetConfigMetaData("ldap_search_scope")
	if err != nil {
		t.Errorf("Error occurred when GetConfigMetaData: %v", err)
	}
	item.Validator = LDAPScopeValidateFunc
	MetaData.writeMap("ldap_search_scope", item)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConfigureValue{
				Key:   tt.fields.Key,
				Value: tt.fields.Value,
			}
			if err := c.Set(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("ConfigureValue.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
