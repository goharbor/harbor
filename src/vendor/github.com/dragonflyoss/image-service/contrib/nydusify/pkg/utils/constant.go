// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package utils

const (
	ManifestOSFeatureNydus   = "nydus.remoteimage.v1"
	MediaTypeNydusBlob       = "application/vnd.oci.image.layer.nydus.blob.v1"
	BootstrapFileNameInLayer = "image/image.boot"

	ManifestNydusCache        = "containerd.io/snapshot/nydus-cache"
	ManifestNydusCacheVersion = "v1"

	LayerAnnotationNydusBlob          = "containerd.io/snapshot/nydus-blob"
	LayerAnnotationNydusBlobDigest    = "containerd.io/snapshot/nydus-blob-digest"
	LayerAnnotationNydusBlobSize      = "containerd.io/snapshot/nydus-blob-size"
	LayerAnnotationNydusBlobIDs       = "containerd.io/snapshot/nydus-blob-ids"
	LayerAnnotationNydusBootstrap     = "containerd.io/snapshot/nydus-bootstrap"
	LayerAnnotationNydusSourceChainID = "containerd.io/snapshot/nydus-source-chainid"

	LayerAnnotationUncompressed = "containerd.io/uncompressed"
)
