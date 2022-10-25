/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package oras

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/remotes"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/pkg/target"
)

// Copy copy a ref from one target.Target to a ref in another target.Target. If toRef is blank, reuses fromRef
// Returns the root
// Descriptor of the copied item. Can use the root to retrieve child elements from target.Target.
func Copy(ctx context.Context, from target.Target, fromRef string, to target.Target, toRef string, opts ...CopyOpt) (ocispec.Descriptor, error) {
	if from == nil {
		return ocispec.Descriptor{}, ErrFromTargetUndefined
	}
	if to == nil {
		return ocispec.Descriptor{}, ErrToTargetUndefined
	}
	// blank toRef
	if toRef == "" {
		toRef = fromRef
	}
	opt := copyOptsDefaults()
	for _, o := range opts {
		if err := o(opt); err != nil {
			return ocispec.Descriptor{}, err
		}
	}

	if from == nil {
		return ocispec.Descriptor{}, ErrFromResolverUndefined
	}
	if to == nil {
		return ocispec.Descriptor{}, ErrToResolverUndefined
	}

	// for the "from", we resolve the ref, then use resolver.Fetcher to fetch the various content blobs
	// for the "to", we simply use resolver.Pusher to push the various content blobs

	_, desc, err := from.Resolve(ctx, fromRef)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	fetcher, err := from.Fetcher(ctx, fromRef)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	// construct the reference we send to the pusher using the digest, so it knows what the root is
	pushRef := fmt.Sprintf("%s@%s", toRef, desc.Digest.String())
	pusher, err := to.Pusher(ctx, pushRef)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	if err := transferContent(ctx, desc, fetcher, pusher, opt); err != nil {
		return ocispec.Descriptor{}, err
	}
	return desc, nil
}

func transferContent(ctx context.Context, desc ocispec.Descriptor, fetcher remotes.Fetcher, pusher remotes.Pusher, opts *copyOpts) error {
	var descriptors, manifests []ocispec.Descriptor
	lock := &sync.Mutex{}
	picker := images.HandlerFunc(func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		if isAllowedMediaType(desc.MediaType, opts.allowedMediaTypes...) {
			if opts.filterName(desc) {
				lock.Lock()
				defer lock.Unlock()
				descriptors = append(descriptors, desc)
			}
			return nil, nil
		}
		return nil, nil
	})

	// we use a hybrid store - a cache wrapping the underlying pusher - for two reasons:
	// 1. so that we can cache the manifests as pushing them, then retrieve them later to push in reverse order after the blobs
	// 2. so that we can retrieve them to analyze and find children in the Dispatch routine
	store := opts.contentProvideIngesterPusherFetcher
	if store == nil {
		store = newHybridStoreFromPusher(pusher, opts.cachedMediaTypes, true)
	}

	// fetchHandler pushes to the *store*, which may or may not cache it
	baseFetchHandler := func(p remotes.Pusher, f remotes.Fetcher) images.HandlerFunc {
		return images.HandlerFunc(func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
			cw, err := p.Push(ctx, desc)
			if err != nil {
				if !errdefs.IsAlreadyExists(err) {
					return nil, err
				}

				return nil, nil
			}
			defer cw.Close()

			rc, err := f.Fetch(ctx, desc)
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return nil, content.Copy(ctx, cw, rc, desc.Size, desc.Digest)
		})
	}

	// track all of our manifests that will be cached
	fetchHandler := images.HandlerFunc(func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		if isAllowedMediaType(desc.MediaType, opts.cachedMediaTypes...) {
			lock.Lock()
			defer lock.Unlock()
			manifests = append(manifests, desc)
		}
		return baseFetchHandler(store, fetcher)(ctx, desc)
	})

	handlers := []images.Handler{
		filterHandler(opts, opts.allowedMediaTypes...),
	}
	handlers = append(handlers, opts.baseHandlers...)
	handlers = append(handlers,
		fetchHandler,
		picker,
		images.ChildrenHandler(&ProviderWrapper{Fetcher: store}),
	)
	handlers = append(handlers, opts.callbackHandlers...)

	if err := opts.dispatch(ctx, images.Handlers(handlers...), nil, desc); err != nil {
		return err
	}

	// we cached all of the manifests, so push those out
	// Iterate in reverse order as seen, parent always uploaded after child
	for i := len(manifests) - 1; i >= 0; i-- {
		_, err := baseFetchHandler(pusher, store)(ctx, manifests[i])
		if err != nil {
			return err
		}
	}

	// if the option to request the root manifest was passed, accommodate it
	if opts.saveManifest != nil && len(manifests) > 0 {
		rc, err := store.Fetch(ctx, manifests[0])
		if err != nil {
			return fmt.Errorf("could not get root manifest to save based on CopyOpt: %v", err)
		}
		defer rc.Close()
		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(rc); err != nil {
			return fmt.Errorf("unable to read data for root manifest to save based on CopyOpt: %v", err)
		}
		// get the root manifest from the store
		opts.saveManifest(buf.Bytes())
	}

	// if the option to request the layers was passed, accommodate it
	if opts.saveLayers != nil && len(descriptors) > 0 {
		opts.saveLayers(descriptors)
	}
	return nil
}

func filterHandler(opts *copyOpts, allowedMediaTypes ...string) images.HandlerFunc {
	return func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		switch {
		case isAllowedMediaType(desc.MediaType, ocispec.MediaTypeImageManifest, ocispec.MediaTypeImageIndex):
			return nil, nil
		case isAllowedMediaType(desc.MediaType, allowedMediaTypes...):
			if opts.filterName(desc) {
				return nil, nil
			}
			log.G(ctx).Warnf("blob no name: %v", desc.Digest)
		default:
			log.G(ctx).Warnf("unknown type: %v", desc.MediaType)
		}
		return nil, images.ErrStopHandler
	}
}

func isAllowedMediaType(mediaType string, allowedMediaTypes ...string) bool {
	if len(allowedMediaTypes) == 0 {
		return true
	}
	for _, allowedMediaType := range allowedMediaTypes {
		if mediaType == allowedMediaType {
			return true
		}
	}
	return false
}
