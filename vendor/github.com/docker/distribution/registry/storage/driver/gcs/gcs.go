// Package gcs provides a storagedriver.StorageDriver implementation to
// store blobs in Google cloud storage.
//
// This package leverages the google.golang.org/cloud/storage client library
//for interfacing with gcs.
//
// Because gcs is a key, value store the Stat call does not support last modification
// time for directories (directories are an abstraction for key, value stores)
//
// Keep in mind that gcs guarantees only eventual consistency, so do not assume
// that a successful write will mean immediate access to the data written (although
// in most regions a new object put has guaranteed read after write). The only true
// guarantee is that once you call Stat and receive a certain file size, that much of
// the file is already accessible.
//
// +build include_gcs

package gcs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/googleapi"
	storageapi "google.golang.org/api/storage/v1"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"

	ctx "github.com/docker/distribution/context"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/base"
	"github.com/docker/distribution/registry/storage/driver/factory"
)

const driverName = "gcs"
const dummyProjectID = "<unknown>"

// driverParameters is a struct that encapsulates all of the driver parameters after all values have been set
type driverParameters struct {
	bucket        string
	config        *jwt.Config
	email         string
	privateKey    []byte
	client        *http.Client
	rootDirectory string
}

func init() {
	factory.Register(driverName, &gcsDriverFactory{})
}

// gcsDriverFactory implements the factory.StorageDriverFactory interface
type gcsDriverFactory struct{}

// Create StorageDriver from parameters
func (factory *gcsDriverFactory) Create(parameters map[string]interface{}) (storagedriver.StorageDriver, error) {
	return FromParameters(parameters)
}

// driver is a storagedriver.StorageDriver implementation backed by GCS
// Objects are stored at absolute keys in the provided bucket.
type driver struct {
	client        *http.Client
	bucket        string
	email         string
	privateKey    []byte
	rootDirectory string
}

// FromParameters constructs a new Driver with a given parameters map
// Required parameters:
// - bucket
func FromParameters(parameters map[string]interface{}) (storagedriver.StorageDriver, error) {
	bucket, ok := parameters["bucket"]
	if !ok || fmt.Sprint(bucket) == "" {
		return nil, fmt.Errorf("No bucket parameter provided")
	}

	rootDirectory, ok := parameters["rootdirectory"]
	if !ok {
		rootDirectory = ""
	}

	var ts oauth2.TokenSource
	jwtConf := new(jwt.Config)
	if keyfile, ok := parameters["keyfile"]; ok {
		jsonKey, err := ioutil.ReadFile(fmt.Sprint(keyfile))
		if err != nil {
			return nil, err
		}
		jwtConf, err = google.JWTConfigFromJSON(jsonKey, storage.ScopeFullControl)
		if err != nil {
			return nil, err
		}
		ts = jwtConf.TokenSource(context.Background())
	} else {
		var err error
		ts, err = google.DefaultTokenSource(context.Background(), storage.ScopeFullControl)
		if err != nil {
			return nil, err
		}

	}

	params := driverParameters{
		bucket:        fmt.Sprint(bucket),
		rootDirectory: fmt.Sprint(rootDirectory),
		email:         jwtConf.Email,
		privateKey:    jwtConf.PrivateKey,
		client:        oauth2.NewClient(context.Background(), ts),
	}

	return New(params)
}

// New constructs a new driver
func New(params driverParameters) (storagedriver.StorageDriver, error) {
	rootDirectory := strings.Trim(params.rootDirectory, "/")
	if rootDirectory != "" {
		rootDirectory += "/"
	}
	d := &driver{
		bucket:        params.bucket,
		rootDirectory: rootDirectory,
		email:         params.email,
		privateKey:    params.privateKey,
		client:        params.client,
	}

	return &base.Base{
		StorageDriver: d,
	}, nil
}

// Implement the storagedriver.StorageDriver interface

func (d *driver) Name() string {
	return driverName
}

