// +build include_rados

package rados

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/distribution/context"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/base"
	"github.com/docker/distribution/registry/storage/driver/factory"
	"github.com/docker/distribution/uuid"
	"github.com/noahdesu/go-ceph/rados"
)

const driverName = "rados"

// Prefix all the stored blob
const objectBlobPrefix = "blob:"

// Stripes objects size to 4M
const defaultChunkSize = 4 << 20
const defaultXattrTotalSizeName = "total-size"

// Max number of keys fetched from omap at each read operation
const defaultKeysFetched = 1

//DriverParameters A struct that encapsulates all of the driver parameters after all values have been set
type DriverParameters struct {
	poolname  string
	username  string
	chunksize uint64
}

func init() {
	factory.Register(driverName, &radosDriverFactory{})
}

// radosDriverFactory implements the factory.StorageDriverFactory interface
type radosDriverFactory struct{}

func (factory *radosDriverFactory) Create(parameters map[string]interface{}) (storagedriver.StorageDriver, error) {
	return FromParameters(parameters)
}

type driver struct {
	Conn      *rados.Conn
	Ioctx     *rados.IOContext
	chunksize uint64
}

type baseEmbed struct {
	base.Base
}

// Driver is a storagedriver.StorageDriver implementation backed by Ceph RADOS
// Objects are stored at absolute keys in the provided bucket.
type Driver struct {
	baseEmbed
}

// FromParameters constructs a new Driver with a given parameters map
// Required parameters:
// - poolname: the ceph pool name
func FromParameters(parameters map[string]interface{}) (*Driver, error) {

	pool, ok := parameters["poolname"]
	if !ok {
		return nil, fmt.Errorf("No poolname parameter provided")
	}

	username, ok := parameters["username"]
	if !ok {
		username = ""
	}

	chunksize := uint64(defaultChunkSize)
	chunksizeParam, ok := parameters["chunksize"]
	if ok {
		chunksize, ok = chunksizeParam.(uint64)
		if !ok {
			return nil, fmt.Errorf("The chunksize parameter should be a number")
		}
	}

	params := DriverParameters{
		fmt.Sprint(pool),
		fmt.Sprint(username),
		chunksize,
	}

	return New(params)
}

// New constructs a new Driver
func New(params DriverParameters) (*Driver, error) {
	var conn *rados.Conn
	var err error

	if params.username != "" {
		log.Infof("Opening connection to pool %s using user %s", params.poolname, params.username)
		conn, err = rados.NewConnWithUser(params.username)
	} else {
		log.Infof("Opening connection to pool %s", params.poolname)
		conn, err = rados.NewConn()
	}

	if err != nil {
		return nil, err
	}

	err = conn.ReadDefaultConfigFile()
	if err != nil {
		return nil, err
	}

	err = conn.Connect()
	if err != nil {
		return nil, err
	}

	log.Infof("Connected")

	ioctx, err := conn.OpenIOContext(params.poolname)

	log.Infof("Connected to pool %s", params.poolname)

	if err != nil {
		return nil, err
	}

	d := &driver{
		Ioctx:     ioctx,
		Conn:      conn,
		chunksize: params.chunksize,
	}

	return &Driver{
		baseEmbed: baseEmbed{
			Base: base.Base{
				StorageDriver: d,
			},
		},
	}, nil
}

// Implement the storagedriver.StorageDriver interface

func (d *driver) Name() string {
	return driverName
}

