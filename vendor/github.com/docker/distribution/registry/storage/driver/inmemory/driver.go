package inmemory

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/docker/distribution/context"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/base"
	"github.com/docker/distribution/registry/storage/driver/factory"
)

const driverName = "inmemory"

func init() {
	factory.Register(driverName, &inMemoryDriverFactory{})
}

// inMemoryDriverFacotry implements the factory.StorageDriverFactory interface.
type inMemoryDriverFactory struct{}

func (factory *inMemoryDriverFactory) Create(parameters map[string]interface{}) (storagedriver.StorageDriver, error) {
	return New(), nil
}

type driver struct {
	root  *dir
	mutex sync.RWMutex
}

// baseEmbed allows us to hide the Base embed.
type baseEmbed struct {
	base.Base
}

// Driver is a storagedriver.StorageDriver implementation backed by a local map.
// Intended solely for example and testing purposes.
type Driver struct {
	baseEmbed // embedded, hidden base driver.
}

var _ storagedriver.StorageDriver = &Driver{}

// New constructs a new Driver.
func New() *Driver {
	return &Driver{
		baseEmbed: baseEmbed{
			Base: base.Base{
				StorageDriver: &driver{
					root: &dir{
						common: common{
							p:   "/",
							mod: time.Now(),
						},
					},
				},
			},
		},
	}
}

// Implement the storagedriver.StorageDriver interface.

func (d *driver) Name() string {
	return driverName
}

// GetContent retrieves the content stored at "path" as a []byte.
func (d *driver) GetContent(ctx context.Context, path string) ([]byte, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	rc, err := d.ReadStream(ctx, path, 0)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return ioutil.ReadAll(rc)
}

// PutContent stores the []byte content at a location designated by "path".
func (d *driver) PutContent(ctx context.Context, p string, contents []byte) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	f, err := d.root.mkfile(p)
	if err != nil {
		// TODO(stevvooe): Again, we need to clarify when this is not a
		// directory in StorageDriver API.
		return fmt.Errorf("not a file")
	}

	f.truncate()
	f.WriteAt(contents, 0)

	return nil
}

// ReadStream retrieves an io.ReadCloser for the content stored at "path" with a
// given byte offset.
func (d *driver) ReadStream(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if offset < 0 {
		return nil, storagedriver.InvalidOffsetError{Path: path, Offset: offset}
	}

	path = normalize(path)
	found := d.root.find(path)

	if found.path() != path {
		return nil, storagedriver.PathNotFoundError{Path: path}
	}

	if found.isdir() {
		return nil, fmt.Errorf("%q is a directory", path)
	}

	return ioutil.NopCloser(found.(*file).sectionReader(offset)), nil
}

// WriteStream stores the contents of the provided io.ReadCloser at a location
// designated by the given path.
func (d *driver) WriteStream(ctx context.Context, path string, offset int64, reader io.Reader) (nn int64, err error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if offset < 0 {
		return 0, storagedriver.InvalidOffsetError{Path: path, Offset: offset}
	}

	normalized := normalize(path)

	f, err := d.root.mkfile(normalized)
	if err != nil {
		return 0, fmt.Errorf("not a file")
	}

	// Unlock while we are reading from the source, in case we are reading
	// from the same mfs instance. This can be fixed by a more granular
	// locking model.
	d.mutex.Unlock()
	d.mutex.RLock() // Take the readlock to block other writers.
	var buf bytes.Buffer

	nn, err = buf.ReadFrom(reader)
	if err != nil {
		// TODO(stevvooe): This condition is odd and we may need to clarify:
		// we've read nn bytes from reader but have written nothing to the
		// backend. What is the correct return value? Really, the caller needs
		// to know that the reader has been advanced and reattempting the
		// operation is incorrect.
		d.mutex.RUnlock()
		d.mutex.Lock()
		return nn, err
	}

	d.mutex.RUnlock()
	d.mutex.Lock()
	f.WriteAt(buf.Bytes(), offset)
	return nn, err
}

// Stat returns info about the provided path.
func (d *driver) Stat(ctx context.Context, path string) (storagedriver.FileInfo, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	normalized := normalize(path)
	found := d.root.find(path)

	if found.path() != normalized {
		return nil, storagedriver.PathNotFoundError{Path: path}
	}

	fi := storagedriver.FileInfoFields{
		Path:    path,
		IsDir:   found.isdir(),
		ModTime: found.modtime(),
	}

	if !fi.IsDir {
		fi.Size = int64(len(found.(*file).data))
	}

	return storagedriver.FileInfoInternal{FileInfoFields: fi}, nil
}

// List returns a list of the objects that are direct descendants of the given
// path.
func (d *driver) List(ctx context.Context, path string) ([]string, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	normalized := normalize(path)

	found := d.root.find(normalized)

	if !found.isdir() {
		return nil, fmt.Errorf("not a directory") // TODO(stevvooe): Need error type for this...
	}

	entries, err := found.(*dir).list(normalized)

	if err != nil {
		switch err {
		case errNotExists:
			return nil, storagedriver.PathNotFoundError{Path: path}
		case errIsNotDir:
			return nil, fmt.Errorf("not a directory")
		default:
			return nil, err
		}
	}

	return entries, nil
}

// Move moves an object stored at sourcePath to destPath, removing the original
// object.
func (d *driver) Move(ctx context.Context, sourcePath string, destPath string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	normalizedSrc, normalizedDst := normalize(sourcePath), normalize(destPath)

	err := d.root.move(normalizedSrc, normalizedDst)
	switch err {
	case errNotExists:
		return storagedriver.PathNotFoundError{Path: destPath}
	default:
		return err
	}
}

// Delete recursively deletes all objects stored at "path" and its subpaths.
func (d *driver) Delete(ctx context.Context, path string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	normalized := normalize(path)

	err := d.root.delete(normalized)
	switch err {
	case errNotExists:
		return storagedriver.PathNotFoundError{Path: path}
	default:
		return err
	}
}

// URLFor returns a URL which may be used to retrieve the content stored at the given path.
// May return an UnsupportedMethodErr in certain StorageDriver implementations.
func (d *driver) URLFor(ctx context.Context, path string, options map[string]interface{}) (string, error) {
	return "", storagedriver.ErrUnsupportedMethod{}
}
