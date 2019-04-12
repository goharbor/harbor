package models

import "time"

// RepPolicy is the model for a ng replication policy.
type RepPolicy struct {
	ID             int64  `orm:"pk;auto;column(id)" json:"id"`
	Name           string `orm:"column(name)" json:"name"`
	Description    string `orm:"column(description)" json:"description"`
	Creator        string `orm:"column(creator)" json:"creator"`
	SrcRegistryID  int64  `orm:"column(src_registry_id)" json:"src_registry_id"`
	SrcNamespaces  string `orm:"column(src_namespaces)" json:"src_namespaces"`
	DestRegistryID int64  `orm:"column(dest_registry_id)" json:"dest_registry_id"`
	DestNamespace  string `orm:"column(dest_namespace)" json:"dest_namespace"`
	Override       bool   `orm:"column(override)" json:"override"`
	Enabled        bool   `orm:"column(enabled)" json:"enabled"`
	// TODO rename the db column to trigger
	Trigger           string    `orm:"column(cron_str)" json:"trigger"`
	Filters           string    `orm:"column(filters)" json:"filters"`
	ReplicateDeletion bool      `orm:"column(replicate_deletion)" json:"replicate_deletion"`
	CreationTime      time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime        time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName set table name for ORM.
func (r *RepPolicy) TableName() string {
	return "replication_policy"
}
