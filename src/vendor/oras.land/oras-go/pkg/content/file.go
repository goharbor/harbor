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
	"compress/gzip"
	"context"
	_ "crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

// File provides content via files from the file system
type File struct {
	DisableOverwrite          bool
	AllowPathTraversalOnWrite bool

	// Reproducible enables stripping times from added files
	Reproducible bool

	root         string
	descriptor   *sync.Map // map[digest.Digest]ocispec.Descriptor
	pathMap      *sync.Map // map[name string](file string)
	memoryMap    *sync.Map // map[digest.Digest]([]byte)
	refMap       *sync.Map // map[string]ocispec.Descriptor
	tmpFiles     *sync.Map
	ignoreNoName bool
}

// NewFile creats a new file target. It represents a single root reference and all of its components.
func NewFile(rootPath string, opts ...WriterOpt) *File {
	// we have to process the opts to find if they told us to change defaults
	wOpts := DefaultWriterOpts()
	for _, opt := range opts {
		if err := opt(&wOpts); err != nil {
			continue
		}
	}
	return &File{
		root:         rootPath,
		descriptor:   &sync.Map{},
		pathMap:      &sync.Map{},
		memoryMap:    &sync.Map{},
		refMap:       &sync.Map{},
		tmpFiles:     &sync.Map{},
		ignoreNoName: wOpts.IgnoreNoName,
	}
}

func (s *File) Resolver() remotes.Resolver {
	return s
}

func (s *File) Resolve(ctx context.Context, ref string) (name string, desc ocispec.Descriptor, err error) {
	desc, ok := s.getRef(ref)
	if !ok {
		return "", ocispec.Descriptor{}, fmt.Errorf("unknown reference: %s", ref)
	}
	return ref, desc, nil
}

func (s *File) Fetcher(ctx context.Context, ref string) (remotes.Fetcher, error) {
	if _, ok := s.refMap.Load(ref); !ok {
		return nil, fmt.Errorf("unknown reference: %s", ref)
	}
	return s, nil
}

// Fetch get an io.ReadCloser for the specific content
func (s *File) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	// first see if it is in the in-memory manifest map
	manifest, ok := s.getMemory(desc)
	if ok {
		return ioutil.NopCloser(bytes.NewReader(manifest)), nil
	}
	desc, ok = s.get(desc)
	if !ok {
		return nil, ErrNotFound
	}
	name, ok := ResolveName(desc)
	if !ok {
		return nil, ErrNoName
	}
	path := s.ResolvePath(name)
	return os.Open(path)
}

func (s *File) Pusher(ctx context.Context, ref string) (remotes.Pusher, error) {
	var tag, hash string
	parts := strings.SplitN(ref, "@", 2)
	if len(parts) > 0 {
		tag = parts[0]
	}
	if len(parts) > 1 {
		hash = parts[1]
	}
	return &filePusher{
		store: s,
		ref:   tag,
		hash:  hash,
	}, nil
}

type filePusher struct {
	store *File
	ref   string
	hash  string
}

func (s *filePusher) Push(ctx context.Context, desc ocispec.Descriptor) (content.Writer, error) {
	name, ok := ResolveName(desc)
	now := time.Now()
	if !ok {
		// if we were not told to ignore NoName, then return an error
		if !s.store.ignoreNoName {
			return nil, ErrNoName
		}

		// just return a nil writer - we do not want to calculate the hash, so just use
		// whatever was passed in the descriptor
		return NewIoContentWriter(ioutil.Discard, WithOutputHash(desc.Digest)), nil
	}
	path, err := s.store.resolveWritePath(name)
	if err != nil {
		return nil, err
	}
	file, afterCommit, err := s.store.createWritePath(path, desc, name)
	if err != nil {
		return nil, err
	}

	return &fileWriter{
		store:    s.store,
		file:     file,
		desc:     desc,
		digester: digest.Canonical.Digester(),
		status: content.Status{
			Ref:       name,
			Total:     desc.Size,
			StartedAt: now,
			UpdatedAt: now,
		},
		afterCommit: afterCommit,
	}, nil
}

// Add adds a file reference from a path, either directory or single file,
// and returns the reference descriptor.
func (s *File) Add(name, mediaType, path string) (ocispec.Descriptor, error) {
	if path == "" {
		path = name
	}
	path = s.MapPath(name, path)

	fileInfo, err := os.Stat(path)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	var desc ocispec.Descriptor
	if fileInfo.IsDir() {
		desc, err = s.descFromDir(name, mediaType, path)
	} else {
		desc, err = s.descFromFile(fileInfo, mediaType, path)
	}
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	if desc.Annotations == nil {
		desc.Annotations = make(map[string]string)
	}
	desc.Annotations[ocispec.AnnotationTitle] = name

	s.set(desc)
	return desc, nil
}

