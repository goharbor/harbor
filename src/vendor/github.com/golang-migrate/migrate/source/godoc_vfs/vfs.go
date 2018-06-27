// Package vfs contains a driver that reads migrations from a virtual file
// system.
//
// Implementations of the filesystem interface that read from zip files and
// maps, as well as the definition of the filesystem interface can be found in
// the golang.org/x/tools/godoc/vfs package.
package godoc_vfs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/golang-migrate/migrate/source"
	"golang.org/x/tools/godoc/vfs"
)

func init() {
	source.Register("godoc-vfs", &VFS{})
}

// VFS is an implementation of driver that returns migrations from a virtual
// file system.
type VFS struct {
	migrations *source.Migrations
	fs         vfs.FileSystem
	path       string
}

// Open implements the source.Driver interface for VFS.
//
// Calling this function panics, instead use the WithInstance function.
// See the package level documentation for an example.
func (b *VFS) Open(url string) (source.Driver, error) {
	panic("not implemented")
}

// WithInstance creates a new driver from a virtual file system.
// If a tree named searchPath exists in the virtual filesystem, WithInstance
// searches for migration files there.
// It defaults to "/".
func WithInstance(fs vfs.FileSystem, searchPath string) (source.Driver, error) {
	if searchPath == "" {
		searchPath = "/"
	}

	bn := &VFS{
		fs:         fs,
		path:       searchPath,
		migrations: source.NewMigrations(),
	}

	files, err := fs.ReadDir(searchPath)
	if err != nil {
		return nil, err
	}

	for _, fi := range files {
		m, err := source.DefaultParse(fi.Name())
		if err != nil {
			continue // ignore files that we can't parse
		}

		if !bn.migrations.Append(m) {
			return nil, fmt.Errorf("unable to parse file %v", fi)
		}
	}

	return bn, nil
}

// Close implements the source.Driver interface for VFS.
// It is a no-op and should not be used.
func (b *VFS) Close() error {
	return nil
}

// First returns the first migration verion found in the file system.
// If no version is available os.ErrNotExist is returned.
func (b *VFS) First() (version uint, err error) {
	v, ok := b.migrations.First()
	if !ok {
		return 0, &os.PathError{"first", "<vfs>://" + b.path, os.ErrNotExist}
	}
	return v, nil
}

// Prev returns the previous version available to the driver.
// If no previous version is available os.ErrNotExist is returned.
func (b *VFS) Prev(version uint) (prevVersion uint, err error) {
	v, ok := b.migrations.Prev(version)
	if !ok {
		return 0, &os.PathError{fmt.Sprintf("prev for version %v", version), "<vfs>://" + b.path, os.ErrNotExist}
	}
	return v, nil
}

// Prev returns the next version available to the driver.
// If no previous version is available os.ErrNotExist is returned.
func (b *VFS) Next(version uint) (nextVersion uint, err error) {
	v, ok := b.migrations.Next(version)
	if !ok {
		return 0, &os.PathError{fmt.Sprintf("next for version %v", version), "<vfs>://" + b.path, os.ErrNotExist}
	}
	return v, nil
}

// ReadUp returns the up migration body and an identifier that helps with
// finding this migration in the source.
// If there is no up migration available for this version it returns
// os.ErrNotExist.
func (b *VFS) ReadUp(version uint) (r io.ReadCloser, identifier string, err error) {
	if m, ok := b.migrations.Up(version); ok {
		body, err := vfs.ReadFile(b.fs, path.Join(b.path, m.Raw))
		if err != nil {
			return nil, "", err
		}
		return ioutil.NopCloser(bytes.NewReader(body)), m.Identifier, nil
	}
	return nil, "", &os.PathError{fmt.Sprintf("read version %v", version), "<vfs>://" + b.path, os.ErrNotExist}
}

// ReadDown returns the down migration body and an identifier that helps with
// finding this migration  in the source.
func (b *VFS) ReadDown(version uint) (r io.ReadCloser, identifier string, err error) {
	if m, ok := b.migrations.Down(version); ok {
		body, err := vfs.ReadFile(b.fs, path.Join(b.path, m.Raw))
		if err != nil {
			return nil, "", err
		}
		return ioutil.NopCloser(bytes.NewReader(body)), m.Identifier, nil
	}
	return nil, "", &os.PathError{fmt.Sprintf("read version %v", version), "<vfs>://" + b.path, os.ErrNotExist}
}
