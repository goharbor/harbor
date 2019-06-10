package models

import (
	"encoding/json"
)

// FilterMetadata defines type and argument information for a retention filter in a policy
type FilterMetadata struct {
	ID   int64  `orm:"column(id);pk;auto" json:"id"`
	Type string `orm:"column(type)" json:"type"`

	RawOptions string                 `orm:"column(options);type(json)" json:"-"`
	Options    map[string]interface{} `orm:"-" json:"options"`

	Policy *Policy `orm:"column(policy);rel(fk)" json:"-"`
}

// SyncJSONToORM marshals metadata stored in f.Options into f.RawOptions
func (f *FilterMetadata) SyncJSONToORM() error {
	bytes, err := json.Marshal(&f.Options)
	if err != nil {
		return err
	}

	f.RawOptions = string(bytes)

	return nil
}

// SyncORMToJSON unmarshals metadata stored in f.RawOptions to f.Options
func (f *FilterMetadata) SyncORMToJSON() error {
	if f.RawOptions == "" {
		f.Options = map[string]interface{}{}
		return nil
	}

	return json.Unmarshal([]byte(f.RawOptions), &f.Options)
}

// TableName returns the name of the table to store filter metadata in
func (f *FilterMetadata) TableName() string {
	return "retention_filter_metadata"
}
