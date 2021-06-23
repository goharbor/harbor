//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package rest

import (
	"context"
	"encoding/json"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func ConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	cfgs := map[string]interface{}{
		"ldap_url":         &Value{Val: "ldaps://ldap.vmware.com", Editable: true},
		"ldap_scope":       &Value{Val: 5, Editable: true},
		"ldap_verify_cert": &Value{Val: true, Editable: true},
	}
	b, err := json.Marshal(cfgs)
	if err != nil {
		return
	}
	w.Write(b)
}

func TestHTTPDriver_Load(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(ConfigGetHandler))
	defer server.Close()
	httpDriver := NewRESTDriver(server.URL)
	configMap, err := httpDriver.Load(context.Background())
	if err != nil {
		t.Errorf("Error when testing http driver %v", err)
	}
	assert.Equal(t, "ldaps://ldap.vmware.com", configMap["ldap_url"])
	// json.Marshal() always convert number to float64, configvalue can handle it by convert it to string
	assert.Equal(t, float64(5), configMap["ldap_scope"])
	assert.Equal(t, true, configMap["ldap_verify_cert"])
}

var configMapForTest = map[string]interface{}{}

func ConfigPutHandler(w http.ResponseWriter, r *http.Request) {
	cfgs := map[string]interface{}{}
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(content, &cfgs)
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range cfgs {
		configMapForTest[k] = v
	}
}

func TestHTTPDriver_Save(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(ConfigPutHandler))
	defer server.Close()
	httpDriver := NewRESTDriver(server.URL)
	configMap := map[string]interface{}{
		"ldap_url":         "ldap://www.example.com",
		"ldap_timeout":     10,
		"ldap_verify_cert": false,
	}
	err := httpDriver.Save(context.Background(), configMap)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ldap://www.example.com", configMapForTest["ldap_url"])
	assert.Equal(t, float64(10), configMapForTest["ldap_timeout"])
	assert.Equal(t, false, configMapForTest["ldap_verify_cert"])

}