// GetContent retrieves the content stored at "path" as a []byte.
func (d *driver) GetContent(ctx context.Context, path string) ([]byte, error) {
	rc, err := d.ReadStream(ctx, path, 0)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	p, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// PutContent stores the []byte content at a location designated by "path".
func (d *driver) PutContent(ctx context.Context, path string, contents []byte) error {
	if _, err := d.WriteStream(ctx, path, 0, bytes.NewReader(contents)); err != nil {
		return err
	}

	return nil
}

// ReadStream retrieves an io.ReadCloser for the content stored at "path" with a
// given byte offset.
type readStreamReader struct {
	driver *driver
	oid    string
	size   uint64
	offset uint64
}

func (r *readStreamReader) Read(b []byte) (n int, err error) {
	// Determine the part available to read
	bufferOffset := uint64(0)
	bufferSize := uint64(len(b))

	// End of the object, read less than the buffer size
	if bufferSize > r.size-r.offset {
		bufferSize = r.size - r.offset
	}

	// Fill `b`
	for bufferOffset < bufferSize {
		// Get the offset in the object chunk
		chunkedOid, chunkedOffset := r.driver.getChunkNameFromOffset(r.oid, r.offset)

		// Determine the best size to read
		bufferEndOffset := bufferSize
		if bufferEndOffset-bufferOffset > r.driver.chunksize-chunkedOffset {
			bufferEndOffset = bufferOffset + (r.driver.chunksize - chunkedOffset)
		}

		// Read the chunk
		n, err = r.driver.Ioctx.Read(chunkedOid, b[bufferOffset:bufferEndOffset], chunkedOffset)

		if err != nil {
			return int(bufferOffset), err
		}

		bufferOffset += uint64(n)
		r.offset += uint64(n)
	}

	// EOF if the offset is at the end of the object
	if r.offset == r.size {
		return int(bufferOffset), io.EOF
	}

	return int(bufferOffset), nil
}

func (r *readStreamReader) Close() error {
	return nil
}

func (d *driver) ReadStream(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	// get oid from filename
	oid, err := d.getOid(path)

	if err != nil {
		return nil, err
	}

	// get object stat
	stat, err := d.Stat(ctx, path)

	if err != nil {
		return nil, err
	}

	if offset > stat.Size() {
		return nil, storagedriver.InvalidOffsetError{Path: path, Offset: offset}
	}

	return &readStreamReader{
		driver: d,
		oid:    oid,
		size:   uint64(stat.Size()),
		offset: uint64(offset),
	}, nil
}

func (d *driver) WriteStream(ctx context.Context, path string, offset int64, reader io.Reader) (totalRead int64, err error) {
	buf := make([]byte, d.chunksize)
	totalRead = 0

	oid, err := d.getOid(path)
	if err != nil {
		switch err.(type) {
		// Trying to write new object, generate new blob identifier for it
		case storagedriver.PathNotFoundError:
			oid = d.generateOid()
			err = d.putOid(path, oid)
			if err != nil {
				return 0, err
			}
		default:
			return 0, err
		}
	} else {
		// Check total object size only for existing ones
		totalSize, err := d.getXattrTotalSize(ctx, oid)
		if err != nil {
			return 0, err
		}

		// If offset if after the current object size, fill the gap with zeros
		for totalSize < uint64(offset) {
			sizeToWrite := d.chunksize
			if totalSize-uint64(offset) < sizeToWrite {
				sizeToWrite = totalSize - uint64(offset)
			}

			chunkName, chunkOffset := d.getChunkNameFromOffset(oid, uint64(totalSize))
			err = d.Ioctx.Write(chunkName, buf[:sizeToWrite], uint64(chunkOffset))
			if err != nil {
				return totalRead, err
			}

			totalSize += sizeToWrite
		}
	}

	// Writer
	for {
		// Align to chunk size
		sizeRead := uint64(0)
		sizeToRead := uint64(offset+totalRead) % d.chunksize
		if sizeToRead == 0 {
			sizeToRead = d.chunksize
		}

		// Read from `reader`
		for sizeRead < sizeToRead {
			nn, err := reader.Read(buf[sizeRead:sizeToRead])
			sizeRead += uint64(nn)

			if err != nil {
				if err != io.EOF {
					return totalRead, err
				}

				break
			}
		}

		// End of file and nothing was read
		if sizeRead == 0 {
			break
		}

		// Write chunk object
		chunkName, chunkOffset := d.getChunkNameFromOffset(oid, uint64(offset+totalRead))
		err = d.Ioctx.Write(chunkName, buf[:sizeRead], uint64(chunkOffset))

		if err != nil {
			return totalRead, err
		}

		// Update total object size as xattr in the first chunk of the object
		err = d.setXattrTotalSize(oid, uint64(offset+totalRead)+sizeRead)
		if err != nil {
			return totalRead, err
		}

		totalRead += int64(sizeRead)

		// End of file
		if sizeRead < sizeToRead {
			break
		}
	}

	return totalRead, nil
}

// Stat retrieves the FileInfo for the given path, including the current size
func (d *driver) Stat(ctx context.Context, path string) (storagedriver.FileInfo, error) {
	// get oid from filename
	oid, err := d.getOid(path)

	if err != nil {
		return nil, err
	}

	// the path is a virtual directory?
	if oid == "" {
		return storagedriver.FileInfoInternal{
			FileInfoFields: storagedriver.FileInfoFields{
				Path:  path,
				Size:  0,
				IsDir: true,
			},
		}, nil
	}

	// stat first chunk
	stat, err := d.Ioctx.Stat(oid + "-0")

	if err != nil {
		return nil, err
	}

	// get total size of chunked object
	totalSize, err := d.getXattrTotalSize(ctx, oid)

	if err != nil {
		return nil, err
	}

	return storagedriver.FileInfoInternal{
		FileInfoFields: storagedriver.FileInfoFields{
			Path:    path,
			Size:    int64(totalSize),
			ModTime: stat.ModTime,
		},
	}, nil
}

// List returns a list of the objects that are direct descendants of the given path.
func (d *driver) List(ctx context.Context, dirPath string) ([]string, error) {
	files, err := d.listDirectoryOid(dirPath)

	if err != nil {
		return nil, storagedriver.PathNotFoundError{Path: dirPath}
	}

	keys := make([]string, 0, len(files))
	for k := range files {
		if k != dirPath {
			keys = append(keys, path.Join(dirPath, k))
		}
	}

	return keys, nil
}

// Move moves an object stored at sourcePath to destPath, removing the original
// object.
func (d *driver) Move(ctx context.Context, sourcePath string, destPath string) error {
	// Get oid
	oid, err := d.getOid(sourcePath)

	if err != nil {
		return err
	}

	// Move reference
	err = d.putOid(destPath, oid)

	if err != nil {
		return err
	}

	// Delete old reference
	err = d.deleteOid(sourcePath)

	if err != nil {
		return err
	}

	return nil
}

// Delete recursively deletes all objects stored at "path" and its subpaths.
func (d *driver) Delete(ctx context.Context, objectPath string) error {
	// Get oid
	oid, err := d.getOid(objectPath)

	if err != nil {
		return err
	}

	// Deleting virtual directory
	if oid == "" {
		objects, err := d.listDirectoryOid(objectPath)
		if err != nil {
			return err
		}

		for object := range objects {
			err = d.Delete(ctx, path.Join(objectPath, object))
			if err != nil {
				return err
			}
		}
	} else {
		// Delete object chunks
		totalSize, err := d.getXattrTotalSize(ctx, oid)

		if err != nil {
			return err
		}

		for offset := uint64(0); offset < totalSize; offset += d.chunksize {
			chunkName, _ := d.getChunkNameFromOffset(oid, offset)

			err = d.Ioctx.Delete(chunkName)
			if err != nil {
				return err
			}
		}

		// Delete reference
		err = d.deleteOid(objectPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// URLFor returns a URL which may be used to retrieve the content stored at the given path.
// May return an UnsupportedMethodErr in certain StorageDriver implementations.
func (d *driver) URLFor(ctx context.Context, path string, options map[string]interface{}) (string, error) {
	return "", storagedriver.ErrUnsupportedMethod{}
}

// Generate a blob identifier
func (d *driver) generateOid() string {
	return objectBlobPrefix + uuid.Generate().String()
}

// Reference a object and its hierarchy
func (d *driver) putOid(objectPath string, oid string) error {
	directory := path.Dir(objectPath)
	base := path.Base(objectPath)
	createParentReference := true

	// After creating this reference, skip the parents referencing since the
	// hierarchy already exists
	if oid == "" {
		firstReference, err := d.Ioctx.GetOmapValues(directory, "", "", 1)
		if (err == nil) && (len(firstReference) > 0) {
			createParentReference = false
		}
	}

	oids := map[string][]byte{
		base: []byte(oid),
	}

	// Reference object
	err := d.Ioctx.SetOmap(directory, oids)
	if err != nil {
		return err
	}

	// Esure parent virtual directories
	if createParentReference {
		return d.putOid(directory, "")
	}

	return nil
}

// Get the object identifier from an object name
func (d *driver) getOid(objectPath string) (string, error) {
	directory := path.Dir(objectPath)
	base := path.Base(objectPath)

	files, err := d.Ioctx.GetOmapValues(directory, "", base, 1)

	if (err != nil) || (files[base] == nil) {
		return "", storagedriver.PathNotFoundError{Path: objectPath}
	}

	return string(files[base]), nil
}

// List the objects of a virtual directory
func (d *driver) listDirectoryOid(path string) (list map[string][]byte, err error) {
	return d.Ioctx.GetAllOmapValues(path, "", "", defaultKeysFetched)
}

// Remove a file from the files hierarchy
func (d *driver) deleteOid(objectPath string) error {
	// Remove object reference
	directory := path.Dir(objectPath)
	base := path.Base(objectPath)
	err := d.Ioctx.RmOmapKeys(directory, []string{base})

	if err != nil {
		return err
	}

	// Remove virtual directory if empty (no more references)
	firstReference, err := d.Ioctx.GetOmapValues(directory, "", "", 1)

	if err != nil {
		return err
	}

	if len(firstReference) == 0 {
		// Delete omap
		err := d.Ioctx.Delete(directory)

		if err != nil {
			return err
		}

		// Remove reference on parent omaps
		if directory != "" {
			return d.deleteOid(directory)
		}
	}

	return nil
}

// Takes an offset in an chunked object and return the chunk name and a new
// offset in this chunk object
func (d *driver) getChunkNameFromOffset(oid string, offset uint64) (string, uint64) {
	chunkID := offset / d.chunksize
	chunkedOid := oid + "-" + strconv.FormatInt(int64(chunkID), 10)
	chunkedOffset := offset % d.chunksize
	return chunkedOid, chunkedOffset
}

// Set the total size of a chunked object `oid`
func (d *driver) setXattrTotalSize(oid string, size uint64) error {
	// Convert uint64 `size` to []byte
	xattr := make([]byte, binary.MaxVarintLen64)
	binary.LittleEndian.PutUint64(xattr, size)

	// Save the total size as a xattr in the first chunk
	return d.Ioctx.SetXattr(oid+"-0", defaultXattrTotalSizeName, xattr)
}

// Get the total size of the chunked object `oid` stored as xattr
func (d *driver) getXattrTotalSize(ctx context.Context, oid string) (uint64, error) {
	// Fetch xattr as []byte
	xattr := make([]byte, binary.MaxVarintLen64)
	xattrLength, err := d.Ioctx.GetXattr(oid+"-0", defaultXattrTotalSizeName, xattr)

	if err != nil {
		return 0, err
	}

	if xattrLength != len(xattr) {
		context.GetLogger(ctx).Errorf("object %s xattr length mismatch: %d != %d", oid, xattrLength, len(xattr))
		return 0, storagedriver.PathNotFoundError{Path: oid}
	}

	// Convert []byte as uint64
	totalSize := binary.LittleEndian.Uint64(xattr)

	return totalSize, nil
}
