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
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/remotes"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

// Memory provides content from the memory
type Memory struct {
	descriptor map[digest.Digest]ocispec.Descriptor
	content    map[digest.Digest][]byte
	nameMap    map[string]ocispec.Descriptor
	refMap     map[string]ocispec.Descriptor
	lock       *sync.Mutex
}

// NewMemory creats a new memory store
func NewMemory() *Memory {
	return &Memory{
		descriptor: make(map[digest.Digest]ocispec.Descriptor),
		content:    make(map[digest.Digest][]byte),
		nameMap:    make(map[string]ocispec.Descriptor),
		refMap:     make(map[string]ocispec.Descriptor),
		lock:       &sync.Mutex{},
	}
}

func (s *Memory) Resolver() remotes.Resolver {
	return s
}

func (s *Memory) Resolve(ctx context.Context, ref string) (name string, desc ocispec.Descriptor, err error) {
	desc, ok := s.refMap[ref]
	if !ok {
		return "", ocispec.Descriptor{}, fmt.Errorf("unknown reference: %s", ref)
	}
	return ref, desc, nil
}

func (s *Memory) Fetcher(ctx context.Context, ref string) (remotes.Fetcher, error) {
	if _, ok := s.refMap[ref]; !ok {
		return nil, fmt.Errorf("unknown reference: %s", ref)
	}
	return s, nil
}

// Fetch get an io.ReadCloser for the specific content
func (s *Memory) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	_, content, ok := s.Get(desc)
	if !ok {
		return nil, ErrNotFound
	}
	return ioutil.NopCloser(bytes.NewReader(content)), nil
}

func (s *Memory) Pusher(ctx context.Context, ref string) (remotes.Pusher, error) {
	var tag, hash string
	parts := strings.SplitN(ref, "@", 2)
	if len(parts) > 0 {
		tag = parts[0]
	}
	if len(parts) > 1 {
		hash = parts[1]
	}
	return &memoryPusher{
		store: s,
		ref:   tag,
		hash:  hash,
	}, nil
}

type memoryPusher struct {
	store *Memory
	ref   string
	hash  string
}

func (s *memoryPusher) Push(ctx context.Context, desc ocispec.Descriptor) (content.Writer, error) {
	name, _ := ResolveName(desc)
	now := time.Now()
	// is this the root?
	if desc.Digest.String() == s.hash {
		s.store.refMap[s.ref] = desc
	}
	return &memoryWriter{
		store:    s.store,
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

// Add adds content, generating a descriptor and returning it.
func (s *Memory) Add(name, mediaType string, content []byte) (ocispec.Descriptor, error) {
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
	return desc, nil
}

// Set adds the content to the store
func (s *Memory) Set(desc ocispec.Descriptor, content []byte) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.descriptor[desc.Digest] = desc
	s.content[desc.Digest] = content

	if name, ok := ResolveName(desc); ok && name != "" {
		s.nameMap[name] = desc
	}
}

// Get finds the content from the store
func (s *Memory) Get(desc ocispec.Descriptor) (ocispec.Descriptor, []byte, bool) {
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
func (s *Memory) GetByName(name string) (ocispec.Descriptor, []byte, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	desc, ok := s.nameMap[name]
	if !ok {
		return ocispec.Descriptor{}, nil, false
	}
	content, ok := s.content[desc.Digest]
	return desc, content, ok
}

// StoreManifest stores a manifest linked to by the provided ref. The children of the
// manifest, such as layers and config, should already exist in the file store, either
// as files linked via Add(), or via Set(). If they do not exist, then a typical
// Fetcher that walks the manifest will hit an unresolved hash.
//
// StoreManifest does *not* validate their presence.
func (s *Memory) StoreManifest(ref string, desc ocispec.Descriptor, manifest []byte) error {
	s.refMap[ref] = desc
	s.Add("", desc.MediaType, manifest)
	return nil
}

func descFromBytes(b []byte, mediaType string) (ocispec.Descriptor, error) {
	digest, err := digest.FromReader(bytes.NewReader(b))
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	if mediaType == "" {
		mediaType = DefaultBlobMediaType
	}
	return ocispec.Descriptor{
		MediaType: mediaType,
		Digest:    digest,
		Size:      int64(len(b)),
	}, nil
}

type memoryWriter struct {
	store    *Memory
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
