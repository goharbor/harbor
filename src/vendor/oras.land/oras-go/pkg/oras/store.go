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
	"time"

	"github.com/containerd/containerd/content"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"

	orascontent "oras.land/oras-go/pkg/content"
)

// ensure interface
var (
	_ content.Store = &hybridStore{}
)

type hybridStore struct {
	cache            *orascontent.Memorystore
	cachedMediaTypes []string
	provider         content.Provider
	ingester         content.Ingester
}

func newHybridStoreFromProvider(provider content.Provider, cachedMediaTypes []string) *hybridStore {
	return &hybridStore{
		cache:            orascontent.NewMemoryStore(),
		cachedMediaTypes: cachedMediaTypes,
		provider:         provider,
	}
}

func newHybridStoreFromIngester(ingester content.Ingester, cachedMediaTypes []string) *hybridStore {
	return &hybridStore{
		cache:            orascontent.NewMemoryStore(),
		cachedMediaTypes: cachedMediaTypes,
		ingester:         ingester,
	}
}

func (s *hybridStore) Set(desc ocispec.Descriptor, content []byte) {
	s.cache.Set(desc, content)
}

// ReaderAt provides contents
func (s *hybridStore) ReaderAt(ctx context.Context, desc ocispec.Descriptor) (content.ReaderAt, error) {
	readerAt, err := s.cache.ReaderAt(ctx, desc)
	if err == nil {
		return readerAt, nil
	}
	if s.provider != nil {
		return s.provider.ReaderAt(ctx, desc)
	}
	return nil, err
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
		cacheWriter, err := s.cache.Writer(ctx, opts...)
		if err != nil {
			return nil, err
		}
		ingesterWriter, err := s.ingester.Writer(ctx, opts...)
		if err != nil {
			return nil, err
		}
		return newTeeWriter(wOpts.Desc, cacheWriter, ingesterWriter), nil
	}
	return s.ingester.Writer(ctx, opts...)
}

// TODO: implement (needed to create a content.Store)
// TODO: do not return empty content.Info
// Abort completely cancels the ingest operation targeted by ref.
func (s *hybridStore) Info(ctx context.Context, dgst digest.Digest) (content.Info, error) {
	return content.Info{}, nil
}

// TODO: implement (needed to create a content.Store)
// Update updates mutable information related to content.
// If one or more fieldpaths are provided, only those
// fields will be updated.
// Mutable fields:
//  labels.*
func (s *hybridStore) Update(ctx context.Context, info content.Info, fieldpaths ...string) (content.Info, error) {
	return content.Info{}, errors.New("not yet implemented: Update (content.Store interface)")
}

// TODO: implement (needed to create a content.Store)
// Walk will call fn for each item in the content store which
// match the provided filters. If no filters are given all
// items will be walked.
func (s *hybridStore) Walk(ctx context.Context, fn content.WalkFunc, filters ...string) error {
	return errors.New("not yet implemented: Walk (content.Store interface)")
}

// TODO: implement (needed to create a content.Store)
// Delete removes the content from the store.
func (s *hybridStore) Delete(ctx context.Context, dgst digest.Digest) error {
	return errors.New("not yet implemented: Delete (content.Store interface)")
}

// TODO: implement (needed to create a content.Store)
func (s *hybridStore) Status(ctx context.Context, ref string) (content.Status, error) {
	// Status returns the status of the provided ref.
	return content.Status{}, errors.New("not yet implemented: Status (content.Store interface)")
}

// TODO: implement (needed to create a content.Store)
// ListStatuses returns the status of any active ingestions whose ref match the
// provided regular expression. If empty, all active ingestions will be
// returned.
func (s *hybridStore) ListStatuses(ctx context.Context, filters ...string) ([]content.Status, error) {
	return []content.Status{}, errors.New("not yet implemented: ListStatuses (content.Store interface)")
}

// TODO: implement (needed to create a content.Store)
// Abort completely cancels the ingest operation targeted by ref.
func (s *hybridStore) Abort(ctx context.Context, ref string) error {
	return errors.New("not yet implemented: Abort (content.Store interface)")
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
