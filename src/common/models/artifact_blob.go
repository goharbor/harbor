package models

import (
	"time"
)

// ArtifactAndBlob holds the relationship between manifest and blob.
type ArtifactAndBlob struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	DigestAF     string    `orm:"column(digest_af)" json:"digest_af"`
	DigestBlob   string    `orm:"column(digest_blob)" json:"digest_blob"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
}

// TableName ...
func (afb *ArtifactAndBlob) TableName() string {
	return "artifact_blob"
}
