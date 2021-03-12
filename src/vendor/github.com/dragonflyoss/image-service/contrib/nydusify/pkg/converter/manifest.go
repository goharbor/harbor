// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	digest "github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"

	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/backend"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/converter/provider"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/remote"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/utils"
)

const defaultOS = "linux"
const defaultArch = "amd64"

// manifestManager merges OCI and Nydus manifest, pushes them to
// remote registry
type manifestManager struct {
	sourceProvider provider.SourceProvider
	backend        backend.Backend
	remote         *remote.Remote
	multiPlatform  bool
	dockerV2Format bool
}

// Try to get manifests from exists target image
func (mm *manifestManager) getExistsManifests(ctx context.Context) ([]ocispec.Descriptor, error) {
	desc, err := mm.remote.Resolve(ctx)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return []ocispec.Descriptor{}, nil
		}
		return nil, errors.Wrap(err, "Resolve image manifest index")
	}

	if desc.MediaType == images.MediaTypeDockerSchema2ManifestList ||
		desc.MediaType == ocispec.MediaTypeImageIndex {

		reader, err := mm.remote.Pull(ctx, *desc, true)
		if err != nil {
			return nil, errors.Wrap(err, "Pull image manifest index")
		}
		defer reader.Close()

		indexBytes, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, errors.Wrap(err, "Read image manifest index")
		}

		var index ocispec.Index
		if err := json.Unmarshal(indexBytes, &index); err != nil {
			return nil, errors.Wrap(err, "Unmarshal image manifest index")
		}

		return index.Manifests, nil
	}

	if desc.MediaType == images.MediaTypeDockerSchema2Manifest ||
		desc.MediaType == ocispec.MediaTypeImageManifest {
		return []ocispec.Descriptor{*desc}, nil
	}

	return []ocispec.Descriptor{}, nil
}

func (mm *manifestManager) makeManifestIndex(
	ctx context.Context, ociManifest, nydusManifest ocispec.Descriptor,
) (*ocispec.Index, error) {
	manifestDescs, err := mm.getExistsManifests(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Get image manifest index")
	}

	foundOCI := false
	foundNydus := false
	for idx, desc := range manifestDescs {
		if desc.Platform != nil {
			if desc.Platform.OS == "linux" && desc.Platform.Architecture == "amd64" ||
				desc.Platform.OS == "" && desc.Platform.Architecture == "" {
				if desc.Platform.OSFeatures != nil &&
					len(desc.Platform.OSFeatures) == 1 &&
					desc.Platform.OSFeatures[0] == utils.ManifestOSFeatureNydus {
					manifestDescs[idx] = nydusManifest
					foundNydus = true
				} else {
					manifestDescs[idx] = ociManifest
					foundOCI = true
				}
			}
		} else {
			manifestDescs[idx] = ociManifest
			foundOCI = true
		}
	}

	if !foundOCI {
		manifestDescs = append(manifestDescs, ociManifest)
	}

	if !foundNydus {
		manifestDescs = append(manifestDescs, nydusManifest)
	}

	// Merge exists OCI manifests and Nydus manifest to manifest index
	index := ocispec.Index{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		Manifests: manifestDescs,
	}

	return &index, nil
}

