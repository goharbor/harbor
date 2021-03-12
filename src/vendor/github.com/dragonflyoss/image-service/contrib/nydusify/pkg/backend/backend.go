// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type Backend interface {
	// TODO: Hopefully, we can pass `Layer` struct in, thus to be able to cook both
	// file handle and file path.
	Upload(blobID string, blobPath string) error
	Check(blobID string) (bool, error)
}

func NewBackend(backendType, backendConfig string) (Backend, error) {
	switch backendType {
	case "registry":
		return nil, nil
	case "oss":
		var config map[string]string
		if err := json.Unmarshal([]byte(backendConfig), &config); err != nil {
			return nil, errors.Wrap(err, "parse backend config")
		}
		return newOSSBackend(
			config["endpoint"],
			config["bucket_name"],
			config["object_prefix"],
			config["access_key_id"],
			config["access_key_secret"],
		)
	default:
		return nil, fmt.Errorf("unsupported backend type: %s", backendType)
	}

}
