package oras

import (
	"context"
	"encoding/json"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	artifact "github.com/deislabs/oras/pkg/artifact"
	digest "github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Push pushes files to the remote
func Push(ctx context.Context, resolver remotes.Resolver, ref string, provider content.Provider, descriptors []ocispec.Descriptor, opts ...PushOpt) (ocispec.Descriptor, error) {
	if resolver == nil {
		return ocispec.Descriptor{}, ErrResolverUndefined
	}
	opt := pushOptsDefaults()
	for _, o := range opts {
		if err := o(opt); err != nil {
			return ocispec.Descriptor{}, err
		}
	}
	if opt.validateName != nil {
		for _, desc := range descriptors {
			if err := opt.validateName(desc); err != nil {
				return ocispec.Descriptor{}, err
			}
		}
	}

	pusher, err := resolver.Pusher(ctx, ref)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	desc, store, err := pack(provider, descriptors, opt)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	var wrapper func(images.Handler) images.Handler
	if len(opt.baseHandlers) > 0 {
		wrapper = func(h images.Handler) images.Handler {
			return images.Handlers(append(opt.baseHandlers, h)...)
		}
	}

	if err := remotes.PushContent(ctx, pusher, desc, store, nil, wrapper); err != nil {
		return ocispec.Descriptor{}, err
	}
	return desc, nil
}

//func pack(store *hybridStore, descriptors []ocispec.Descriptor, opts *pushOpts) (ocispec.Descriptor, error) {
func pack(provider content.Provider, descriptors []ocispec.Descriptor, opts *pushOpts) (ocispec.Descriptor, content.Store, error) {
	store := newHybridStoreFromProvider(provider, nil)

	// Config
	var config ocispec.Descriptor
	if opts.config == nil {
		configBytes := []byte("{}")
		config = ocispec.Descriptor{
			MediaType: artifact.UnknownConfigMediaType,
			Digest:    digest.FromBytes(configBytes),
			Size:      int64(len(configBytes)),
		}
		store.Set(config, configBytes)
	} else {
		config = *opts.config
	}
	if opts.configAnnotations != nil {
		config.Annotations = opts.configAnnotations
	}
	if opts.configMediaType != "" {
		config.MediaType = opts.configMediaType
	}

	// Manifest
	if opts.manifest != nil {
		return *opts.manifest, store, nil
	}

	if descriptors == nil {
		descriptors = []ocispec.Descriptor{} // make it an empty array to prevent potential server-side bugs
	}
	manifest := ocispec.Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2, // historical value. does not pertain to OCI or docker version
		},
		Config:      config,
		Layers:      descriptors,
		Annotations: opts.manifestAnnotations,
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return ocispec.Descriptor{}, nil, err
	}
	manifestDescriptor := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}
	store.Set(manifestDescriptor, manifestBytes)

	return manifestDescriptor, store, nil
}
