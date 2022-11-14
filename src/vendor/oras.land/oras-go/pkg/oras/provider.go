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
	"errors"
	"io"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/remotes"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// ProviderWrapper wraps a remote.Fetcher to make a content.Provider, which is useful for things
type ProviderWrapper struct {
	Fetcher remotes.Fetcher
}

func (p *ProviderWrapper) ReaderAt(ctx context.Context, desc ocispec.Descriptor) (content.ReaderAt, error) {
	if p.Fetcher == nil {
		return nil, errors.New("no Fetcher provided")
	}
	return &fetcherReaderAt{
		ctx:     ctx,
		fetcher: p.Fetcher,
		desc:    desc,
		offset:  0,
	}, nil
}

type fetcherReaderAt struct {
	ctx     context.Context
	fetcher remotes.Fetcher
	desc    ocispec.Descriptor
	rc      io.ReadCloser
	offset  int64
}

func (f *fetcherReaderAt) Close() error {
	if f.rc == nil {
		return nil
	}
	return f.rc.Close()
}

func (f *fetcherReaderAt) Size() int64 {
	return f.desc.Size
}

func (f *fetcherReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	// if we do not have a readcloser, get it
	if f.rc == nil || f.offset != off {
		rc, err := f.fetcher.Fetch(f.ctx, f.desc)
		if err != nil {
			return 0, err
		}
		f.rc = rc
	}

	n, err = io.ReadFull(f.rc, p)
	if err != nil {
		return n, err
	}
	f.offset += int64(n)
	return n, err
}
