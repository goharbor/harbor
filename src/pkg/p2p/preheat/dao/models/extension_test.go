package models

import (
	"encoding/json"
	"testing"
)

var testingJSONData = `
{
	"description": "adf",
	"endpoint": "adf",
	"auth_mode": "BASIC",
	"auth_data": {
	  "adsf": "asdf"
	},
	"enabled": true
}
`

func TestPropertyDecode(t *testing.T) {
	propertySet := make(PropertySet)
	if err := json.Unmarshal([]byte(testingJSONData), &propertySet); err != nil {
		t.Fatal(err)
	}

	meta := &Metadata{}
	if err := propertySet.Apply(meta); err != nil {
		t.Fatal(err)
	}
}

func TestPropertySet(t *testing.T) {
	flag := theChangableProperties.Match("enabled")
	if !flag {
		t.Errorf("expect true flag but got false")
	}
}