// GetContent retrieves the content stored at "path" as a []byte.
// This should primarily be used for small objects.
func (d *driver) GetContent(context ctx.Context, path string) ([]byte, error) {
	rc, err := d.ReadStream(context, path, 0)
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
// This should primarily be used for small objects.
func (d *driver) PutContent(context ctx.Context, path string, contents []byte) error {
	wc := storage.NewWriter(d.context(context), d.bucket, d.pathToKey(path))
	wc.ContentType = "application/octet-stream"
	defer wc.Close()
	_, err := wc.Write(contents)
	return err
}

// ReadStream retrieves an io.ReadCloser for the content stored at "path"
// with a given byte offset.
// May be used to resume reading a stream by providing a nonzero offset.
func (d *driver) ReadStream(context ctx.Context, path string, offset int64) (io.ReadCloser, error) {
	name := d.pathToKey(path)

	// copied from google.golang.org/cloud/storage#NewReader :
	// to set the additional "Range" header
	u := &url.URL{
		Scheme: "https",
		Host:   "storage.googleapis.com",
		Path:   fmt.Sprintf("/%s/%s", d.bucket, name),
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if offset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%v-", offset))
	}
	res, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		res.Body.Close()
		return nil, storagedriver.PathNotFoundError{Path: path}
	}
	if res.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		res.Body.Close()
		obj, err := storageStatObject(d.context(context), d.bucket, name)
		if err != nil {
			return nil, err
		}
		if offset == int64(obj.Size) {
			return ioutil.NopCloser(bytes.NewReader([]byte{})), nil
		}
		return nil, storagedriver.InvalidOffsetError{Path: path, Offset: offset}
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		res.Body.Close()
		return nil, fmt.Errorf("storage: can't read object %v/%v, status code: %v", d.bucket, name, res.Status)
	}
	return res.Body, nil
}

// WriteStream stores the contents of the provided io.ReadCloser at a
// location designated by the given path.
// May be used to resume writing a stream by providing a nonzero offset.
// The offset must be no larger than the CurrentSize for this path.
func (d *driver) WriteStream(context ctx.Context, path string, offset int64, reader io.Reader) (totalRead int64, err error) {
	if offset < 0 {
		return 0, storagedriver.InvalidOffsetError{Path: path, Offset: offset}
	}

	if offset == 0 {
		return d.writeCompletely(context, path, 0, reader)
	}

	service, err := storageapi.New(d.client)
	if err != nil {
		return 0, err
	}
	objService := storageapi.NewObjectsService(service)
	var obj *storageapi.Object
	err = retry(5, func() error {
		o, err := objService.Get(d.bucket, d.pathToKey(path)).Do()
		obj = o
		return err
	})
	//	obj, err := retry(5, objService.Get(d.bucket, d.pathToKey(path)).Do)
	if err != nil {
		return 0, err
	}

	// cannot append more chunks, so redo from scratch
	if obj.ComponentCount >= 1023 {
		return d.writeCompletely(context, path, offset, reader)
	}

	// skip from reader
	objSize := int64(obj.Size)
	nn, err := skip(reader, objSize-offset)
	if err != nil {
		return nn, err
	}

	// Size <= offset
	partName := fmt.Sprintf("%v#part-%d#", d.pathToKey(path), obj.ComponentCount)
	gcsContext := d.context(context)
	wc := storage.NewWriter(gcsContext, d.bucket, partName)
	wc.ContentType = "application/octet-stream"

	if objSize < offset {
		err = writeZeros(wc, offset-objSize)
		if err != nil {
			wc.CloseWithError(err)
			return nn, err
		}
	}
	n, err := io.Copy(wc, reader)
	if err != nil {
		wc.CloseWithError(err)
		return nn, err
	}
	err = wc.Close()
	if err != nil {
		return nn, err
	}
	// wc was closed successfully, so the temporary part exists, schedule it for deletion at the end
	// of the function
	defer storageDeleteObject(gcsContext, d.bucket, partName)

	req := &storageapi.ComposeRequest{
		Destination: &storageapi.Object{Bucket: obj.Bucket, Name: obj.Name, ContentType: obj.ContentType},
		SourceObjects: []*storageapi.ComposeRequestSourceObjects{
			{
				Name:       obj.Name,
				Generation: obj.Generation,
			}, {
				Name:       partName,
				Generation: wc.Object().Generation,
			}},
	}

	err = retry(5, func() error { _, err := objService.Compose(d.bucket, obj.Name, req).Do(); return err })
	if err == nil {
		nn = nn + n
	}

	return nn, err
}

type request func() error

func retry(maxTries int, req request) error {
	backoff := time.Second
	var err error
	for i := 0; i < maxTries; i++ {
		err = req()
		if err == nil {
			return nil
		}

		status, ok := err.(*googleapi.Error)
		if !ok || (status.Code != 429 && status.Code < http.StatusInternalServerError) {
			return err
		}

		time.Sleep(backoff - time.Second + (time.Duration(rand.Int31n(1000)) * time.Millisecond))
		if i <= 4 {
			backoff = backoff * 2
		}
	}
	return err
}

func (d *driver) writeCompletely(context ctx.Context, path string, offset int64, reader io.Reader) (totalRead int64, err error) {
	wc := storage.NewWriter(d.context(context), d.bucket, d.pathToKey(path))
	wc.ContentType = "application/octet-stream"
	defer wc.Close()

	// Copy the first offset bytes of the existing contents
	// (padded with zeros if needed) into the writer
	if offset > 0 {
		existing, err := d.ReadStream(context, path, 0)
		if err != nil {
			return 0, err
		}
		defer existing.Close()
		n, err := io.CopyN(wc, existing, offset)
		if err == io.EOF {
			err = writeZeros(wc, offset-n)
		}
		if err != nil {
			return 0, err
		}
	}
	return io.Copy(wc, reader)
}

