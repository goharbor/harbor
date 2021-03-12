// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/remote"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/utils"

	"github.com/containerd/containerd/images"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Parser parses OCI & Nydus image manifest, manifest index and
// image config into Parsed object, see the Nydus image example:
// examples/manifest/index.json, examples/manifest/manifest.json.
type Parser struct {
	Remote *remote.Remote
}

// Image presents image contents.
type Image struct {
	Desc     ocispec.Descriptor
	Manifest ocispec.Manifest
	Config   ocispec.Image
}

// Parsed presents OCI and Nydus image manifest.
type Parsed struct {
	Index      *ocispec.Index
	OCIImage   *Image
	NydusImage *Image
}

// New creates Nydus image parser instance.
func New(remote *remote.Remote) *Parser {
	return &Parser{
		Remote: remote,
	}
}

// Try to find the topmost layer in Nydus manifest, it should
// be a Nydus bootstrap layer, see examples/manifest/manifest.json
func findNydusBootstrapDesc(manifest *ocispec.Manifest) *ocispec.Descriptor {
	layers := manifest.Layers
	if len(layers) != 0 {
		desc := &layers[len(layers)-1]
		if (desc.MediaType == ocispec.MediaTypeImageLayerGzip ||
			desc.MediaType == images.MediaTypeDockerSchema2LayerGzip) &&
			desc.Annotations[utils.LayerAnnotationNydusBootstrap] == "true" {
			return desc
		}
	}
	return nil
}

func (parser *Parser) pull(ctx context.Context, desc *ocispec.Descriptor, res interface{}) error {
	reader, err := parser.Remote.Pull(ctx, *desc, true)
	if err != nil {
		return errors.Wrap(err, "pull image resource")
	}
	defer reader.Close()

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "read image resource")
	}

	if err := json.Unmarshal(bytes, res); err != nil {
		return errors.Wrap(err, "unmarshal image resource")
	}

	return nil
}

func (parser *Parser) pullManifest(ctx context.Context, desc *ocispec.Descriptor) (*ocispec.Manifest, error) {
	var manifest ocispec.Manifest
	if err := parser.pull(ctx, desc, &manifest); err != nil {
		return nil, errors.Wrap(err, "pull image manifest")
	}
	return &manifest, nil
}

func (parser *Parser) pullConfig(ctx context.Context, desc *ocispec.Descriptor) (*ocispec.Image, error) {
	var config ocispec.Image
	if err := parser.pull(ctx, desc, &config); err != nil {
		return nil, errors.Wrap(err, "pull image config")
	}
	return &config, nil
}

func (parser *Parser) pullIndex(ctx context.Context, desc *ocispec.Descriptor) (*ocispec.Index, error) {
	var index ocispec.Index
	if err := parser.pull(ctx, desc, &index); err != nil {
		return nil, errors.Wrap(err, "pull image index")
	}
	return &index, nil
}

func (parser *Parser) parseImage(
	ctx context.Context, desc *ocispec.Descriptor, onlyManifest *ocispec.Manifest,
) (*Image, error) {
	var manifest *ocispec.Manifest
	var err error
	if onlyManifest != nil {
		manifest = onlyManifest
	} else {
		manifest, err = parser.pullManifest(ctx, desc)
		if err != nil {
			return nil, errors.Wrap(err, "pull image manifest")
		}
	}
	config, err := parser.pullConfig(ctx, &manifest.Config)
	if err != nil {
		return nil, errors.Wrap(err, "pull image config")
	}
	return &Image{
		Desc:     *desc,
		Manifest: *manifest,
		Config:   *config,
	}, nil
}

// PullNydusBootstrap pulls Nydus bootstrap layer from Nydus image.
func (parser *Parser) PullNydusBootstrap(ctx context.Context, image *Image) (io.ReadCloser, error) {
	bootstrapDesc := findNydusBootstrapDesc(&image.Manifest)
	if bootstrapDesc == nil {
		return nil, fmt.Errorf("not found Nydus bootstrap layer in manifest")
	}
	reader, err := parser.Remote.Pull(ctx, *bootstrapDesc, true)
	if err != nil {
		return nil, errors.Wrap(err, "pull Nydus bootstrap layer")
	}
	return reader, nil
}

// Parse parses Nydus image reference into Parsed object.
func (parser *Parser) Parse(ctx context.Context) (*Parsed, error) {
	logrus.Infof("Parsing image %s", parser.Remote.Ref)

	parsed := Parsed{}

	imageDesc, err := parser.Remote.Resolve(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "resolve image")
	}

	var ociDesc *ocispec.Descriptor
	var nydusDesc *ocispec.Descriptor
	var onlyManifest *ocispec.Manifest

	switch imageDesc.MediaType {
	// Handle image manifest
	case ocispec.MediaTypeImageManifest, images.MediaTypeDockerSchema2Manifest:
		onlyManifest, err = parser.pullManifest(ctx, imageDesc)
		if err != nil {
			return nil, err
		}
		bootstrapDesc := findNydusBootstrapDesc(onlyManifest)
		if bootstrapDesc != nil {
			nydusDesc = imageDesc
		} else {
			ociDesc = imageDesc
		}

	// Handle image manifest index
	case ocispec.MediaTypeImageIndex, images.MediaTypeDockerSchema2ManifestList:
		index, err := parser.pullIndex(ctx, imageDesc)
		if err != nil {
			return nil, err
		}
		parsed.Index = index

		for idx := range index.Manifests {
			desc := index.Manifests[idx]
			if desc.Platform != nil {
				if desc.Platform.OS == "linux" && desc.Platform.Architecture == "amd64" ||
					desc.Platform.OS == "" && desc.Platform.Architecture == "" {
					if desc.Platform.OSFeatures != nil &&
						len(desc.Platform.OSFeatures) == 1 &&
						desc.Platform.OSFeatures[0] == utils.ManifestOSFeatureNydus {
						nydusDesc = &desc
					} else {
						ociDesc = &desc
					}
				}
			} else {
				ociDesc = &desc
			}
		}
	}

	if ociDesc != nil {
		parsed.OCIImage, err = parser.parseImage(ctx, ociDesc, onlyManifest)
		if err != nil {
			return nil, errors.Wrap(err, "parse OCI image")
		}
	}

	if nydusDesc != nil {
		parsed.NydusImage, err = parser.parseImage(ctx, nydusDesc, onlyManifest)
		if err != nil {
			return nil, errors.Wrap(err, "parse Nydus image")
		}
	}

	return &parsed, nil
}