func (mm *manifestManager) Push(ctx context.Context, buildLayers []*buildLayer) error {
	layers := []ocispec.Descriptor{}
	blobIDs := []string{}
	for idx, _layer := range buildLayers {
		record := _layer.GetCacheRecord()

		// Maybe no blob file be outputted in this building,
		// so just ignore it in target layers
		if mm.backend == nil && record.NydusBlobDesc != nil {
			layers = append(layers, *record.NydusBlobDesc)
		}

		if record.NydusBlobDesc != nil {
			blobIDs = append(blobIDs, record.NydusBlobDesc.Digest.Hex())
		}

		// Only need to write lastest bootstrap layer in nydus manifest
		if idx == len(buildLayers)-1 {
			blobIDsBytes, err := json.Marshal(blobIDs)
			if err != nil {
				return errors.Wrap(err, "Marshal blob list")
			}
			record.NydusBootstrapDesc.Annotations[utils.LayerAnnotationNydusBlobIDs] = string(blobIDsBytes)
			layers = append(layers, *record.NydusBootstrapDesc)
		}
	}

	ociConfig, err := mm.sourceProvider.Config(ctx)
	if err != nil {
		return errors.Wrap(err, "Get source image config")
	}
	ociConfig.RootFS.DiffIDs = []digest.Digest{}
	ociConfig.History = []ocispec.History{}

	// Remove useless annotations from layer
	validAnnotationKeys := map[string]bool{
		utils.LayerAnnotationNydusBlob:      true,
		utils.LayerAnnotationNydusBlobIDs:   true,
		utils.LayerAnnotationNydusBootstrap: true,
	}
	for idx, desc := range layers {
		layerDiffID := digest.Digest(desc.Annotations[utils.LayerAnnotationUncompressed])
		if layerDiffID == "" {
			layerDiffID = desc.Digest
		}
		ociConfig.RootFS.DiffIDs = append(ociConfig.RootFS.DiffIDs, layerDiffID)
		if desc.Annotations != nil {
			newAnnotations := make(map[string]string)
			for key, value := range desc.Annotations {
				if validAnnotationKeys[key] {
					newAnnotations[key] = value
				}
			}
			layers[idx].Annotations = newAnnotations
		}
	}

	// Push Nydus image config
	configMediaType := ocispec.MediaTypeImageConfig
	if mm.dockerV2Format {
		configMediaType = images.MediaTypeDockerSchema2Config
	}
	configDesc, configBytes, err := utils.MarshalToDesc(ociConfig, configMediaType)
	if err != nil {
		return errors.Wrap(err, "Marshal source image config")
	}

	if err := mm.remote.Push(ctx, *configDesc, true, bytes.NewReader(configBytes)); err != nil {
		return errors.Wrap(err, "Push Nydus image config")
	}

	manifestMediaType := ocispec.MediaTypeImageManifest
	if mm.dockerV2Format {
		manifestMediaType = images.MediaTypeDockerSchema2Manifest
	}

	// Push Nydus image manifest
	nydusManifest := struct {
		MediaType string `json:"mediaType,omitempty"`
		ocispec.Manifest
	}{
		MediaType: manifestMediaType,
		Manifest: ocispec.Manifest{
			Versioned: specs.Versioned{
				SchemaVersion: 2,
			},
			Config: *configDesc,
			Layers: layers,
		},
	}

	nydusManifestDesc, manifestBytes, err := utils.MarshalToDesc(nydusManifest, manifestMediaType)
	if err != nil {
		return errors.Wrap(err, "Marshal Nydus image manifest")
	}
	nydusManifestDesc.Platform = &ocispec.Platform{
		OS:           defaultOS,
		Architecture: defaultArch,
		OSFeatures:   []string{utils.ManifestOSFeatureNydus},
	}

	if !mm.multiPlatform {
		if err := mm.remote.Push(ctx, *nydusManifestDesc, false, bytes.NewReader(manifestBytes)); err != nil {
			return errors.Wrap(err, "Push nydus image manifest")
		}
		return nil
	}

	if err := mm.remote.Push(ctx, *nydusManifestDesc, true, bytes.NewReader(manifestBytes)); err != nil {
		return errors.Wrap(err, "Push nydus image manifest")
	}

	// Push manifest index, includes OCI manifest and Nydus manifest
	ociManifestDesc, err := mm.sourceProvider.Manifest(ctx)
	if err != nil {
		return errors.Wrap(err, "Get source image manifest")
	}
	ociManifestDesc.Platform = &ocispec.Platform{
		OS:           defaultOS,
		Architecture: defaultArch,
	}
	_index, err := mm.makeManifestIndex(ctx, *ociManifestDesc, *nydusManifestDesc)
	if err != nil {
		return errors.Wrap(err, "Make manifest index")
	}

	indexMediaType := ocispec.MediaTypeImageIndex
	if mm.dockerV2Format {
		indexMediaType = images.MediaTypeDockerSchema2ManifestList
	}

	index := struct {
		MediaType string `json:"mediaType,omitempty"`
		ocispec.Index
	}{
		MediaType: indexMediaType,
		Index:     *_index,
	}

	indexDesc, indexBytes, err := utils.MarshalToDesc(index, indexMediaType)
	if err != nil {
		return errors.Wrap(err, "Marshal image manifest index")
	}

	if err := mm.remote.Push(ctx, *indexDesc, false, bytes.NewReader(indexBytes)); err != nil {
		return errors.Wrap(err, "Push image manifest index")
	}

	return nil
}
