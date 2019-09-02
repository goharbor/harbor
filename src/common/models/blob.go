package models

import (
	"time"

	"github.com/docker/distribution/manifest/schema2"
)

// Blob holds the details of a blob.
type Blob struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Digest       string    `orm:"column(digest)" json:"digest"`
	ContentType  string    `orm:"column(content_type)" json:"content_type"`
	Size         int64     `orm:"column(size)" json:"size"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
}

// TableName ...
func (b *Blob) TableName() string {
	return "blob"
}

// IsForeignLayer returns true if the blob is foreign layer
func (b *Blob) IsForeignLayer() bool {
	return b.ContentType == schema2.MediaTypeForeignLayer
}

// BlobQuery ...
type BlobQuery struct {
	Digest      string
	ContentType string
	Digests     []string
	Pagination
}
