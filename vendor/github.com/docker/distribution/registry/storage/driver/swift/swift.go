// Package swift provides a storagedriver.StorageDriver implementation to
// store blobs in Openstack Swift object storage.
//
// This package leverages the ncw/swift client library for interfacing with
// Swift.
//
// It supports both TempAuth authentication and Keystone authentication
// (up to version 3).
//
// As Swift has a limit on the size of a single uploaded object (by default
// this is 5GB), the driver makes use of the Swift Large Object Support
// (http://docs.openstack.org/developer/swift/overview_large_objects.html).
// Only one container is used for both manifests and data objects. Manifests
// are stored in the 'files' pseudo directory, data objects are stored under
// 'segments'.
package swift

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/ncw/swift"

	"github.com/docker/distribution/context"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/base"
	"github.com/docker/distribution/registry/storage/driver/factory"
	"github.com/docker/distribution/version"
)

const driverName = "swift"

// defaultChunkSize defines the default size of a segment
const defaultChunkSize = 20 * 1024 * 1024

// minChunkSize defines the minimum size of a segment
const minChunkSize = 1 << 20

// readAfterWriteTimeout defines the time we wait before an object appears after having been uploaded
var readAfterWriteTimeout = 15 * time.Second

// readAfterWriteWait defines the time to sleep between two retries
var readAfterWriteWait = 200 * time.Millisecond

// Parameters A struct that encapsulates all of the driver parameters after all values have been set
type Parameters struct {
	Username            string
	Password            string
	AuthURL             string
	Tenant              string
	TenantID            string
	Domain              string
	DomainID            string
	TrustID             string
	Region              string
	Container           string
	Prefix              string
	InsecureSkipVerify  bool
	ChunkSize           int
	SecretKey           string
	AccessKey           string
	TempURLContainerKey bool
	TempURLMethods      []string
}

// swiftInfo maps the JSON structure returned by Swift /info endpoint
type swiftInfo struct {
	Swift struct {
		Version string `mapstructure:"version"`
	}
	Tempurl struct {
		Methods []string `mapstructure:"methods"`
	}
}

func init() {
	factory.Register(driverName, &swiftDriverFactory{})
}

// swiftDriverFactory implements the factory.StorageDriverFactory interface
type swiftDriverFactory struct{}

func (factory *swiftDriverFactory) Create(parameters map[string]interface{}) (storagedriver.StorageDriver, error) {
	return FromParameters(parameters)
}

type driver struct {
	Conn                swift.Connection
	Container           string
	Prefix              string
	BulkDeleteSupport   bool
	ChunkSize           int
	SecretKey           string
	AccessKey           string
	TempURLContainerKey bool
	TempURLMethods      []string
}

type baseEmbed struct {
	base.Base
}

// Driver is a storagedriver.StorageDriver implementation backed by Openstack Swift
// Objects are stored at absolute keys in the provided container.
type Driver struct {
	baseEmbed
}

// FromParameters constructs a new Driver with a given parameters map
// Required parameters:
// - username
// - password
// - authurl
// - container
func FromParameters(parameters map[string]interface{}) (*Driver, error) {
	params := Parameters{
		ChunkSize:          defaultChunkSize,
		InsecureSkipVerify: false,
	}

	if err := mapstructure.Decode(parameters, &params); err != nil {
		return nil, err
	}

	if params.Username == "" {
		return nil, fmt.Errorf("No username parameter provided")
	}

	if params.Password == "" {
		return nil, fmt.Errorf("No password parameter provided")
	}

	if params.AuthURL == "" {
		return nil, fmt.Errorf("No authurl parameter provided")
	}

	if params.Container == "" {
		return nil, fmt.Errorf("No container parameter provided")
	}

	if params.ChunkSize < minChunkSize {
		return nil, fmt.Errorf("The chunksize %#v parameter should be a number that is larger than or equal to %d", params.ChunkSize, minChunkSize)
	}

	return New(params)
}