// Load is a lower-level memory-only version of Add. Rather than taking a path,
// generating a descriptor and creating a reference, it takes raw data and a descriptor
// that describes that data and stores it in memory. It will disappear at process
// termination.
//
// It is especially useful for adding ephemeral data, such as config, that must
// exist in order to walk a manifest.
func (s *File) Load(desc ocispec.Descriptor, data []byte) error {
	s.memoryMap.Store(desc.Digest, data)
	return nil
}

// Ref gets a reference's descriptor and content
func (s *File) Ref(ref string) (ocispec.Descriptor, []byte, error) {
	desc, ok := s.getRef(ref)
	if !ok {
		return ocispec.Descriptor{}, nil, ErrNotFound
	}
	// first see if it is in the in-memory manifest map
	manifest, ok := s.getMemory(desc)
	if !ok {
		return ocispec.Descriptor{}, nil, ErrNotFound
	}
	return desc, manifest, nil
}

func (s *File) descFromFile(info os.FileInfo, mediaType, path string) (ocispec.Descriptor, error) {
	file, err := os.Open(path)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	defer file.Close()
	digest, err := digest.FromReader(file)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	if mediaType == "" {
		mediaType = DefaultBlobMediaType
	}
	return ocispec.Descriptor{
		MediaType: mediaType,
		Digest:    digest,
		Size:      info.Size(),
	}, nil
}

func (s *File) descFromDir(name, mediaType, root string) (ocispec.Descriptor, error) {
	// generate temp file
	file, err := s.tempFile()
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	defer file.Close()
	s.MapPath(name, file.Name())

	// compress directory
	digester := digest.Canonical.Digester()
	zw := gzip.NewWriter(io.MultiWriter(file, digester.Hash()))
	defer zw.Close()
	tarDigester := digest.Canonical.Digester()
	if err := tarDirectory(root, name, io.MultiWriter(zw, tarDigester.Hash()), s.Reproducible); err != nil {
		return ocispec.Descriptor{}, err
	}

	// flush all
	if err := zw.Close(); err != nil {
		return ocispec.Descriptor{}, err
	}
	if err := file.Sync(); err != nil {
		return ocispec.Descriptor{}, err
	}

	// generate descriptor
	if mediaType == "" {
		mediaType = DefaultBlobDirMediaType
	}
	info, err := file.Stat()
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	return ocispec.Descriptor{
		MediaType: mediaType,
		Digest:    digester.Digest(),
		Size:      info.Size(),
		Annotations: map[string]string{
			AnnotationDigest: tarDigester.Digest().String(),
			AnnotationUnpack: "true",
		},
	}, nil
}

func (s *File) tempFile() (*os.File, error) {
	file, err := ioutil.TempFile("", TempFilePattern)
	if err != nil {
		return nil, err
	}
	s.tmpFiles.Store(file.Name(), file)
	return file, nil
}