func skip(reader io.Reader, count int64) (int64, error) {
	if count <= 0 {
		return 0, nil
	}
	return io.CopyN(ioutil.Discard, reader, count)
}

func writeZeros(wc io.Writer, count int64) error {
	buf := make([]byte, 32*1024)
	for count > 0 {
		size := cap(buf)
		if int64(size) > count {
			size = int(count)
		}
		n, err := wc.Write(buf[0:size])
		if err != nil {
			return err
		}
		count = count - int64(n)
	}
	return nil
}

// Stat retrieves the FileInfo for the given path, including the current
// size in bytes and the creation time.
func (d *driver) Stat(context ctx.Context, path string) (storagedriver.FileInfo, error) {
	var fi storagedriver.FileInfoFields
	//try to get as file
	gcsContext := d.context(context)
	obj, err := storageStatObject(gcsContext, d.bucket, d.pathToKey(path))
	if err == nil {
		fi = storagedriver.FileInfoFields{
			Path:    path,
			Size:    obj.Size,
			ModTime: obj.Updated,
			IsDir:   false,
		}
		return storagedriver.FileInfoInternal{FileInfoFields: fi}, nil
	}
	//try to get as folder
	dirpath := d.pathToDirKey(path)

	var query *storage.Query
	query = &storage.Query{}
	query.Prefix = dirpath
	query.MaxResults = 1

	objects, err := storageListObjects(gcsContext, d.bucket, query)
	if err != nil {
		return nil, err
	}
	if len(objects.Results) < 1 {
		return nil, storagedriver.PathNotFoundError{Path: path}
	}
	fi = storagedriver.FileInfoFields{
		Path:  path,
		IsDir: true,
	}
	obj = objects.Results[0]
	if obj.Name == dirpath {
		fi.Size = obj.Size
		fi.ModTime = obj.Updated
	}
	return storagedriver.FileInfoInternal{FileInfoFields: fi}, nil
}

// List returns a list of the objects that are direct descendants of the
//given path.
func (d *driver) List(context ctx.Context, path string) ([]string, error) {
	var query *storage.Query
	query = &storage.Query{}
	query.Delimiter = "/"
	query.Prefix = d.pathToDirKey(path)
	list := make([]string, 0, 64)
	for {
		objects, err := storageListObjects(d.context(context), d.bucket, query)
		if err != nil {
			return nil, err
		}
		for _, object := range objects.Results {
			// GCS does not guarantee strong consistency between
			// DELETE and LIST operationsCheck that the object is not deleted,
			// so filter out any objects with a non-zero time-deleted
			if object.Deleted.IsZero() {
				name := object.Name
				// Ignore objects with names that end with '#' (these are uploaded parts)
				if name[len(name)-1] != '#' {
					name = d.keyToPath(name)
					list = append(list, name)
				}
			}
		}
		for _, subpath := range objects.Prefixes {
			subpath = d.keyToPath(subpath)
			list = append(list, subpath)
		}
		query = objects.Next
		if query == nil {
			break
		}
	}
	if path != "/" && len(list) == 0 {
		// Treat empty response as missing directory, since we don't actually
		// have directories in Google Cloud Storage.
		return nil, storagedriver.PathNotFoundError{Path: path}
	}
	return list, nil
}

// Move moves an object stored at sourcePath to destPath, removing the
// original object.
func (d *driver) Move(context ctx.Context, sourcePath string, destPath string) error {
	prefix := d.pathToDirKey(sourcePath)
	gcsContext := d.context(context)
	keys, err := d.listAll(gcsContext, prefix)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		destPrefix := d.pathToDirKey(destPath)
		copies := make([]string, 0, len(keys))
		sort.Strings(keys)
		var err error
		for _, key := range keys {
			dest := destPrefix + key[len(prefix):]
			_, err = storageCopyObject(gcsContext, d.bucket, key, d.bucket, dest, nil)
			if err == nil {
				copies = append(copies, dest)
			} else {
				break
			}
		}
		// if an error occurred, attempt to cleanup the copies made
		if err != nil {
			for i := len(copies) - 1; i >= 0; i-- {
				_ = storageDeleteObject(gcsContext, d.bucket, copies[i])
			}
			return err
		}
		// delete originals
		for i := len(keys) - 1; i >= 0; i-- {
			err2 := storageDeleteObject(gcsContext, d.bucket, keys[i])
			if err2 != nil {
				err = err2
			}
		}
		return err
	}
	_, err = storageCopyObject(gcsContext, d.bucket, d.pathToKey(sourcePath), d.bucket, d.pathToKey(destPath), nil)
	if err != nil {
		if status := err.(*googleapi.Error); status != nil {
			if status.Code == http.StatusNotFound {
				return storagedriver.PathNotFoundError{Path: sourcePath}
			}
		}
		return err
	}
	return storageDeleteObject(gcsContext, d.bucket, d.pathToKey(sourcePath))
}

