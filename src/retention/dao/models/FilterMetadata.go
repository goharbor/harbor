package models

import (
	"encoding/json"
)

type FilterMetadata struct {
	ID   int64  `orm:"column(id);pk;auto" json:"id"`
	Type string `orm:"column(type)" json:"type"`

	RawOptions string                 `orm:"column(options);type(json)" json:"-"`
	Options    map[string]interface{} `orm:"-" json:"options"`

	Policy *Policy `orm:"column(policy);rel(fk)" json:"-"`
}

func (f *FilterMetadata) SyncJsonToORM() error {
	if bytes, err := json.Marshal(&f.Options); err != nil {
		return err
	} else {
		f.RawOptions = string(bytes)
	}

	return nil
}

func (f *FilterMetadata) SyncORMToJson() error {
	if f.RawOptions == "" {
		f.Options = map[string]interface{}{}
		return nil
	}

	return json.Unmarshal([]byte(f.RawOptions), &f.Options)
}

func (f *FilterMetadata) TableName() string {
	return "retention_filter_metadata"
}