// Close frees up resources used by the file store
func (s *File) Close() error {
	var errs []string
	s.tmpFiles.Range(func(name, _ interface{}) bool {
		if err := os.Remove(name.(string)); err != nil {
			errs = append(errs, err.Error())
		}
		return true
	})
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func (s *File) resolveWritePath(name string) (string, error) {
	path := s.ResolvePath(name)
	if !s.AllowPathTraversalOnWrite {
		base, err := filepath.Abs(s.root)
		if err != nil {
			return "", err
		}
		target, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		rel, err := filepath.Rel(base, target)
		if err != nil {
			return "", ErrPathTraversalDisallowed
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "../") || rel == ".." {
			return "", ErrPathTraversalDisallowed
		}
	}
	if s.DisableOverwrite {
		if _, err := os.Stat(path); err == nil {
			return "", ErrOverwriteDisallowed
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	return path, nil
}

func (s *File) createWritePath(path string, desc ocispec.Descriptor, prefix string) (*os.File, func() error, error) {
	if value, ok := desc.Annotations[AnnotationUnpack]; !ok || value != "true" {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, nil, err
		}
		file, err := os.Create(path)
		return file, nil, err
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, nil, err
	}
	file, err := s.tempFile()
	checksum := desc.Annotations[AnnotationDigest]
	afterCommit := func() error {
		return extractTarGzip(path, prefix, file.Name(), checksum)
	}
	return file, afterCommit, err
}

// MapPath maps name to path
func (s *File) MapPath(name, path string) string {
	path = s.resolvePath(path)
	s.pathMap.Store(name, path)
	return path
}

// ResolvePath returns the path by name
func (s *File) ResolvePath(name string) string {
	if value, ok := s.pathMap.Load(name); ok {
		if path, ok := value.(string); ok {
			return path
		}
	}

	// using the name as a fallback solution
	return s.resolvePath(name)
}

func (s *File) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(s.root, path)
}

func (s *File) set(desc ocispec.Descriptor) {
	s.descriptor.Store(desc.Digest, desc)
}

func (s *File) get(desc ocispec.Descriptor) (ocispec.Descriptor, bool) {
	value, ok := s.descriptor.Load(desc.Digest)
	if !ok {
		return ocispec.Descriptor{}, false
	}
	desc, ok = value.(ocispec.Descriptor)
	return desc, ok
}

func (s *File) getMemory(desc ocispec.Descriptor) ([]byte, bool) {
	value, ok := s.memoryMap.Load(desc.Digest)
	if !ok {
		return nil, false
	}
	content, ok := value.([]byte)
	return content, ok
}

func (s *File) getRef(ref string) (ocispec.Descriptor, bool) {
	value, ok := s.refMap.Load(ref)
	if !ok {
		return ocispec.Descriptor{}, false
	}
	desc, ok := value.(ocispec.Descriptor)
	return desc, ok
}

// StoreManifest stores a manifest linked to by the provided ref. The children of the
// manifest, such as layers and config, should already exist in the file store, either
// as files linked via Add(), or via Load(). If they do not exist, then a typical
// Fetcher that walks the manifest will hit an unresolved hash.
//
// StoreManifest does *not* validate their presence.
func (s *File) StoreManifest(ref string, desc ocispec.Descriptor, manifest []byte) error {
	s.refMap.Store(ref, desc)
	s.memoryMap.Store(desc.Digest, manifest)
	return nil
}

type fileWriter struct {
	store       *File
	file        *os.File
	desc        ocispec.Descriptor
	digester    digest.Digester
	status      content.Status
	afterCommit func() error
}

func (w *fileWriter) Status() (content.Status, error) {
	return w.status, nil
}

// Digest returns the current digest of the content, up to the current write.
//
// Cannot be called concurrently with `Write`.
func (w *fileWriter) Digest() digest.Digest {
	return w.digester.Digest()
}

// Write p to the transaction.
func (w *fileWriter) Write(p []byte) (n int, err error) {
	n, err = w.file.Write(p)
	w.digester.Hash().Write(p[:n])
	w.status.Offset += int64(len(p))
	w.status.UpdatedAt = time.Now()
	return n, err
}

func (w *fileWriter) Commit(ctx context.Context, size int64, expected digest.Digest, opts ...content.Opt) error {
	var base content.Info
	for _, opt := range opts {
		if err := opt(&base); err != nil {
			return err
		}
	}

	if w.file == nil {
		return errors.Wrap(errdefs.ErrFailedPrecondition, "cannot commit on closed writer")
	}
	file := w.file
	w.file = nil

	if err := file.Sync(); err != nil {
		file.Close()
		return errors.Wrap(err, "sync failed")
	}

	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return errors.Wrap(err, "stat failed")
	}
	if err := file.Close(); err != nil {
		return errors.Wrap(err, "failed to close file")
	}

	if size > 0 && size != fileInfo.Size() {
		return errors.Wrapf(errdefs.ErrFailedPrecondition, "unexpected commit size %d, expected %d", fileInfo.Size(), size)
	}
	if dgst := w.digester.Digest(); expected != "" && expected != dgst {
		return errors.Wrapf(errdefs.ErrFailedPrecondition, "unexpected commit digest %s, expected %s", dgst, expected)
	}

	w.store.set(w.desc)
	if w.afterCommit != nil {
		return w.afterCommit()
	}
	return nil
}

// Close the writer, flushing any unwritten data and leaving the progress in
// tact.
func (w *fileWriter) Close() error {
	if w.file == nil {
		return nil
	}

	w.file.Sync()
	err := w.file.Close()
	w.file = nil
	return err
}

func (w *fileWriter) Truncate(size int64) error {
	if size != 0 {
		return ErrUnsupportedSize
	}
	w.status.Offset = 0
	w.digester.Hash().Reset()
	if _, err := w.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return w.file.Truncate(0)
}
