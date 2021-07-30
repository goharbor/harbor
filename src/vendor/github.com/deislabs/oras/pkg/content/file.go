package content

import (
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
	_ ProvideIngester = &FileStore{}
)

// FileStore provides content from the file system
type FileStore struct {
	DisableOverwrite          bool
	AllowPathTraversalOnWrite bool

	// Reproducible enables stripping times from added files
	Reproducible bool

	root         string
	descriptor   *sync.Map // map[digest.Digest]ocispec.Descriptor
	pathMap      *sync.Map
	tmpFiles     *sync.Map
	ignoreNoName bool
}

// NewFileStore creats a new file store
func NewFileStore(rootPath string, opts ...WriterOpt) *FileStore {
	// we have to process the opts to find if they told us to change defaults
	wOpts := DefaultWriterOpts()
	for _, opt := range opts {
		if err := opt(&wOpts); err != nil {
			continue
		}
	}
	return &FileStore{
		root:         rootPath,
		descriptor:   &sync.Map{},
		pathMap:      &sync.Map{},
		tmpFiles:     &sync.Map{},
		ignoreNoName: wOpts.IgnoreNoName,
	}
}

// Add adds a file reference
func (s *FileStore) Add(name, mediaType, path string) (ocispec.Descriptor, error) {
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

func (s *FileStore) descFromFile(info os.FileInfo, mediaType, path string) (ocispec.Descriptor, error) {
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

func (s *FileStore) descFromDir(name, mediaType, root string) (ocispec.Descriptor, error) {
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

func (s *FileStore) tempFile() (*os.File, error) {
	file, err := ioutil.TempFile("", TempFilePattern)
	if err != nil {
		return nil, err
	}
	s.tmpFiles.Store(file.Name(), file)
	return file, nil
}

// Close frees up resources used by the file store
func (s *FileStore) Close() error {
	var errs []string
	s.tmpFiles.Range(func(name, _ interface{}) bool {
		if err := os.Remove(name.(string)); err != nil {
			errs = append(errs, err.Error())
		}
		return true
	})
	return errors.New(strings.Join(errs, "; "))
}

// ReaderAt provides contents
func (s *FileStore) ReaderAt(ctx context.Context, desc ocispec.Descriptor) (content.ReaderAt, error) {
	desc, ok := s.get(desc)
	if !ok {
		return nil, ErrNotFound
	}
	name, ok := ResolveName(desc)
	if !ok {
		return nil, ErrNoName
	}
	path := s.ResolvePath(name)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return sizeReaderAt{
		readAtCloser: file,
		size:         desc.Size,
	}, nil
}

// Writer begins or resumes the active writer identified by desc
func (s *FileStore) Writer(ctx context.Context, opts ...content.WriterOpt) (content.Writer, error) {
	var wOpts content.WriterOpts
	for _, opt := range opts {
		if err := opt(&wOpts); err != nil {
			return nil, err
		}
	}
	desc := wOpts.Desc

	name, ok := ResolveName(desc)
	if !ok {
		// if we were not told to ignore NoName, then return an error
		if !s.ignoreNoName {
			return nil, ErrNoName
		}

		// just return a nil writer - we do not want to calculate the hash, so just use
		// whatever was passed in the descriptor
		return NewIoContentWriter(ioutil.Discard, WithOutputHash(desc.Digest)), nil
	}
	path, err := s.resolveWritePath(name)
	if err != nil {
		return nil, err
	}
	file, afterCommit, err := s.createWritePath(path, desc, name)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &fileWriter{
		store:    s,
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

func (s *FileStore) resolveWritePath(name string) (string, error) {
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

func (s *FileStore) createWritePath(path string, desc ocispec.Descriptor, prefix string) (*os.File, func() error, error) {
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
func (s *FileStore) MapPath(name, path string) string {
	path = s.resolvePath(path)
	s.pathMap.Store(name, path)
	return path
}

// ResolvePath returns the path by name
func (s *FileStore) ResolvePath(name string) string {
	if value, ok := s.pathMap.Load(name); ok {
		if path, ok := value.(string); ok {
			return path
		}
	}

	// using the name as a fallback solution
	return s.resolvePath(name)
}

func (s *FileStore) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(s.root, path)
}

func (s *FileStore) set(desc ocispec.Descriptor) {
	s.descriptor.Store(desc.Digest, desc)
}

func (s *FileStore) get(desc ocispec.Descriptor) (ocispec.Descriptor, bool) {
	value, ok := s.descriptor.Load(desc.Digest)
	if !ok {
		return ocispec.Descriptor{}, false
	}
	desc, ok = value.(ocispec.Descriptor)
	return desc, ok
}

type fileWriter struct {
	store       *FileStore
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
