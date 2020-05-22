package driver

import (
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
		"ldap_url":         "ldaps://ldap.vmware.com",
		"ldap_scope":       5,
		"ldap_verify_cert": true,
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
	configMap, err := httpDriver.Load()
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
	err := httpDriver.Save(configMap)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ldap://www.example.com", configMapForTest["ldap_url"])
	assert.Equal(t, float64(10), configMapForTest["ldap_timeout"])
	assert.Equal(t, false, configMapForTest["ldap_verify_cert"])

}
