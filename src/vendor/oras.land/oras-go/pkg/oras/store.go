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
	"io"
	"io/ioutil"
	"time"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/remotes"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"

	orascontent "oras.land/oras-go/pkg/content"
)

type hybridStore struct {
	cache            *orascontent.Memory
	cachedMediaTypes []string
	cacheOnly        bool
	provider         content.Provider
	ingester         content.Ingester
}

func newHybridStoreFromPusher(pusher remotes.Pusher, cachedMediaTypes []string, cacheOnly bool) *hybridStore {
	// construct an ingester from a pusher
	ingester := pusherIngester{
		pusher: pusher,
	}
	return &hybridStore{
		cache:            orascontent.NewMemory(),
		cachedMediaTypes: cachedMediaTypes,
		ingester:         ingester,
		cacheOnly:        cacheOnly,
	}
}

func (s *hybridStore) Set(desc ocispec.Descriptor, content []byte) {
	s.cache.Set(desc, content)
}

func (s *hybridStore) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	reader, err := s.cache.Fetch(ctx, desc)
	if err == nil {
		return reader, err
	}
	if s.provider != nil {
		rat, err := s.provider.ReaderAt(ctx, desc)
		return ioutil.NopCloser(orascontent.NewReaderAtWrapper(rat)), err
	}
	return nil, err
}

func (s *hybridStore) Push(ctx context.Context, desc ocispec.Descriptor) (content.Writer, error) {
	return s.Writer(ctx, content.WithDescriptor(desc))
}

// Writer begins or resumes the active writer identified by desc
func (s *hybridStore) Writer(ctx context.Context, opts ...content.WriterOpt) (content.Writer, error) {
	var wOpts content.WriterOpts
	for _, opt := range opts {
		if err := opt(&wOpts); err != nil {
			return nil, err
		}
	}

	if isAllowedMediaType(wOpts.Desc.MediaType, s.cachedMediaTypes...) || s.ingester == nil {
		pusher, err := s.cache.Pusher(ctx, "")
		if err != nil {
			return nil, err
		}
		cacheWriter, err := pusher.Push(ctx, wOpts.Desc)
		if err != nil {
			return nil, err
		}
		// if we cache it only, do not pass it through
		if s.cacheOnly {
			return cacheWriter, nil
		}
		ingesterWriter, err := s.ingester.Writer(ctx, opts...)
		switch {
		case err == nil:
			return newTeeWriter(wOpts.Desc, cacheWriter, ingesterWriter), nil
		case errdefs.IsAlreadyExists(err):
			return cacheWriter, nil
		}
		return nil, err
	}
	return s.ingester.Writer(ctx, opts...)
}

// teeWriter tees the content to one or more content.Writer
type teeWriter struct {
	writers  []content.Writer
	digester digest.Digester
	status   content.Status
}

func newTeeWriter(desc ocispec.Descriptor, writers ...content.Writer) *teeWriter {
	now := time.Now()
	return &teeWriter{
		writers:  writers,
		digester: digest.Canonical.Digester(),
		status: content.Status{
			Total:     desc.Size,
			StartedAt: now,
			UpdatedAt: now,
		},
	}
}

func (t *teeWriter) Close() error {
	g := new(errgroup.Group)
	for _, w := range t.writers {
		w := w // closure issues, see https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			return w.Close()
		})
	}
	return g.Wait()
}

func (t *teeWriter) Write(p []byte) (n int, err error) {
	g := new(errgroup.Group)
	for _, w := range t.writers {
		w := w // closure issues, see https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			n, err := w.Write(p[:])
			if err != nil {
				return err
			}
			if n != len(p) {
				return io.ErrShortWrite
			}
			return nil
		})
	}
	err = g.Wait()
	n = len(p)
	if err != nil {
		return n, err
	}
	_, _ = t.digester.Hash().Write(p[:n])
	t.status.Offset += int64(len(p))
	t.status.UpdatedAt = time.Now()

	return n, nil
}

// Digest may return empty digest or panics until committed.
func (t *teeWriter) Digest() digest.Digest {
	return t.digester.Digest()
}

func (t *teeWriter) Commit(ctx context.Context, size int64, expected digest.Digest, opts ...content.Opt) error {
	g := new(errgroup.Group)
	for _, w := range t.writers {
		w := w // closure issues, see https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			return w.Commit(ctx, size, expected, opts...)
		})
	}
	return g.Wait()
}

// Status returns the current state of write
func (t *teeWriter) Status() (content.Status, error) {
	return t.status, nil
}

// Truncate updates the size of the target blob
func (t *teeWriter) Truncate(size int64) error {
	g := new(errgroup.Group)
	for _, w := range t.writers {
		w := w // closure issues, see https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			return w.Truncate(size)
		})
	}
	return g.Wait()
}

// pusherIngester simple wrapper to get an ingester from a pusher
type pusherIngester struct {
	pusher remotes.Pusher
}

func (p pusherIngester) Writer(ctx context.Context, opts ...content.WriterOpt) (content.Writer, error) {
	var wOpts content.WriterOpts
	for _, opt := range opts {
		if err := opt(&wOpts); err != nil {
			return nil, err
		}
	}
	return p.pusher.Push(ctx, wOpts.Desc)
}
