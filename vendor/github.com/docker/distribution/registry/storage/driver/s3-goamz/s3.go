// Package s3 provides a storagedriver.StorageDriver implementation to
// store blobs in Amazon S3 cloud storage.
//
// This package leverages the docker/goamz client library for interfacing with
// S3. It is intended to be deprecated in favor of the s3-aws driver
// implementation.
//
// Because S3 is a key, value store the Stat call does not support last modification
// time for directories (directories are an abstraction for key, value stores)
//
// Keep in mind that S3 guarantees only read-after-write consistency for new
// objects, but no read-after-update or list-after-write consistency.
package s3

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

	"github.com/Sirupsen/logrus"
	"github.com/docker/goamz/aws"
	"github.com/docker/goamz/s3"

	"github.com/docker/distribution/context"
	"github.com/docker/distribution/registry/client/transport"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/base"
	"github.com/docker/distribution/registry/storage/driver/factory"
)

const driverName = "s3goamz"

// minChunkSize defines the minimum multipart upload chunk size
// S3 API requires multipart upload chunks to be at least 5MB
const minChunkSize = 5 << 20

const defaultChunkSize = 2 * minChunkSize

// listMax is the largest amount of objects you can request from S3 in a list call
const listMax = 1000

//DriverParameters A struct that encapsulates all of the driver parameters after all values have been set
type DriverParameters struct {
	AccessKey     string
	SecretKey     string
	Bucket        string
	Region        aws.Region
	Encrypt       bool
	Secure        bool
	V4Auth        bool
	ChunkSize     int64
	RootDirectory string
	StorageClass  s3.StorageClass
	UserAgent     string
}

func init() {
	factory.Register(driverName, &s3DriverFactory{})
}

// s3DriverFactory implements the factory.StorageDriverFactory interface
type s3DriverFactory struct{}

func (factory *s3DriverFactory) Create(parameters map[string]interface{}) (storagedriver.StorageDriver, error) {
	return FromParameters(parameters)
}

type driver struct {
	S3            *s3.S3
	Bucket        *s3.Bucket
	ChunkSize     int64
	Encrypt       bool
	RootDirectory string
	StorageClass  s3.StorageClass

	pool  sync.Pool // pool []byte buffers used for WriteStream
	zeros []byte    // shared, zero-valued buffer used for WriteStream
}

type baseEmbed struct {
	base.Base
}

// Driver is a storagedriver.StorageDriver implementation backed by Amazon S3
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
	accessKey := parameters["accesskey"]
	if accessKey == nil {
		accessKey = ""
	}

	secretKey := parameters["secretkey"]
	if secretKey == nil {
		secretKey = ""
	}

	regionName := parameters["region"]
	if regionName == nil || fmt.Sprint(regionName) == "" {
		return nil, fmt.Errorf("No region parameter provided")
	}
	region := aws.GetRegion(fmt.Sprint(regionName))
	if region.Name == "" {
		return nil, fmt.Errorf("Invalid region provided: %v", region)
	}

	bucket := parameters["bucket"]
	if bucket == nil || fmt.Sprint(bucket) == "" {
		return nil, fmt.Errorf("No bucket parameter provided")
	}

	encryptBool := false
	encrypt := parameters["encrypt"]
	switch encrypt := encrypt.(type) {
	case string:
		b, err := strconv.ParseBool(encrypt)
		if err != nil {
			return nil, fmt.Errorf("The encrypt parameter should be a boolean")
		}
		encryptBool = b
	case bool:
		encryptBool = encrypt
	case nil:
		// do nothing
	default:
		return nil, fmt.Errorf("The encrypt parameter should be a boolean")
	}

	secureBool := true
	secure := parameters["secure"]
	switch secure := secure.(type) {
	case string:
		b, err := strconv.ParseBool(secure)
		if err != nil {
			return nil, fmt.Errorf("The secure parameter should be a boolean")
		}
		secureBool = b
	case bool:
		secureBool = secure
	case nil:
		// do nothing
	default:
		return nil, fmt.Errorf("The secure parameter should be a boolean")
	}

	v4AuthBool := false
	v4Auth := parameters["v4auth"]
	switch v4Auth := v4Auth.(type) {
	case string:
		b, err := strconv.ParseBool(v4Auth)
		if err != nil {
			return nil, fmt.Errorf("The v4auth parameter should be a boolean")
		}
		v4AuthBool = b
	case bool:
		v4AuthBool = v4Auth
	case nil:
		// do nothing
	default:
		return nil, fmt.Errorf("The v4auth parameter should be a boolean")
	}

	chunkSize := int64(defaultChunkSize)
	chunkSizeParam := parameters["chunksize"]
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
	case nil:
		// do nothing
	default:
		return nil, fmt.Errorf("invalid value for chunksize: %#v", chunkSizeParam)
	}

	if chunkSize < minChunkSize {
		return nil, fmt.Errorf("The chunksize %#v parameter should be a number that is larger than or equal to %d", chunkSize, minChunkSize)
	}

	rootDirectory := parameters["rootdirectory"]
	if rootDirectory == nil {
		rootDirectory = ""
	}

	storageClass := s3.StandardStorage
	storageClassParam := parameters["storageclass"]
	if storageClassParam != nil {
		storageClassString, ok := storageClassParam.(string)
		if !ok {
			return nil, fmt.Errorf("The storageclass parameter must be one of %v, %v invalid", []s3.StorageClass{s3.StandardStorage, s3.ReducedRedundancy}, storageClassParam)
		}
		// All valid storage class parameters are UPPERCASE, so be a bit more flexible here
		storageClassCasted := s3.StorageClass(strings.ToUpper(storageClassString))
		if storageClassCasted != s3.StandardStorage && storageClassCasted != s3.ReducedRedundancy {
			return nil, fmt.Errorf("The storageclass parameter must be one of %v, %v invalid", []s3.StorageClass{s3.StandardStorage, s3.ReducedRedundancy}, storageClassParam)
		}
		storageClass = storageClassCasted
	}

	userAgent := parameters["useragent"]
	if userAgent == nil {
		userAgent = ""
	}

	params := DriverParameters{
		fmt.Sprint(accessKey),
		fmt.Sprint(secretKey),
		fmt.Sprint(bucket),
		region,
		encryptBool,
		secureBool,
		v4AuthBool,
		chunkSize,
		fmt.Sprint(rootDirectory),
		storageClass,
		fmt.Sprint(userAgent),
	}

	return New(params)
}

