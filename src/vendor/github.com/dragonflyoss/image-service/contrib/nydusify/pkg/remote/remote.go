// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package remote

import (
	"context"
	"io"
	"sync"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/remotes"
	"github.com/docker/distribution/reference"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Remote provides the ability to access remote registry
type Remote struct {
	Ref      string
	parsed   reference.Named
	resolver remotes.Resolver
	pushed   sync.Map
}

// New creates remote instance from docker remote resolver
func New(ref string, resolver remotes.Resolver) (*Remote, error) {
	parsed, err := reference.ParseNormalizedNamed(ref)
	if err != nil {
		return nil, err
	}

	return &Remote{
		Ref:      ref,
		parsed:   parsed,
		resolver: resolver,
	}, nil
}

// Push pushes blob to registry
func (remote *Remote) Push(ctx context.Context, desc ocispec.Descriptor, byDigest bool, reader io.Reader) error {
	// Concurrently push blob with same digest using containerd
	// docker remote client will cause error:
	// `failed commit on ref: unexpected size x, expected y`
	// use ref key leveled mutex lock to avoid the issue.
	refKey := remotes.MakeRefKey(ctx, desc)
	lock, _ := remote.pushed.LoadOrStore(refKey, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()

	var ref string
	if byDigest {
		ref = remote.parsed.Name()
	} else {
		ref = reference.TagNameOnly(remote.parsed).String()
	}

	pusher, err := remote.resolver.Pusher(ctx, ref)
	if err != nil {
		return err
	}

	writer, err := pusher.Push(ctx, desc)
	if err != nil {
		if errdefs.IsAlreadyExists(err) {
			return nil
		}
		return err
	}
	defer writer.Close()

	return content.Copy(ctx, writer, reader, desc.Size, desc.Digest)
}

// Pull pulls blob from registry
func (remote *Remote) Pull(ctx context.Context, desc ocispec.Descriptor, byDigest bool) (io.ReadCloser, error) {
	var ref string
	if byDigest {
		ref = remote.parsed.Name()
	} else {
		ref = reference.TagNameOnly(remote.parsed).String()
	}

	puller, err := remote.resolver.Fetcher(ctx, ref)
	if err != nil {
		return nil, err
	}

	reader, err := puller.Fetch(ctx, desc)
	if err != nil {
		return nil, err
	}

	return reader, nil
}

// Resolve parses descriptor for given image reference
func (remote *Remote) Resolve(ctx context.Context) (*ocispec.Descriptor, error) {
	ref := reference.TagNameOnly(remote.parsed).String()

	_, desc, err := remote.resolver.Resolve(ctx, ref)
	if err != nil {
		return nil, err
	}

	return &desc, nil
}
