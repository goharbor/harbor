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

package content

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

// ensure interface
var (
	_ content.Provider = &Memorystore{}
	_ content.Ingester = &Memorystore{}
)

// Memorystore provides content from the memory
type Memorystore struct {
	descriptor map[digest.Digest]ocispec.Descriptor
	content    map[digest.Digest][]byte
	nameMap    map[string]ocispec.Descriptor
	lock       *sync.Mutex
}

// NewMemoryStore creats a new memory store
func NewMemoryStore() *Memorystore {
	return &Memorystore{
		descriptor: make(map[digest.Digest]ocispec.Descriptor),
		content:    make(map[digest.Digest][]byte),
		nameMap:    make(map[string]ocispec.Descriptor),
		lock:       &sync.Mutex{},
	}
}

// Add adds content
func (s *Memorystore) Add(name, mediaType string, content []byte) ocispec.Descriptor {
	var annotations map[string]string
	if name != "" {
		annotations = map[string]string{
			ocispec.AnnotationTitle: name,
		}
	}

	if mediaType == "" {
		mediaType = DefaultBlobMediaType
	}

	desc := ocispec.Descriptor{
		MediaType:   mediaType,
		Digest:      digest.FromBytes(content),
		Size:        int64(len(content)),
		Annotations: annotations,
	}

	s.Set(desc, content)
	return desc
}

// ReaderAt provides contents
func (s *Memorystore) ReaderAt(ctx context.Context, desc ocispec.Descriptor) (content.ReaderAt, error) {
	desc, content, ok := s.Get(desc)
	if !ok {
		return nil, ErrNotFound
	}

	return sizeReaderAt{
		readAtCloser: nopCloser{
			ReaderAt: bytes.NewReader(content),
		},
		size: desc.Size,
	}, nil
}

// Writer begins or resumes the active writer identified by desc
func (s *Memorystore) Writer(ctx context.Context, opts ...content.WriterOpt) (content.Writer, error) {
	var wOpts content.WriterOpts
	for _, opt := range opts {
		if err := opt(&wOpts); err != nil {
			return nil, err
		}
	}
	desc := wOpts.Desc

	name, _ := ResolveName(desc)
	now := time.Now()
	return &memoryWriter{
		store:    s,
		buffer:   bytes.NewBuffer(nil),
		desc:     desc,
		digester: digest.Canonical.Digester(),
		status: content.Status{
			Ref:       name,
			Total:     desc.Size,
			StartedAt: now,
			UpdatedAt: now,
		},
	}, nil
}

// Set adds the content to the store
func (s *Memorystore) Set(desc ocispec.Descriptor, content []byte) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.descriptor[desc.Digest] = desc
	s.content[desc.Digest] = content

	if name, ok := ResolveName(desc); ok && name != "" {
		s.nameMap[name] = desc
	}
}

// Get finds the content from the store
func (s *Memorystore) Get(desc ocispec.Descriptor) (ocispec.Descriptor, []byte, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	desc, ok := s.descriptor[desc.Digest]
	if !ok {
		return ocispec.Descriptor{}, nil, false
	}
	content, ok := s.content[desc.Digest]
	return desc, content, ok
}

// GetByName finds the content from the store by name (i.e. AnnotationTitle)
func (s *Memorystore) GetByName(name string) (ocispec.Descriptor, []byte, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	desc, ok := s.nameMap[name]
	if !ok {
		return ocispec.Descriptor{}, nil, false
	}
	content, ok := s.content[desc.Digest]
	return desc, content, ok
}

type memoryWriter struct {
	store    *Memorystore
	buffer   *bytes.Buffer
	desc     ocispec.Descriptor
	digester digest.Digester
	status   content.Status
}

func (w *memoryWriter) Status() (content.Status, error) {
	return w.status, nil
}

// Digest returns the current digest of the content, up to the current write.
//
// Cannot be called concurrently with `Write`.
func (w *memoryWriter) Digest() digest.Digest {
	return w.digester.Digest()
}

// Write p to the transaction.
func (w *memoryWriter) Write(p []byte) (n int, err error) {
	n, err = w.buffer.Write(p)
	w.digester.Hash().Write(p[:n])
	w.status.Offset += int64(len(p))
	w.status.UpdatedAt = time.Now()
	return n, err
}

func (w *memoryWriter) Commit(ctx context.Context, size int64, expected digest.Digest, opts ...content.Opt) error {
	var base content.Info
	for _, opt := range opts {
		if err := opt(&base); err != nil {
			return err
		}
	}

	if w.buffer == nil {
		return errors.Wrap(errdefs.ErrFailedPrecondition, "cannot commit on closed writer")
	}
	content := w.buffer.Bytes()
	w.buffer = nil

	if size > 0 && size != int64(len(content)) {
		return errors.Wrapf(errdefs.ErrFailedPrecondition, "unexpected commit size %d, expected %d", len(content), size)
	}
	if dgst := w.digester.Digest(); expected != "" && expected != dgst {
		return errors.Wrapf(errdefs.ErrFailedPrecondition, "unexpected commit digest %s, expected %s", dgst, expected)
	}

	w.store.Set(w.desc, content)
	return nil
}

func (w *memoryWriter) Close() error {
	w.buffer = nil
	return nil
}

func (w *memoryWriter) Truncate(size int64) error {
	if size != 0 {
		return ErrUnsupportedSize
	}
	w.status.Offset = 0
	w.digester.Hash().Reset()
	w.buffer.Truncate(0)
	return nil
}