// New constructs a new Driver with the given AWS credentials, region, encryption flag, and
// bucketName
func New(params DriverParameters) (*Driver, error) {
	auth, err := aws.GetAuth(params.AccessKey, params.SecretKey, "", time.Time{})
	if err != nil {
		return nil, fmt.Errorf("unable to resolve aws credentials, please ensure that 'accesskey' and 'secretkey' are properly set or the credentials are available in $HOME/.aws/credentials: %v", err)
	}

	if !params.Secure {
		params.Region.S3Endpoint = strings.Replace(params.Region.S3Endpoint, "https", "http", 1)
	}

	s3obj := s3.New(auth, params.Region)

	if params.UserAgent != "" {
		s3obj.Client = &http.Client{
			Transport: transport.NewTransport(http.DefaultTransport,
				transport.NewHeaderRequestModifier(http.Header{
					http.CanonicalHeaderKey("User-Agent"): []string{params.UserAgent},
				}),
			),
		}
	}

	if params.V4Auth {
		s3obj.Signature = aws.V4Signature
	} else {
		if params.Region.Name == "eu-central-1" {
			return nil, fmt.Errorf("The eu-central-1 region only works with v4 authentication")
		}
	}

	bucket := s3obj.Bucket(params.Bucket)

	// TODO Currently multipart uploads have no timestamps, so this would be unwise
	// if you initiated a new s3driver while another one is running on the same bucket.
	// multis, _, err := bucket.ListMulti("", "")
	// if err != nil {
	// 	return nil, err
	// }

	// for _, multi := range multis {
	// 	err := multi.Abort()
	// 	//TODO appropriate to do this error checking?
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	d := &driver{
		S3:            s3obj,
		Bucket:        bucket,
		ChunkSize:     params.ChunkSize,
		Encrypt:       params.Encrypt,
		RootDirectory: params.RootDirectory,
		StorageClass:  params.StorageClass,
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
	content, err := d.Bucket.Get(d.s3Path(path))
	if err != nil {
		return nil, parseError(path, err)
	}
	return content, nil
}

// PutContent stores the []byte content at a location designated by "path".
func (d *driver) PutContent(ctx context.Context, path string, contents []byte) error {
	return parseError(path, d.Bucket.Put(d.s3Path(path), contents, d.getContentType(), getPermissions(), d.getOptions()))
}

// ReadStream retrieves an io.ReadCloser for the content stored at "path" with a
// given byte offset.
func (d *driver) ReadStream(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	headers := make(http.Header)
	headers.Add("Range", "bytes="+strconv.FormatInt(offset, 10)+"-")

	resp, err := d.Bucket.GetResponseWithHeaders(d.s3Path(path), headers)
	if err != nil {
		if s3Err, ok := err.(*s3.Error); ok && s3Err.Code == "InvalidRange" {
			return ioutil.NopCloser(bytes.NewReader(nil)), nil
		}

		return nil, parseError(path, err)
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
	parts := []s3.Part{}
	var part s3.Part
	done := make(chan struct{}) // stopgap to free up waiting goroutines

	multi, err := d.Bucket.InitMulti(d.s3Path(path), d.getContentType(), getPermissions(), d.getOptions())
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
			// covered by the silly defer below. The other aspect is the s3
			// retry backoff to deal with RequestTimeout errors. Even though
			// the underlying s3 library should handle it, it doesn't seem to
			// be part of the shouldRetry function (see AdRoll/goamz/s3).
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
			var part s3.Part

		loop:
			for retries := 0; retries < 5; retries++ {
				part, err = multi.PutPart(int(partNumber), bytes.NewReader(buf[0:int64(bytesRead)+from]))
				if err == nil {
					break // success!
				}

				// NOTE(stevvooe): This retry code tries to only retry under
				// conditions where the s3 package does not. We may add s3
				// error codes to the below if we see others bubble up in the
				// application. Right now, the most troubling is
				// RequestTimeout, which seems to only triggered when a tcp
				// connection to s3 slows to a crawl. If the RequestTimeout
				// ends up getting added to the s3 library and we don't see
				// other errors, this retry loop can be removed.
				switch err := err.(type) {
				case *s3.Error:
					switch err.Code {
					case "RequestTimeout":
						// allow retries on only this error.
					default:
						break loop
					}
				}

				backoff := 100 * time.Millisecond * time.Duration(retries+1)
				logrus.Errorf("error putting part, retrying after %v: %v", err, backoff.String())
				time.Sleep(backoff)
			}

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
		resp, err := d.Bucket.Head(d.s3Path(path), nil)
		if err != nil {
			if s3Err, ok := err.(*s3.Error); !ok || s3Err.Code != "NoSuchKey" {
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
					s3.CopyOptions{CopySourceOptions: "bytes=0-" + strconv.FormatInt(offset-1, 10)},
					d.Bucket.Name+"/"+d.s3Path(path))
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
					part, err := multi.PutPart(int(partNumber), bytes.NewReader(d.zeros))
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

					part, err = multi.PutPart(int(partNumber), bytes.NewReader(buf))
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
					s3.CopyOptions{},
					d.Bucket.Name+"/"+d.s3Path(path))
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
	listResponse, err := d.Bucket.List(d.s3Path(path), "", "", 1)
	if err != nil {
		return nil, err
	}

	fi := storagedriver.FileInfoFields{
		Path: path,
	}

	if len(listResponse.Contents) == 1 {
		if listResponse.Contents[0].Key != d.s3Path(path) {
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
	if path != "/" && path[len(path)-1] != '/' {
		path = path + "/"
	}

	// This is to cover for the cases when the rootDirectory of the driver is either "" or "/".
	// In those cases, there is no root prefix to replace and we must actually add a "/" to all
	// results in order to keep them as valid paths as recognized by storagedriver.PathRegexp
	prefix := ""
	if d.s3Path("") == "" {
		prefix = "/"
	}

	listResponse, err := d.Bucket.List(d.s3Path(path), "/", "", listMax)
	if err != nil {
		return nil, parseError(opath, err)
	}

	files := []string{}
	directories := []string{}

	for {
		for _, key := range listResponse.Contents {
			files = append(files, strings.Replace(key.Key, d.s3Path(""), prefix, 1))
		}

		for _, commonPrefix := range listResponse.CommonPrefixes {
			directories = append(directories, strings.Replace(commonPrefix[0:len(commonPrefix)-1], d.s3Path(""), prefix, 1))
		}

		if listResponse.IsTruncated {
			listResponse, err = d.Bucket.List(d.s3Path(path), "/", listResponse.NextMarker, listMax)
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
	/* This is terrible, but aws doesn't have an actual move. */
	_, err := d.Bucket.PutCopy(d.s3Path(destPath), getPermissions(),
		s3.CopyOptions{Options: d.getOptions(), ContentType: d.getContentType()}, d.Bucket.Name+"/"+d.s3Path(sourcePath))
	if err != nil {
		return parseError(sourcePath, err)
	}

	return d.Delete(ctx, sourcePath)
}

// Delete recursively deletes all objects stored at "path" and its subpaths.
func (d *driver) Delete(ctx context.Context, path string) error {
	listResponse, err := d.Bucket.List(d.s3Path(path), "", "", listMax)
	if err != nil || len(listResponse.Contents) == 0 {
		return storagedriver.PathNotFoundError{Path: path}
	}

	s3Objects := make([]s3.Object, listMax)

	for len(listResponse.Contents) > 0 {
		for index, key := range listResponse.Contents {
			s3Objects[index].Key = key.Key
		}

		err := d.Bucket.DelMulti(s3.Delete{Quiet: false, Objects: s3Objects[0:len(listResponse.Contents)]})
		if err != nil {
			return nil
		}

		listResponse, err = d.Bucket.List(d.s3Path(path), "", "", listMax)
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

	return d.Bucket.SignedURLWithMethod(methodString, d.s3Path(path), expiresTime, nil, nil), nil
}

func (d *driver) s3Path(path string) string {
	return strings.TrimLeft(strings.TrimRight(d.RootDirectory, "/")+path, "/")
}

// S3BucketKey returns the s3 bucket key for the given storage driver path.
func (d *Driver) S3BucketKey(path string) string {
	return d.StorageDriver.(*driver).s3Path(path)
}

func parseError(path string, err error) error {
	if s3Err, ok := err.(*s3.Error); ok && s3Err.Code == "NoSuchKey" {
		return storagedriver.PathNotFoundError{Path: path}
	}

	return err
}

func hasCode(err error, code string) bool {
	s3err, ok := err.(*aws.Error)
	return ok && s3err.Code == code
}

func (d *driver) getOptions() s3.Options {
	return s3.Options{
		SSE:          d.Encrypt,
		StorageClass: d.StorageClass,
	}
}

func getPermissions() s3.ACL {
	return s3.Private
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