// New constructs a new Driver with the given Openstack Swift credentials and container name
func New(params Parameters) (*Driver, error) {
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConnsPerHost: 2048,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: params.InsecureSkipVerify},
	}

	ct := swift.Connection{
		UserName:       params.Username,
		ApiKey:         params.Password,
		AuthUrl:        params.AuthURL,
		Region:         params.Region,
		UserAgent:      "distribution/" + version.Version,
		Tenant:         params.Tenant,
		TenantId:       params.TenantID,
		Domain:         params.Domain,
		DomainId:       params.DomainID,
		TrustId:        params.TrustID,
		Transport:      transport,
		ConnectTimeout: 60 * time.Second,
		Timeout:        15 * 60 * time.Second,
	}
	err := ct.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("Swift authentication failed: %s", err)
	}

	if _, _, err := ct.Container(params.Container); err == swift.ContainerNotFound {
		if err := ct.ContainerCreate(params.Container, nil); err != nil {
			return nil, fmt.Errorf("Failed to create container %s (%s)", params.Container, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("Failed to retrieve info about container %s (%s)", params.Container, err)
	}

	d := &driver{
		Conn:           ct,
		Container:      params.Container,
		Prefix:         params.Prefix,
		ChunkSize:      params.ChunkSize,
		TempURLMethods: make([]string, 0),
		AccessKey:      params.AccessKey,
	}

	info := swiftInfo{}
	if config, err := d.Conn.QueryInfo(); err == nil {
		_, d.BulkDeleteSupport = config["bulk_delete"]

		if err := mapstructure.Decode(config, &info); err == nil {
			d.TempURLContainerKey = info.Swift.Version >= "2.3.0"
			d.TempURLMethods = info.Tempurl.Methods
		}
	} else {
		d.TempURLContainerKey = params.TempURLContainerKey
		d.TempURLMethods = params.TempURLMethods
	}

	if len(d.TempURLMethods) > 0 {
		secretKey := params.SecretKey
		if secretKey == "" {
			secretKey, _ = generateSecret()
		}

		// Since Swift 2.2.2, we can now set secret keys on containers
		// in addition to the account secret keys. Use them in preference.
		if d.TempURLContainerKey {
			_, containerHeaders, err := d.Conn.Container(d.Container)
			if err != nil {
				return nil, fmt.Errorf("Failed to fetch container info %s (%s)", d.Container, err)
			}

			d.SecretKey = containerHeaders["X-Container-Meta-Temp-Url-Key"]
			if d.SecretKey == "" || (params.SecretKey != "" && d.SecretKey != params.SecretKey) {
				m := swift.Metadata{}
				m["temp-url-key"] = secretKey
				if d.Conn.ContainerUpdate(d.Container, m.ContainerHeaders()); err == nil {
					d.SecretKey = secretKey
				}
			}
		} else {
			// Use the account secret key
			_, accountHeaders, err := d.Conn.Account()
			if err != nil {
				return nil, fmt.Errorf("Failed to fetch account info (%s)", err)
			}

			d.SecretKey = accountHeaders["X-Account-Meta-Temp-Url-Key"]
			if d.SecretKey == "" || (params.SecretKey != "" && d.SecretKey != params.SecretKey) {
				m := swift.Metadata{}
				m["temp-url-key"] = secretKey
				if err := d.Conn.AccountUpdate(m.AccountHeaders()); err == nil {
					d.SecretKey = secretKey
				}
			}
		}
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
	content, err := d.Conn.ObjectGetBytes(d.Container, d.swiftPath(path))
	if err == swift.ObjectNotFound {
		return nil, storagedriver.PathNotFoundError{Path: path}
	}
	return content, nil
}

// PutContent stores the []byte content at a location designated by "path".
func (d *driver) PutContent(ctx context.Context, path string, contents []byte) error {
	err := d.Conn.ObjectPutBytes(d.Container, d.swiftPath(path), contents, d.getContentType())
	if err == swift.ObjectNotFound {
		return storagedriver.PathNotFoundError{Path: path}
	}
	return err
}

// ReadStream retrieves an io.ReadCloser for the content stored at "path" with a
// given byte offset.
func (d *driver) ReadStream(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	headers := make(swift.Headers)
	headers["Range"] = "bytes=" + strconv.FormatInt(offset, 10) + "-"

	file, _, err := d.Conn.ObjectOpen(d.Container, d.swiftPath(path), false, headers)
	if err == swift.ObjectNotFound {
		return nil, storagedriver.PathNotFoundError{Path: path}
	}
	if swiftErr, ok := err.(*swift.Error); ok && swiftErr.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		return ioutil.NopCloser(bytes.NewReader(nil)), nil
	}
	return file, err
}

// WriteStream stores the contents of the provided io.Reader at a
// location designated by the given path. The driver will know it has
// received the full contents when the reader returns io.EOF. The number
// of successfully READ bytes will be returned, even if an error is
// returned. May be used to resume writing a stream by providing a nonzero
// offset. Offsets past the current size will write from the position
// beyond the end of the file.
func (d *driver) WriteStream(ctx context.Context, path string, offset int64, reader io.Reader) (int64, error) {
	var (
		segments      []swift.Object
		multi         io.Reader
		paddingReader io.Reader
		currentLength int64
		cursor        int64
		segmentPath   string
	)

	partNumber := 1
	chunkSize := int64(d.ChunkSize)
	zeroBuf := make([]byte, d.ChunkSize)
	hash := md5.New()

	getSegment := func() string {
		return fmt.Sprintf("%s/%016d", segmentPath, partNumber)
	}

	max := func(a int64, b int64) int64 {
		if a > b {
			return a
		}
		return b
	}

	createManifest := true
	info, headers, err := d.Conn.Object(d.Container, d.swiftPath(path))
	if err == nil {
		manifest, ok := headers["X-Object-Manifest"]
		if !ok {
			if segmentPath, err = d.swiftSegmentPath(path); err != nil {
				return 0, err
			}
			if err := d.Conn.ObjectMove(d.Container, d.swiftPath(path), d.Container, getSegment()); err != nil {
				return 0, err
			}
			segments = append(segments, info)
		} else {
			_, segmentPath = parseManifest(manifest)
			if segments, err = d.getAllSegments(segmentPath); err != nil {
				return 0, err
			}
			createManifest = false
		}
		currentLength = info.Bytes
	} else if err == swift.ObjectNotFound {
		if segmentPath, err = d.swiftSegmentPath(path); err != nil {
			return 0, err
		}
	} else {
		return 0, err
	}

	// First, we skip the existing segments that are not modified by this call
	for i := range segments {
		if offset < cursor+segments[i].Bytes {
			break
		}
		cursor += segments[i].Bytes
		hash.Write([]byte(segments[i].Hash))
		partNumber++
	}

	// We reached the end of the file but we haven't reached 'offset' yet
	// Therefore we add blocks of zeros
	if offset >= currentLength {
		for offset-currentLength >= chunkSize {
			// Insert a block a zero
			headers, err := d.Conn.ObjectPut(d.Container, getSegment(), bytes.NewReader(zeroBuf), false, "", d.getContentType(), nil)
			if err != nil {
				if err == swift.ObjectNotFound {
					return 0, storagedriver.PathNotFoundError{Path: getSegment()}
				}
				return 0, err
			}
			currentLength += chunkSize
			partNumber++
			hash.Write([]byte(headers["Etag"]))
		}

		cursor = currentLength
		paddingReader = bytes.NewReader(zeroBuf)
	} else if offset-cursor > 0 {
		// Offset is inside the current segment : we need to read the
		// data from the beginning of the segment to offset
		file, _, err := d.Conn.ObjectOpen(d.Container, getSegment(), false, nil)
		if err != nil {
			if err == swift.ObjectNotFound {
				return 0, storagedriver.PathNotFoundError{Path: getSegment()}
			}
			return 0, err
		}
		defer file.Close()
		paddingReader = file
	}

	readers := []io.Reader{}
	if paddingReader != nil {
		readers = append(readers, io.LimitReader(paddingReader, offset-cursor))
	}
	readers = append(readers, io.LimitReader(reader, chunkSize-(offset-cursor)))
	multi = io.MultiReader(readers...)

	writeSegment := func(segment string) (finished bool, bytesRead int64, err error) {
		currentSegment, err := d.Conn.ObjectCreate(d.Container, segment, false, "", d.getContentType(), nil)
		if err != nil {
			if err == swift.ObjectNotFound {
				return false, bytesRead, storagedriver.PathNotFoundError{Path: segment}
			}
			return false, bytesRead, err
		}

		segmentHash := md5.New()
		writer := io.MultiWriter(currentSegment, segmentHash)

		n, err := io.Copy(writer, multi)
		if err != nil {
			return false, bytesRead, err
		}

		if n > 0 {
			defer func() {
				closeError := currentSegment.Close()
				if err != nil {
					err = closeError
				}
				hexHash := hex.EncodeToString(segmentHash.Sum(nil))
				hash.Write([]byte(hexHash))
			}()
			bytesRead += n - max(0, offset-cursor)
		}

		if n < chunkSize {
			// We wrote all the data
			if cursor+n < currentLength {
				// Copy the end of the chunk
				headers := make(swift.Headers)
				headers["Range"] = "bytes=" + strconv.FormatInt(cursor+n, 10) + "-" + strconv.FormatInt(cursor+chunkSize, 10)
				file, _, err := d.Conn.ObjectOpen(d.Container, d.swiftPath(path), false, headers)
				if err != nil {
					if err == swift.ObjectNotFound {
						return false, bytesRead, storagedriver.PathNotFoundError{Path: path}
					}
					return false, bytesRead, err
				}

				_, copyErr := io.Copy(writer, file)

				if err := file.Close(); err != nil {
					if err == swift.ObjectNotFound {
						return false, bytesRead, storagedriver.PathNotFoundError{Path: path}
					}
					return false, bytesRead, err
				}

				if copyErr != nil {
					return false, bytesRead, copyErr
				}
			}

			return true, bytesRead, nil
		}

		multi = io.LimitReader(reader, chunkSize)
		cursor += chunkSize
		partNumber++

		return false, bytesRead, nil
	}

	finished := false
	read := int64(0)
	bytesRead := int64(0)
	for finished == false {
		finished, read, err = writeSegment(getSegment())
		bytesRead += read
		if err != nil {
			return bytesRead, err
		}
	}

	for ; partNumber < len(segments); partNumber++ {
		hash.Write([]byte(segments[partNumber].Hash))
	}

	if createManifest {
		if err := d.createManifest(path, d.Container+"/"+segmentPath); err != nil {
			return 0, err
		}
	}

	expectedHash := hex.EncodeToString(hash.Sum(nil))
	waitingTime := readAfterWriteWait
	endTime := time.Now().Add(readAfterWriteTimeout)
	for {
		var infos swift.Object
		if infos, _, err = d.Conn.Object(d.Container, d.swiftPath(path)); err == nil {
			if strings.Trim(infos.Hash, "\"") == expectedHash {
				return bytesRead, nil
			}
			err = fmt.Errorf("Timeout expired while waiting for segments of %s to show up", path)
		}
		if time.Now().Add(waitingTime).After(endTime) {
			break
		}
		time.Sleep(waitingTime)
		waitingTime *= 2
	}

	return bytesRead, err
}

// Stat retrieves the FileInfo for the given path, including the current size
// in bytes and the creation time.
func (d *driver) Stat(ctx context.Context, path string) (storagedriver.FileInfo, error) {
	swiftPath := d.swiftPath(path)
	opts := &swift.ObjectsOpts{
		Prefix:    swiftPath,
		Delimiter: '/',
	}

	objects, err := d.Conn.ObjectsAll(d.Container, opts)
	if err != nil {
		if err == swift.ContainerNotFound {
			return nil, storagedriver.PathNotFoundError{Path: path}
		}
		return nil, err
	}

	fi := storagedriver.FileInfoFields{
		Path: strings.TrimPrefix(strings.TrimSuffix(swiftPath, "/"), d.swiftPath("/")),
	}

	for _, obj := range objects {
		if obj.PseudoDirectory && obj.Name == swiftPath+"/" {
			fi.IsDir = true
			return storagedriver.FileInfoInternal{FileInfoFields: fi}, nil
		} else if obj.Name == swiftPath {
			// On Swift 1.12, the 'bytes' field is always 0
			// so we need to do a second HEAD request
			info, _, err := d.Conn.Object(d.Container, swiftPath)
			if err != nil {
				if err == swift.ObjectNotFound {
					return nil, storagedriver.PathNotFoundError{Path: path}
				}
				return nil, err
			}
			fi.IsDir = false
			fi.Size = info.Bytes
			fi.ModTime = info.LastModified
			return storagedriver.FileInfoInternal{FileInfoFields: fi}, nil
		}
	}

	return nil, storagedriver.PathNotFoundError{Path: path}
}

// List returns a list of the objects that are direct descendants of the given path.
func (d *driver) List(ctx context.Context, path string) ([]string, error) {
	var files []string

	prefix := d.swiftPath(path)
	if prefix != "" {
		prefix += "/"
	}

	opts := &swift.ObjectsOpts{
		Prefix:    prefix,
		Delimiter: '/',
	}

	objects, err := d.Conn.ObjectsAll(d.Container, opts)
	for _, obj := range objects {
		files = append(files, strings.TrimPrefix(strings.TrimSuffix(obj.Name, "/"), d.swiftPath("/")))
	}

	if err == swift.ContainerNotFound || (len(objects) == 0 && path != "/") {
		return files, storagedriver.PathNotFoundError{Path: path}
	}
	return files, err
}

// Move moves an object stored at sourcePath to destPath, removing the original
// object.
func (d *driver) Move(ctx context.Context, sourcePath string, destPath string) error {
	_, headers, err := d.Conn.Object(d.Container, d.swiftPath(sourcePath))
	if err == nil {
		if manifest, ok := headers["X-Object-Manifest"]; ok {
			if err = d.createManifest(destPath, manifest); err != nil {
				return err
			}
			err = d.Conn.ObjectDelete(d.Container, d.swiftPath(sourcePath))
		} else {
			err = d.Conn.ObjectMove(d.Container, d.swiftPath(sourcePath), d.Container, d.swiftPath(destPath))
		}
	}
	if err == swift.ObjectNotFound {
		return storagedriver.PathNotFoundError{Path: sourcePath}
	}
	return err
}

// Delete recursively deletes all objects stored at "path" and its subpaths.
func (d *driver) Delete(ctx context.Context, path string) error {
	opts := swift.ObjectsOpts{
		Prefix: d.swiftPath(path) + "/",
	}

	objects, err := d.Conn.ObjectsAll(d.Container, &opts)
	if err != nil {
		if err == swift.ContainerNotFound {
			return storagedriver.PathNotFoundError{Path: path}
		}
		return err
	}

	for _, obj := range objects {
		if obj.PseudoDirectory {
			continue
		}
		if _, headers, err := d.Conn.Object(d.Container, obj.Name); err == nil {
			manifest, ok := headers["X-Object-Manifest"]
			if ok {
				_, prefix := parseManifest(manifest)
				segments, err := d.getAllSegments(prefix)
				if err != nil {
					return err
				}
				objects = append(objects, segments...)
			}
		} else {
			if err == swift.ObjectNotFound {
				return storagedriver.PathNotFoundError{Path: obj.Name}
			}
			return err
		}
	}

	if d.BulkDeleteSupport && len(objects) > 0 {
		filenames := make([]string, len(objects))
		for i, obj := range objects {
			filenames[i] = obj.Name
		}
		_, err = d.Conn.BulkDelete(d.Container, filenames)
		// Don't fail on ObjectNotFound because eventual consistency
		// makes this situation normal.
		if err != nil && err != swift.Forbidden && err != swift.ObjectNotFound {
			if err == swift.ContainerNotFound {
				return storagedriver.PathNotFoundError{Path: path}
			}
			return err
		}
	} else {
		for _, obj := range objects {
			if err := d.Conn.ObjectDelete(d.Container, obj.Name); err != nil {
				if err == swift.ObjectNotFound {
					return storagedriver.PathNotFoundError{Path: obj.Name}
				}
				return err
			}
		}
	}

	_, _, err = d.Conn.Object(d.Container, d.swiftPath(path))
	if err == nil {
		if err := d.Conn.ObjectDelete(d.Container, d.swiftPath(path)); err != nil {
			if err == swift.ObjectNotFound {
				return storagedriver.PathNotFoundError{Path: path}
			}
			return err
		}
	} else if err == swift.ObjectNotFound {
		if len(objects) == 0 {
			return storagedriver.PathNotFoundError{Path: path}
		}
	} else {
		return err
	}
	return nil
}

// URLFor returns a URL which may be used to retrieve the content stored at the given path.
func (d *driver) URLFor(ctx context.Context, path string, options map[string]interface{}) (string, error) {
	if d.SecretKey == "" {
		return "", storagedriver.ErrUnsupportedMethod{}
	}

	methodString := "GET"
	method, ok := options["method"]
	if ok {
		if methodString, ok = method.(string); !ok {
			return "", storagedriver.ErrUnsupportedMethod{}
		}
	}

	if methodString == "HEAD" {
		// A "HEAD" request on a temporary URL is allowed if the
		// signature was generated with "GET", "POST" or "PUT"
		methodString = "GET"
	}

	supported := false
	for _, method := range d.TempURLMethods {
		if method == methodString {
			supported = true
			break
		}
	}

	if !supported {
		return "", storagedriver.ErrUnsupportedMethod{}
	}

	expiresTime := time.Now().Add(20 * time.Minute)
	expires, ok := options["expiry"]
	if ok {
		et, ok := expires.(time.Time)
		if ok {
			expiresTime = et
		}
	}

	tempURL := d.Conn.ObjectTempUrl(d.Container, d.swiftPath(path), d.SecretKey, methodString, expiresTime)

	if d.AccessKey != "" {
		// On HP Cloud, the signature must be in the form of tenant_id:access_key:signature
		url, _ := url.Parse(tempURL)
		query := url.Query()
		query.Set("temp_url_sig", fmt.Sprintf("%s:%s:%s", d.Conn.TenantId, d.AccessKey, query.Get("temp_url_sig")))
		url.RawQuery = query.Encode()
		tempURL = url.String()
	}

	return tempURL, nil
}

func (d *driver) swiftPath(path string) string {
	return strings.TrimLeft(strings.TrimRight(d.Prefix+"/files"+path, "/"), "/")
}

func (d *driver) swiftSegmentPath(path string) (string, error) {
	checksum := sha1.New()
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	path = hex.EncodeToString(checksum.Sum(append([]byte(path), random...)))
	return strings.TrimLeft(strings.TrimRight(d.Prefix+"/segments/"+path[0:3]+"/"+path[3:], "/"), "/"), nil
}

func (d *driver) getContentType() string {
	return "application/octet-stream"
}

func (d *driver) getAllSegments(path string) ([]swift.Object, error) {
	segments, err := d.Conn.ObjectsAll(d.Container, &swift.ObjectsOpts{Prefix: path})
	if err == swift.ContainerNotFound {
		return nil, storagedriver.PathNotFoundError{Path: path}
	}
	return segments, err
}

func (d *driver) createManifest(path string, segments string) error {
	headers := make(swift.Headers)
	headers["X-Object-Manifest"] = segments
	manifest, err := d.Conn.ObjectCreate(d.Container, d.swiftPath(path), false, "", d.getContentType(), headers)
	if err != nil {
		if err == swift.ObjectNotFound {
			return storagedriver.PathNotFoundError{Path: path}
		}
		return err
	}
	if err := manifest.Close(); err != nil {
		if err == swift.ObjectNotFound {
			return storagedriver.PathNotFoundError{Path: path}
		}
		return err
	}
	return nil
}

func parseManifest(manifest string) (container string, prefix string) {
	components := strings.SplitN(manifest, "/", 2)
	container = components[0]
	if len(components) > 1 {
		prefix = components[1]
	}
	return container, prefix
}

func generateSecret() (string, error) {
	var secretBytes [32]byte
	if _, err := rand.Read(secretBytes[:]); err != nil {
		return "", fmt.Errorf("could not generate random bytes for Swift secret key: %v", err)
	}
	return hex.EncodeToString(secretBytes[:]), nil
}
