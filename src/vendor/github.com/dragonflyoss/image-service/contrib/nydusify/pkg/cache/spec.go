// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type CacheManifest struct {
	MediaType string `json:"mediaType,omitempty"`
	ocispec.Manifest
}

type CacheRecord struct {
	SourceChainID        digest.Digest
	NydusBlobDesc        *ocispec.Descriptor
	NydusBootstrapDesc   *ocispec.Descriptor
	NydusBootstrapDiffID digest.Digest
}