// listAll recursively lists all names of objects stored at "prefix" and its subpaths.
func (d *driver) listAll(context context.Context, prefix string) ([]string, error) {
	list := make([]string, 0, 64)
	query := &storage.Query{}
	query.Prefix = prefix
	query.Versions = false
	for {
		objects, err := storageListObjects(d.context(context), d.bucket, query)
		if err != nil {
			return nil, err
		}
		for _, obj := range objects.Results {
			// GCS does not guarantee strong consistency between
			// DELETE and LIST operationsCheck that the object is not deleted,
			// so filter out any objects with a non-zero time-deleted
			if obj.Deleted.IsZero() {
				list = append(list, obj.Name)
			}
		}
		query = objects.Next
		if query == nil {
			break
		}
	}
	return list, nil
}

// Delete recursively deletes all objects stored at "path" and its subpaths.
func (d *driver) Delete(context ctx.Context, path string) error {
	prefix := d.pathToDirKey(path)
	gcsContext := d.context(context)
	keys, err := d.listAll(gcsContext, prefix)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		sort.Sort(sort.Reverse(sort.StringSlice(keys)))
		for _, key := range keys {
			err := storageDeleteObject(gcsContext, d.bucket, key)
			// GCS only guarantees eventual consistency, so listAll might return
			// paths that no longer exist. If this happens, just ignore any not
			// found error
			if status, ok := err.(*googleapi.Error); ok {
				if status.Code == http.StatusNotFound {
					err = nil
				}
			}
			if err != nil {
				return err
			}
		}
		return nil
	}
	err = storageDeleteObject(gcsContext, d.bucket, d.pathToKey(path))
	if err != nil {
		if status := err.(*googleapi.Error); status != nil {
			if status.Code == http.StatusNotFound {
				return storagedriver.PathNotFoundError{Path: path}
			}
		}
	}
	return err
}

func storageDeleteObject(context context.Context, bucket string, name string) error {
	return retry(5, func() error {
		return storage.DeleteObject(context, bucket, name)
	})
}

func storageStatObject(context context.Context, bucket string, name string) (*storage.Object, error) {
	var obj *storage.Object
	err := retry(5, func() error {
		var err error
		obj, err = storage.StatObject(context, bucket, name)
		return err
	})
	return obj, err
}

func storageListObjects(context context.Context, bucket string, q *storage.Query) (*storage.Objects, error) {
	var objs *storage.Objects
	err := retry(5, func() error {
		var err error
		objs, err = storage.ListObjects(context, bucket, q)
		return err
	})
	return objs, err
}

func storageCopyObject(context context.Context, srcBucket, srcName string, destBucket, destName string, attrs *storage.ObjectAttrs) (*storage.Object, error) {
	var obj *storage.Object
	err := retry(5, func() error {
		var err error
		obj, err = storage.CopyObject(context, srcBucket, srcName, destBucket, destName, attrs)
		return err
	})
	return obj, err
}

// URLFor returns a URL which may be used to retrieve the content stored at
// the given path, possibly using the given options.
// Returns ErrUnsupportedMethod if this driver has no privateKey
func (d *driver) URLFor(context ctx.Context, path string, options map[string]interface{}) (string, error) {
	if d.privateKey == nil {
		return "", storagedriver.ErrUnsupportedMethod{}
	}

	name := d.pathToKey(path)
	methodString := "GET"
	method, ok := options["method"]
	if ok {
		methodString, ok = method.(string)
		if !ok || (methodString != "GET" && methodString != "HEAD") {
			return "", storagedriver.ErrUnsupportedMethod{}
		}
	}

	expiresTime := time.Now().Add(20 * time.Minute)
	expires, ok := options["expiry"]
	if ok {
		et, ok := expires.(time.Time)
		if ok {
			expiresTime = et
		}
	}

	opts := &storage.SignedURLOptions{
		GoogleAccessID: d.email,
		PrivateKey:     d.privateKey,
		Method:         methodString,
		Expires:        expiresTime,
	}
	return storage.SignedURL(d.bucket, name, opts)
}

func (d *driver) context(context ctx.Context) context.Context {
	return cloud.WithContext(context, dummyProjectID, d.client)
}

func (d *driver) pathToKey(path string) string {
	return strings.TrimRight(d.rootDirectory+strings.TrimLeft(path, "/"), "/")
}

func (d *driver) pathToDirKey(path string) string {
	return d.pathToKey(path) + "/"
}

func (d *driver) keyToPath(key string) string {
	return "/" + strings.Trim(strings.TrimPrefix(key, d.rootDirectory), "/")
}
