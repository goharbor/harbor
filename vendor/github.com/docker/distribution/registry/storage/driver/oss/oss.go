// Package oss provides a storagedriver.StorageDriver implementation to
// store blobs in Aliyun OSS cloud storage.
//
// This package leverages the denverdino/aliyungo client library for interfacing with
// oss.
//
// Because OSS is a key, value store the Stat call does not support last modification
// time for directories (directories are an abstraction for key, value stores)
//
// +build include_oss

package oss

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/distribution/context"

	"github.com/Sirupsen/logrus"
	"github.com/denverdino/aliyungo/oss"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/base"
	"github.com/docker/distribution/registry/storage/driver/factory"
)

const driverName = "oss"

// minChunkSize defines the minimum multipart upload chunk size
// OSS API requires multipart upload chunks to be at least 5MB
const minChunkSize = 5 << 20

const defaultChunkSize = 2 * minChunkSize
const defaultTimeout = 2 * time.Minute // 2 minute timeout per chunk

// listMax is the largest amount of objects you can request from OSS in a list call
const listMax = 1000

//DriverParameters A struct that encapsulates all of the driver parameters after all values have been set
type DriverParameters struct {
	AccessKeyID     string
	AccessKeySecret string
	Bucket          string
	Region          oss.Region
	Internal        bool
	Encrypt         bool
	Secure          bool
	ChunkSize       int64
	RootDirectory   string
	Endpoint        string
}

func init() {
	factory.Register(driverName, &ossDriverFactory{})
}

// ossDriverFactory implements the factory.StorageDriverFactory interface
type ossDriverFactory struct{}

func (factory *ossDriverFactory) Create(parameters map[string]interface{}) (storagedriver.StorageDriver, error) {
	return FromParameters(parameters)
}

type driver struct {
	Client        *oss.Client
	Bucket        *oss.Bucket
	ChunkSize     int64
	Encrypt       bool
	RootDirectory string

	pool  sync.Pool // pool []byte buffers used for WriteStream
	zeros []byte    // shared, zero-valued buffer used for WriteStream
}

type baseEmbed struct {
	base.Base
}

// Driver is a storagedriver.StorageDriver implementation backed by Aliyun OSS
// Objects are stored at absolute keys in the provided bucket.
type Driver struct {
	baseEmbed
}

// FromParameters constructs a new Driver with a given parameters map
// Required parameters:
// - accesskey
// - secretkey
// - region
// - bucket
// - encrypt
func FromParameters(parameters map[string]interface{}) (*Driver, error) {
	// Providing no values for these is valid in case the user is authenticating
	// with an IAM on an ec2 instance (in which case the instance credentials will
	// be summoned when GetAuth is called)
	accessKey, ok := parameters["accesskeyid"]
	if !ok {
		return nil, fmt.Errorf("No accesskeyid parameter provided")
	}
	secretKey, ok := parameters["accesskeysecret"]
	if !ok {
		return nil, fmt.Errorf("No accesskeysecret parameter provided")
	}

	regionName, ok := parameters["region"]
	if !ok || fmt.Sprint(regionName) == "" {
		return nil, fmt.Errorf("No region parameter provided")
	}

	bucket, ok := parameters["bucket"]
	if !ok || fmt.Sprint(bucket) == "" {
		return nil, fmt.Errorf("No bucket parameter provided")
	}

	internalBool := false
	internal, ok := parameters["internal"]
	if ok {
		internalBool, ok = internal.(bool)
		if !ok {
			return nil, fmt.Errorf("The internal parameter should be a boolean")
		}
	}

	encryptBool := false
	encrypt, ok := parameters["encrypt"]
	if ok {
		encryptBool, ok = encrypt.(bool)
		if !ok {
			return nil, fmt.Errorf("The encrypt parameter should be a boolean")
		}
	}

	secureBool := true
	secure, ok := parameters["secure"]
	if ok {
		secureBool, ok = secure.(bool)
		if !ok {
			return nil, fmt.Errorf("The secure parameter should be a boolean")
		}
	}

	chunkSize := int64(defaultChunkSize)
	chunkSizeParam, ok := parameters["chunksize"]
	if ok {
		switch v := chunkSizeParam.(type) {
		case string:
			vv, err := strconv.ParseInt(v, 0, 64)
			if err != nil {
				return nil, fmt.Errorf("chunksize parameter must be an integer, %v invalid", chunkSizeParam)
			}
			chunkSize = vv
		case int64:
			chunkSize = v
		case int, uint, int32, uint32, uint64:
			chunkSize = reflect.ValueOf(v).Convert(reflect.TypeOf(chunkSize)).Int()
		default:
			return nil, fmt.Errorf("invalid valud for chunksize: %#v", chunkSizeParam)
		}

		if chunkSize < minChunkSize {
			return nil, fmt.Errorf("The chunksize %#v parameter should be a number that is larger than or equal to %d", chunkSize, minChunkSize)
		}
	}

	rootDirectory, ok := parameters["rootdirectory"]
	if !ok {
		rootDirectory = ""
	}

	endpoint, ok := parameters["endpoint"]
	if !ok {
		endpoint = ""
	}

	params := DriverParameters{
		AccessKeyID:     fmt.Sprint(accessKey),
		AccessKeySecret: fmt.Sprint(secretKey),
		Bucket:          fmt.Sprint(bucket),
		Region:          oss.Region(fmt.Sprint(regionName)),
		ChunkSize:       chunkSize,
		RootDirectory:   fmt.Sprint(rootDirectory),
		Encrypt:         encryptBool,
		Secure:          secureBool,
		Internal:        internalBool,
		Endpoint:        fmt.Sprint(endpoint),
	}

	return New(params)
}

