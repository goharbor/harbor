// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"os"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type OSSBackend struct {
	objectPrefix string
	bucket       *oss.Bucket
}

func newOSSBackend(endpoint, bucket, objectPrefix, accessKeyID, accessKeySecret string) (*OSSBackend, error) {
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, errors.Wrap(err, "init oss backend")
	}

	_bucket, err := client.Bucket(bucket)
	if err != nil {
		return nil, errors.Wrap(err, "init oss backend")
	}

	return &OSSBackend{
		objectPrefix: objectPrefix,
		bucket:       _bucket,
	}, nil
}

const (
	splitPartsCount = 4
	// Blob size bigger than 100MB, apply multiparts upload.
	multipartsUploadThreshold = 100 * 1024 * 1024
)

// Upload blob as image layer to oss backend. Depending on blob's size, upload it
// by multiparts method or the normal method
func (b *OSSBackend) Upload(blobID string, blobPath string) error {
	blobID = b.objectPrefix + blobID
	if exist, err := b.bucket.IsObjectExist(blobID); err != nil {
		return err
	} else if exist {
		return nil
	}

	var stat os.FileInfo
	stat, err := os.Stat(blobPath)
	if err != nil {
		return err
	}
	blobSize := stat.Size()

	var needMultiparts bool = false
	// Blob size bigger than 100MB, apply multiparts upload.
	if blobSize >= multipartsUploadThreshold {
		needMultiparts = true
	}

	start := time.Now()

	if needMultiparts {
		logrus.Debugf("Upload %s using multiparts method", blobID)
		chunks, err := oss.SplitFileByPartNum(blobPath, splitPartsCount)
		if err != nil {
			return err
		}

		imur, err := b.bucket.InitiateMultipartUpload(blobID)
		if err != nil {
			return err
		}

		var parts []oss.UploadPart

		g := new(errgroup.Group)
		for _, chunk := range chunks {
			ck := chunk
			g.Go(func() error {
				p, err := b.bucket.UploadPartFromFile(imur, blobPath, ck.Offset, ck.Size, ck.Number)
				if err != nil {
					return err
				}
				// TODO: We don't verify data part MD5 from ETag right now.
				// But we can do it if we have to.
				parts = append(parts, p)
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return errors.Wrap(err, "Uploading parts failed")
		}

		_, err = b.bucket.CompleteMultipartUpload(imur, parts)
		if err != nil {
			return err
		}
	} else {
		reader, err := os.Open(blobPath)
		if err != nil {
			return err
		}
		defer reader.Close()
		err = b.bucket.PutObject(blobID, reader)
		if err != nil {
			return err
		}
	}

	end := time.Now()
	elapsed := end.Sub(start)
	logrus.Debugf("Uploading blob %s costs %s", blobID, elapsed)

	return err
}

func (b *OSSBackend) Check(blobID string) (bool, error) {
	blobID = b.objectPrefix + blobID
	return b.bucket.IsObjectExist(blobID)
}
