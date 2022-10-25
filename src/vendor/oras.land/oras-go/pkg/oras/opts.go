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
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"github.com/containerd/containerd/images"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
	orascontent "oras.land/oras-go/pkg/content"
)

func copyOptsDefaults() *copyOpts {
	return &copyOpts{
		dispatch:         images.Dispatch,
		filterName:       filterName,
		cachedMediaTypes: []string{ocispec.MediaTypeImageManifest, ocispec.MediaTypeImageIndex},
		validateName:     ValidateNameAsPath,
	}
}

type CopyOpt func(o *copyOpts) error

type copyOpts struct {
	allowedMediaTypes                   []string
	dispatch                            func(context.Context, images.Handler, *semaphore.Weighted, ...ocispec.Descriptor) error
	baseHandlers                        []images.Handler
	callbackHandlers                    []images.Handler
	contentProvideIngesterPusherFetcher orascontent.Store
	filterName                          func(ocispec.Descriptor) bool
	cachedMediaTypes                    []string

	saveManifest func([]byte)
	saveLayers   func([]ocispec.Descriptor)
	validateName func(desc ocispec.Descriptor) error

	userAgent string
}

// ValidateNameAsPath validates name in the descriptor as file path in order
// to generate good packages intended to be pulled using the FileStore or
// the oras cli.
// For cross-platform considerations, only unix paths are accepted.
func ValidateNameAsPath(desc ocispec.Descriptor) error {
	// no empty name
	path, ok := orascontent.ResolveName(desc)
	if !ok || path == "" {
		return orascontent.ErrNoName
	}

	// path should be clean
	if target := filepath.ToSlash(filepath.Clean(path)); target != path {
		return errors.Wrap(ErrDirtyPath, path)
	}

	// path should be slash-separated
	if strings.Contains(path, "\\") {
		return errors.Wrap(ErrPathNotSlashSeparated, path)
	}

	// disallow absolute path: covers unix and windows format
	if strings.HasPrefix(path, "/") {
		return errors.Wrap(ErrAbsolutePathDisallowed, path)
	}
	if len(path) > 2 {
		c := path[0]
		if path[1] == ':' && path[2] == '/' && ('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z') {
			return errors.Wrap(ErrAbsolutePathDisallowed, path)
		}
	}

	// disallow path traversal
	if strings.HasPrefix(path, "../") || path == ".." {
		return errors.Wrap(ErrPathTraversalDisallowed, path)
	}

	return nil
}

// dispatchBFS behaves the same as images.Dispatch() but in sequence with breath-first search.
func dispatchBFS(ctx context.Context, handler images.Handler, weighted *semaphore.Weighted, descs ...ocispec.Descriptor) error {
	for i := 0; i < len(descs); i++ {
		desc := descs[i]
		children, err := handler.Handle(ctx, desc)
		if err != nil {
			switch err := errors.Cause(err); err {
			case images.ErrSkipDesc:
				continue // don't traverse the children.
			case ErrStopProcessing:
				return nil
			}
			return err
		}
		descs = append(descs, children...)
	}
	return nil
}

func filterName(desc ocispec.Descriptor) bool {
	// needs to be filled in
	return true
}

// WithAdditionalCachedMediaTypes adds media types normally cached in memory when pulling.
// This does not replace the default media types, but appends to them
func WithAdditionalCachedMediaTypes(cachedMediaTypes ...string) CopyOpt {
	return func(o *copyOpts) error {
		o.cachedMediaTypes = append(o.cachedMediaTypes, cachedMediaTypes...)
		return nil
	}
}

// WithAllowedMediaType sets the allowed media types
func WithAllowedMediaType(allowedMediaTypes ...string) CopyOpt {
	return func(o *copyOpts) error {
		o.allowedMediaTypes = append(o.allowedMediaTypes, allowedMediaTypes...)
		return nil
	}
}

// WithAllowedMediaTypes sets the allowed media types
func WithAllowedMediaTypes(allowedMediaTypes []string) CopyOpt {
	return func(o *copyOpts) error {
		o.allowedMediaTypes = append(o.allowedMediaTypes, allowedMediaTypes...)
		return nil
	}
}

// WithPullByBFS opt to pull in sequence with breath-first search
func WithPullByBFS(o *copyOpts) error {
	o.dispatch = dispatchBFS
	return nil
}

// WithPullBaseHandler provides base handlers, which will be called before
// any pull specific handlers.
func WithPullBaseHandler(handlers ...images.Handler) CopyOpt {
	return func(o *copyOpts) error {
		o.baseHandlers = append(o.baseHandlers, handlers...)
		return nil
	}
}

// WithPullCallbackHandler provides callback handlers, which will be called after
// any pull specific handlers.
func WithPullCallbackHandler(handlers ...images.Handler) CopyOpt {
	return func(o *copyOpts) error {
		o.callbackHandlers = append(o.callbackHandlers, handlers...)
		return nil
	}
}

// WithContentProvideIngester opt to the provided Provider and Ingester
// for file system I/O, including caches.
func WithContentStore(store orascontent.Store) CopyOpt {
	return func(o *copyOpts) error {
		o.contentProvideIngesterPusherFetcher = store
		return nil
	}
}

// WithPullEmptyNameAllowed allows pulling blobs with empty name.
func WithPullEmptyNameAllowed() CopyOpt {
	return func(o *copyOpts) error {
		o.filterName = func(ocispec.Descriptor) bool {
			return true
		}
		return nil
	}
}

// WithPullStatusTrack report results to stdout
func WithPullStatusTrack(writer io.Writer) CopyOpt {
	return WithPullCallbackHandler(pullStatusTrack(writer))
}

func pullStatusTrack(writer io.Writer) images.Handler {
	var printLock sync.Mutex
	return images.HandlerFunc(func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		if name, ok := orascontent.ResolveName(desc); ok {
			digestString := desc.Digest.String()
			if err := desc.Digest.Validate(); err == nil {
				if algo := desc.Digest.Algorithm(); algo == digest.SHA256 {
					digestString = desc.Digest.Encoded()[:12]
				}
			}
			printLock.Lock()
			defer printLock.Unlock()
			fmt.Fprintln(writer, "Downloaded", digestString, name)
		}
		return nil, nil
	})
}

// WithNameValidation validates the image title in the descriptor.
// Pass nil to disable name validation.
func WithNameValidation(validate func(desc ocispec.Descriptor) error) CopyOpt {
	return func(o *copyOpts) error {
		o.validateName = validate
		return nil
	}
}

// WithUserAgent set the user agent string in http communications
func WithUserAgent(agent string) CopyOpt {
	return func(o *copyOpts) error {
		o.userAgent = agent
		return nil
	}
}

// WithLayerDescriptors passes the slice of Descriptors for layers to the
// provided func. If the passed parameter is nil, returns an error.
func WithLayerDescriptors(save func([]ocispec.Descriptor)) CopyOpt {
	return func(o *copyOpts) error {
		if save == nil {
			return errors.New("layers save func must be non-nil")
		}
		o.saveLayers = save
		return nil
	}
}

// WithRootManifest passes the root manifest for the artifacts to the provided
// func. If the passed parameter is nil, returns an error.
func WithRootManifest(save func(b []byte)) CopyOpt {
	return func(o *copyOpts) error {
		if save == nil {
			return errors.New("manifest save func must be non-nil")
		}
		o.saveManifest = save
		return nil
	}
}
