// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

// Package provider abstract interface to adapt to different build environments,
// the provider includes these components:
// 	logger: output build progress for nydusify or buildkitd/buildctl;
// 	remote: create a remote resolver, it communicates with remote registry;
// 	source: responsible for getting image manifest, config, and mounting layer;
// Provider provides a default implementation, so we can use it in Nydusify
// directly, but we need to implement it in buildkit or other any projects
// which want to import nydusify package.
package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/identity"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"

	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/parser"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/remote"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/utils"
)

const defaultOS = "linux"
const defaultArch = "amd64"

// SourceLayer is a layer of source image
type SourceLayer interface {
	Mount(ctx context.Context) (string, func() error, error)
	Size() int64
	Digest() digest.Digest
	ChainID() digest.Digest
	ParentChainID() *digest.Digest
}

// SourceProvider provides resource of source image
type SourceProvider interface {
	Manifest(ctx context.Context) (*ocispec.Descriptor, error)
	Config(ctx context.Context) (*ocispec.Image, error)
	Layers(ctx context.Context) ([]SourceLayer, error)
}

type defaultSourceProvider struct {
	workDir string
	image   parser.Image
	remote  *remote.Remote
}

type defaultSourceLayer struct {
	remote        *remote.Remote
	mountDir      string
	desc          ocispec.Descriptor
	chainID       digest.Digest
	parentChainID *digest.Digest
}

func (sp *defaultSourceProvider) Manifest(ctx context.Context) (*ocispec.Descriptor, error) {
	return &sp.image.Desc, nil
}

func (sp *defaultSourceProvider) Config(ctx context.Context) (*ocispec.Image, error) {
	return &sp.image.Config, nil
}

func (sp *defaultSourceProvider) Layers(ctx context.Context) ([]SourceLayer, error) {
	layers := sp.image.Manifest.Layers
	diffIDs := sp.image.Config.RootFS.DiffIDs
	if len(layers) != len(diffIDs) {
		return nil, fmt.Errorf("Mismatched fs layers (%d) and diff ids (%d)", len(layers), len(diffIDs))
	}

	var parentChainID *digest.Digest
	sourceLayers := []SourceLayer{}

	for i, desc := range layers {
		layerDigest := desc.Digest
		chainID := identity.ChainID(diffIDs[:i+1])
		layer := &defaultSourceLayer{
			remote:        sp.remote,
			mountDir:      filepath.Join(sp.workDir, layerDigest.String()),
			desc:          desc,
			chainID:       chainID,
			parentChainID: parentChainID,
		}
		sourceLayers = append(sourceLayers, layer)
		parentChainID = &chainID
	}

	return sourceLayers, nil
}

func (sl *defaultSourceLayer) Mount(ctx context.Context) (string, func() error, error) {
	digestStr := sl.desc.Digest.String()

	// Pull the layer from source
	reader, err := sl.remote.Pull(ctx, sl.desc, true)
	if err != nil {
		return "", nil, errors.Wrap(err, fmt.Sprintf("Decompress source layer %s", digestStr))
	}
	defer reader.Close()

	// Decompress layer from source stream
	if err := utils.UnpackTargz(ctx, sl.mountDir, reader); err != nil {
		return "", nil, errors.Wrap(err, fmt.Sprintf("Decompress source layer %s", digestStr))
	}

	umount := func() error {
		return os.RemoveAll(sl.mountDir)
	}

	return sl.mountDir, umount, nil
}

func (sl *defaultSourceLayer) Digest() digest.Digest {
	return sl.desc.Digest
}

func (sl *defaultSourceLayer) Size() int64 {
	return sl.desc.Size
}

func (sl *defaultSourceLayer) ChainID() digest.Digest {
	return sl.chainID
}

func (sl *defaultSourceLayer) ParentChainID() *digest.Digest {
	return sl.parentChainID
}

// DefaultSource pulls image layers from specify image reference
func DefaultSource(ctx context.Context, remote *remote.Remote, workDir string) (SourceProvider, error) {
	parser := parser.New(remote)
	parsed, err := parser.Parse(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Parse source image")
	}

	if parsed.OCIImage == nil {
		return nil, errors.Wrap(err, "Not found linux/amd64 manifest in source image")
	}

	sp := defaultSourceProvider{
		workDir: workDir,
		image:   *parsed.OCIImage,
		remote:  remote,
	}

	return &sp, nil
}
