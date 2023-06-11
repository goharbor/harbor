package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
)

// FileRemoteStore provides a RemoteStore interface compatible
// implementation that can be used where the RemoteStore is backed by a
// fs.FS. This is useful for example in air-gapped environments where there's no
// possibility to make outbound network connections.
// By having this be a fs.FS instead of directories allows the repository to
// be backed by something that's not persisted to disk.
func NewFileRemoteStore(fsys fs.FS, targetDir string) (*FileRemoteStore, error) {
	if fsys == nil {
		return nil, errors.New("nil fs.FS")
	}
	t := targetDir
	if t == "" {
		t = "targets"
	}
	// Make sure directory exists
	d, err := fsys.Open(t)
	if err != nil {
		return nil, fmt.Errorf("failed to open targets directory %s: %w", t, err)
	}
	fi, err := d.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat targets directory %s: %w", t, err)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("targets directory not a directory %s", t)
	}

	fsysT, err := fs.Sub(fsys, t)
	if err != nil {
		return nil, fmt.Errorf("failed to open targets directory %s: %w", t, err)
	}
	return &FileRemoteStore{fsys: fsys, targetDir: fsysT}, nil
}

type FileRemoteStore struct {
	// Meta directory fs
	fsys fs.FS
	// Target directory fs.
	targetDir fs.FS
	// In order to be able to make write operations (create, delete) we can't
	// use fs.FS for it (it's read only), so we have to know the underlying
	// directory that add/delete test methods can use. This is only necessary
	// for testing purposes.
	testDir string
}

func (f *FileRemoteStore) GetMeta(name string) (io.ReadCloser, int64, error) {
	rc, b, err := f.get(f.fsys, name)
	return handleErrors(name, rc, b, err)
}

func (f *FileRemoteStore) GetTarget(name string) (io.ReadCloser, int64, error) {
	rc, b, err := f.get(f.targetDir, name)
	return handleErrors(name, rc, b, err)
}

func (f *FileRemoteStore) get(fsys fs.FS, s string) (io.ReadCloser, int64, error) {
	if !fs.ValidPath(s) {
		return nil, 0, fmt.Errorf("invalid path %s", s)
	}

	b, err := fs.ReadFile(fsys, s)
	if err != nil {
		return nil, -1, err
	}
	return io.NopCloser(bytes.NewReader(b)), int64(len(b)), nil
}

// handleErrors converts NotFound errors to something that TUF knows how to
// handle properly. For example, when looking for n+1 root files, this is a
// signal that it will stop looking.
func handleErrors(name string, rc io.ReadCloser, b int64, err error) (io.ReadCloser, int64, error) {
	if err == nil {
		return rc, b, err
	}
	if errors.Is(err, fs.ErrNotExist) {
		return rc, b, ErrNotFound{name}
	}
	return rc, b, err
}