// New constructs a new Driver with the given Aliyun credentials, region, encryption flag, and
// bucketName
func New(params DriverParameters) (*Driver, error) {

	client := oss.NewOSSClient(params.Region, params.Internal, params.AccessKeyID, params.AccessKeySecret, params.Secure)
	client.SetEndpoint(params.Endpoint)
	bucket := client.Bucket(params.Bucket)
	client.SetDebug(false)

	// Validate that the given credentials have at least read permissions in the
	// given bucket scope.
	if _, err := bucket.List(strings.TrimRight(params.RootDirectory, "/"), "", "", 1); err != nil {
		return nil, err
	}

	// TODO(tg123): Currently multipart uploads have no timestamps, so this would be unwise
	// if you initiated a new OSS client while another one is running on the same bucket.

	d := &driver{
		Client:        client,
		Bucket:        bucket,
		ChunkSize:     params.ChunkSize,
		Encrypt:       params.Encrypt,
		RootDirectory: params.RootDirectory,
		zeros:         make([]byte, params.ChunkSize),
	}

	d.pool.New = func() interface{} {
		return make([]byte, d.ChunkSize)
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
	content, err := d.Bucket.Get(d.ossPath(path))
	if err != nil {
		return nil, parseError(path, err)
	}
	return content, nil
}

// PutContent stores the []byte content at a location designated by "path".
func (d *driver) PutContent(ctx context.Context, path string, contents []byte) error {
	return parseError(path, d.Bucket.Put(d.ossPath(path), contents, d.getContentType(), getPermissions(), d.getOptions()))
}

// ReadStream retrieves an io.ReadCloser for the content stored at "path" with a
// given byte offset.
func (d *driver) ReadStream(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	headers := make(http.Header)
	headers.Add("Range", "bytes="+strconv.FormatInt(offset, 10)+"-")

	resp, err := d.Bucket.GetResponseWithHeaders(d.ossPath(path), headers)
	if err != nil {
		return nil, parseError(path, err)
	}

	// Due to Aliyun OSS API, status 200 and whole object will be return instead of an
	// InvalidRange error when range is invalid.
	//
	// OSS sever will always return http.StatusPartialContent if range is acceptable.
	if resp.StatusCode != http.StatusPartialContent {
		resp.Body.Close()
		return ioutil.NopCloser(bytes.NewReader(nil)), nil
	}

	return resp.Body, nil
}

// WriteStream stores the contents of the provided io.Reader at a
// location designated by the given path. The driver will know it has
// received the full contents when the reader returns io.EOF. The number
// of successfully READ bytes will be returned, even if an error is
// returned. May be used to resume writing a stream by providing a nonzero
// offset. Offsets past the current size will write from the position
// beyond the end of the file.
func (d *driver) WriteStream(ctx context.Context, path string, offset int64, reader io.Reader) (totalRead int64, err error) {
	partNumber := 1
	bytesRead := 0
	var putErrChan chan error
	parts := []oss.Part{}
	var part oss.Part
	done := make(chan struct{}) // stopgap to free up waiting goroutines

	multi, err := d.Bucket.InitMulti(d.ossPath(path), d.getContentType(), getPermissions(), d.getOptions())
	if err != nil {
		return 0, err
	}

	buf := d.getbuf()

	// We never want to leave a dangling multipart upload, our only consistent state is
	// when there is a whole object at path. This is in order to remain consistent with
	// the stat call.
	//
	// Note that if the machine dies before executing the defer, we will be left with a dangling
	// multipart upload, which will eventually be cleaned up, but we will lose all of the progress
	// made prior to the machine crashing.
	defer func() {
		if putErrChan != nil {
			if putErr := <-putErrChan; putErr != nil {
				err = putErr
			}
		}

		if len(parts) > 0 {
			if multi == nil {
				// Parts should be empty if the multi is not initialized
				panic("Unreachable")
			} else {
				if multi.Complete(parts) != nil {
					multi.Abort()
				}
			}
		}

		d.putbuf(buf) // needs to be here to pick up new buf value
		close(done)   // free up any waiting goroutines
	}()

	// Fills from 0 to total from current
	fromSmallCurrent := func(total int64) error {
		current, err := d.ReadStream(ctx, path, 0)
		if err != nil {
			return err
		}

		bytesRead = 0
		for int64(bytesRead) < total {
			//The loop should very rarely enter a second iteration
			nn, err := current.Read(buf[bytesRead:total])
			bytesRead += nn
			if err != nil {
				if err != io.EOF {
					return err
				}

				break
			}

		}
		return nil
	}

	// Fills from parameter to chunkSize from reader
	fromReader := func(from int64) error {
		bytesRead = 0
		for from+int64(bytesRead) < d.ChunkSize {
			nn, err := reader.Read(buf[from+int64(bytesRead):])
			totalRead += int64(nn)
			bytesRead += nn

			if err != nil {
				if err != io.EOF {
					return err
				}

				break
			}
		}

		if putErrChan == nil {
			putErrChan = make(chan error)
		} else {
			if putErr := <-putErrChan; putErr != nil {
				putErrChan = nil
				return putErr
			}
		}

		go func(bytesRead int, from int64, buf []byte) {
			defer d.putbuf(buf) // this buffer gets dropped after this call

			// DRAGONS(stevvooe): There are few things one might want to know
			// about this section. First, the putErrChan is expecting an error
			// and a nil or just a nil to come through the channel. This is
			// covered by the silly defer below. The other aspect is the OSS
			// retry backoff to deal with RequestTimeout errors. Even though
			// the underlying OSS library should handle it, it doesn't seem to
			// be part of the shouldRetry function (see denverdino/aliyungo/oss).
			defer func() {
				select {
				case putErrChan <- nil: // for some reason, we do this no matter what.
				case <-done:
					return // ensure we don't leak the goroutine
				}
			}()

			if bytesRead <= 0 {
				return
			}

			var err error
			var part oss.Part

			part, err = multi.PutPartWithTimeout(int(partNumber), bytes.NewReader(buf[0:int64(bytesRead)+from]), defaultTimeout)

			if err != nil {
				logrus.Errorf("error putting part, aborting: %v", err)
				select {
				case putErrChan <- err:
				case <-done:
					return // don't leak the goroutine
				}
			}

			// parts and partNumber are safe, because this function is the
			// only one modifying them and we force it to be executed
			// serially.
			parts = append(parts, part)
			partNumber++
		}(bytesRead, from, buf)

		buf = d.getbuf() // use a new buffer for the next call
		return nil
	}

	if offset > 0 {
		resp, err := d.Bucket.Head(d.ossPath(path), nil)
		if err != nil {
			if ossErr, ok := err.(*oss.Error); !ok || ossErr.StatusCode != http.StatusNotFound {
				return 0, err
			}
		}

		currentLength := int64(0)
		if err == nil {
			currentLength = resp.ContentLength
		}

		if currentLength >= offset {
			if offset < d.ChunkSize {
				// chunkSize > currentLength >= offset
				if err = fromSmallCurrent(offset); err != nil {
					return totalRead, err
				}

				if err = fromReader(offset); err != nil {
					return totalRead, err
				}

				if totalRead+offset < d.ChunkSize {
					return totalRead, nil
				}
			} else {
				// currentLength >= offset >= chunkSize
				_, part, err = multi.PutPartCopy(partNumber,
					oss.CopyOptions{CopySourceOptions: "bytes=0-" + strconv.FormatInt(offset-1, 10)},
					d.Bucket.Path(d.ossPath(path)))
				if err != nil {
					return 0, err
				}

				parts = append(parts, part)
				partNumber++
			}
		} else {
			// Fills between parameters with 0s but only when to - from <= chunkSize
			fromZeroFillSmall := func(from, to int64) error {
				bytesRead = 0
				for from+int64(bytesRead) < to {
					nn, err := bytes.NewReader(d.zeros).Read(buf[from+int64(bytesRead) : to])
					bytesRead += nn
					if err != nil {
						return err
					}
				}

				return nil
			}

			// Fills between parameters with 0s, making new parts
			fromZeroFillLarge := func(from, to int64) error {
				bytesRead64 := int64(0)
				for to-(from+bytesRead64) >= d.ChunkSize {
					part, err := multi.PutPartWithTimeout(int(partNumber), bytes.NewReader(d.zeros), defaultTimeout)
					if err != nil {
						return err
					}
					bytesRead64 += d.ChunkSize

					parts = append(parts, part)
					partNumber++
				}

				return fromZeroFillSmall(0, (to-from)%d.ChunkSize)
			}

			// currentLength < offset
			if currentLength < d.ChunkSize {
				if offset < d.ChunkSize {
					// chunkSize > offset > currentLength
					if err = fromSmallCurrent(currentLength); err != nil {
						return totalRead, err
					}

					if err = fromZeroFillSmall(currentLength, offset); err != nil {
						return totalRead, err
					}

					if err = fromReader(offset); err != nil {
						return totalRead, err
					}

					if totalRead+offset < d.ChunkSize {
						return totalRead, nil
					}
				} else {
					// offset >= chunkSize > currentLength
					if err = fromSmallCurrent(currentLength); err != nil {
						return totalRead, err
					}

					if err = fromZeroFillSmall(currentLength, d.ChunkSize); err != nil {
						return totalRead, err
					}

					part, err = multi.PutPartWithTimeout(int(partNumber), bytes.NewReader(buf), defaultTimeout)
					if err != nil {
						return totalRead, err
					}

					parts = append(parts, part)
					partNumber++

					//Zero fill from chunkSize up to offset, then some reader
					if err = fromZeroFillLarge(d.ChunkSize, offset); err != nil {
						return totalRead, err
					}

					if err = fromReader(offset % d.ChunkSize); err != nil {
						return totalRead, err
					}

					if totalRead+(offset%d.ChunkSize) < d.ChunkSize {
						return totalRead, nil
					}
				}
			} else {
				// offset > currentLength >= chunkSize
				_, part, err = multi.PutPartCopy(partNumber,
					oss.CopyOptions{},
					d.Bucket.Path(d.ossPath(path)))
				if err != nil {
					return 0, err
				}

				parts = append(parts, part)
				partNumber++

				//Zero fill from currentLength up to offset, then some reader
				if err = fromZeroFillLarge(currentLength, offset); err != nil {
					return totalRead, err
				}

				if err = fromReader((offset - currentLength) % d.ChunkSize); err != nil {
					return totalRead, err
				}

				if totalRead+((offset-currentLength)%d.ChunkSize) < d.ChunkSize {
					return totalRead, nil
				}
			}

		}
	}

	for {
		if err = fromReader(0); err != nil {
			return totalRead, err
		}

		if int64(bytesRead) < d.ChunkSize {
			break
		}
	}

	return totalRead, nil
}

// Stat retrieves the FileInfo for the given path, including the current size
// in bytes and the creation time.
func (d *driver) Stat(ctx context.Context, path string) (storagedriver.FileInfo, error) {
	listResponse, err := d.Bucket.List(d.ossPath(path), "", "", 1)
	if err != nil {
		return nil, err
	}

	fi := storagedriver.FileInfoFields{
		Path: path,
	}

	if len(listResponse.Contents) == 1 {
		if listResponse.Contents[0].Key != d.ossPath(path) {
			fi.IsDir = true
		} else {
			fi.IsDir = false
			fi.Size = listResponse.Contents[0].Size

			timestamp, err := time.Parse(time.RFC3339Nano, listResponse.Contents[0].LastModified)
			if err != nil {
				return nil, err
			}
			fi.ModTime = timestamp
		}
	} else if len(listResponse.CommonPrefixes) == 1 {
		fi.IsDir = true
	} else {
		return nil, storagedriver.PathNotFoundError{Path: path}
	}

	return storagedriver.FileInfoInternal{FileInfoFields: fi}, nil
}

// List returns a list of the objects that are direct descendants of the given path.
func (d *driver) List(ctx context.Context, opath string) ([]string, error) {
	path := opath
	if path != "/" && opath[len(path)-1] != '/' {
		path = path + "/"
	}

	// This is to cover for the cases when the rootDirectory of the driver is either "" or "/".
	// In those cases, there is no root prefix to replace and we must actually add a "/" to all
	// results in order to keep them as valid paths as recognized by storagedriver.PathRegexp
	prefix := ""
	if d.ossPath("") == "" {
		prefix = "/"
	}

	listResponse, err := d.Bucket.List(d.ossPath(path), "/", "", listMax)
	if err != nil {
		return nil, parseError(opath, err)
	}

	files := []string{}
	directories := []string{}

	for {
		for _, key := range listResponse.Contents {
			files = append(files, strings.Replace(key.Key, d.ossPath(""), prefix, 1))
		}

		for _, commonPrefix := range listResponse.CommonPrefixes {
			directories = append(directories, strings.Replace(commonPrefix[0:len(commonPrefix)-1], d.ossPath(""), prefix, 1))
		}

		if listResponse.IsTruncated {
			listResponse, err = d.Bucket.List(d.ossPath(path), "/", listResponse.NextMarker, listMax)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	if opath != "/" {
		if len(files) == 0 && len(directories) == 0 {
			// Treat empty response as missing directory, since we don't actually
			// have directories in s3.
			return nil, storagedriver.PathNotFoundError{Path: opath}
		}
	}

	return append(files, directories...), nil
}

// Move moves an object stored at sourcePath to destPath, removing the original
// object.
func (d *driver) Move(ctx context.Context, sourcePath string, destPath string) error {
	logrus.Infof("Move from %s to %s", d.ossPath(sourcePath), d.ossPath(destPath))

	err := d.Bucket.CopyLargeFile(d.ossPath(sourcePath), d.ossPath(destPath),
		d.getContentType(),
		getPermissions(),
		oss.Options{})
	if err != nil {
		logrus.Errorf("Failed for move from %s to %s: %v", d.ossPath(sourcePath), d.ossPath(destPath), err)
		return parseError(sourcePath, err)
	}

	return d.Delete(ctx, sourcePath)
}

// Delete recursively deletes all objects stored at "path" and its subpaths.
func (d *driver) Delete(ctx context.Context, path string) error {
	listResponse, err := d.Bucket.List(d.ossPath(path), "", "", listMax)
	if err != nil || len(listResponse.Contents) == 0 {
		return storagedriver.PathNotFoundError{Path: path}
	}

	ossObjects := make([]oss.Object, listMax)

	for len(listResponse.Contents) > 0 {
		for index, key := range listResponse.Contents {
			ossObjects[index].Key = key.Key
		}

		err := d.Bucket.DelMulti(oss.Delete{Quiet: false, Objects: ossObjects[0:len(listResponse.Contents)]})
		if err != nil {
			return nil
		}

		listResponse, err = d.Bucket.List(d.ossPath(path), "", "", listMax)
		if err != nil {
			return err
		}
	}

	return nil
}

// URLFor returns a URL which may be used to retrieve the content stored at the given path.
// May return an UnsupportedMethodErr in certain StorageDriver implementations.
func (d *driver) URLFor(ctx context.Context, path string, options map[string]interface{}) (string, error) {
	methodString := "GET"
	method, ok := options["method"]
	if ok {
		methodString, ok = method.(string)
		if !ok || (methodString != "GET") {
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
	logrus.Infof("methodString: %s, expiresTime: %v", methodString, expiresTime)
	signedURL := d.Bucket.SignedURLWithMethod(methodString, d.ossPath(path), expiresTime, nil, nil)
	logrus.Infof("signed URL: %s", signedURL)
	return signedURL, nil
}

func (d *driver) ossPath(path string) string {
	return strings.TrimLeft(strings.TrimRight(d.RootDirectory, "/")+path, "/")
}

func parseError(path string, err error) error {
	if ossErr, ok := err.(*oss.Error); ok && ossErr.StatusCode == http.StatusNotFound && (ossErr.Code == "NoSuchKey" || ossErr.Code == "") {
		return storagedriver.PathNotFoundError{Path: path}
	}

	return err
}

func hasCode(err error, code string) bool {
	ossErr, ok := err.(*oss.Error)
	return ok && ossErr.Code == code
}

func (d *driver) getOptions() oss.Options {
	return oss.Options{ServerSideEncryption: d.Encrypt}
}

func getPermissions() oss.ACL {
	return oss.Private
}

func (d *driver) getContentType() string {
	return "application/octet-stream"
}

// getbuf returns a buffer from the driver's pool with length d.ChunkSize.
func (d *driver) getbuf() []byte {
	return d.pool.Get().([]byte)
}

func (d *driver) putbuf(p []byte) {
	copy(p, d.zeros)
	d.pool.Put(p)
}
