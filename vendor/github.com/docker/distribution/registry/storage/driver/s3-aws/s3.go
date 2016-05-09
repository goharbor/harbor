// Package s3 provides a storagedriver.StorageDriver implementation to
// store blobs in Amazon S3 cloud storage.
//
// This package leverages the official aws client library for interfacing with
// S3.
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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/docker/distribution/context"
	"github.com/docker/distribution/registry/client/transport"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/base"
	"github.com/docker/distribution/registry/storage/driver/factory"
)

const driverName = "s3aws"

// minChunkSize defines the minimum multipart upload chunk size
// S3 API requires multipart upload chunks to be at least 5MB
const minChunkSize = 5 << 20

const defaultChunkSize = 2 * minChunkSize

// listMax is the largest amount of objects you can request from S3 in a list call
const listMax = 1000

// validRegions maps known s3 region identifiers to region descriptors
var validRegions = map[string]struct{}{}

//DriverParameters A struct that encapsulates all of the driver parameters after all values have been set
type DriverParameters struct {
	AccessKey     string
	SecretKey     string
	Bucket        string
	Region        string
	Encrypt       bool
	Secure        bool
	ChunkSize     int64
	RootDirectory string
	StorageClass  string
	UserAgent     string
}

func init() {
	for _, region := range []string{
		"us-east-1",
		"us-west-1",
		"us-west-2",
		"eu-west-1",
		"eu-central-1",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
		"ap-northeast-2",
		"sa-east-1",
	} {
		validRegions[region] = struct{}{}
	}

	// Register this as the default s3 driver in addition to s3aws
	factory.Register("s3", &s3DriverFactory{})
	factory.Register(driverName, &s3DriverFactory{})
}

// s3DriverFactory implements the factory.StorageDriverFactory interface
type s3DriverFactory struct{}

func (factory *s3DriverFactory) Create(parameters map[string]interface{}) (storagedriver.StorageDriver, error) {
	return FromParameters(parameters)
}

type driver struct {
	S3            *s3.S3
	Bucket        string
	ChunkSize     int64
	Encrypt       bool
	RootDirectory string
	StorageClass  string

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
	accessKey, ok := parameters["accesskey"]
	if !ok {
		accessKey = ""
	}
	secretKey, ok := parameters["secretkey"]
	if !ok {
		secretKey = ""
	}

	regionName, ok := parameters["region"]
	if !ok || fmt.Sprint(regionName) == "" {
		return nil, fmt.Errorf("No region parameter provided")
	}
	region := fmt.Sprint(regionName)
	_, ok = validRegions[region]
	if !ok {
		return nil, fmt.Errorf("Invalid region provided: %v", region)
	}

	bucket, ok := parameters["bucket"]
	if !ok || fmt.Sprint(bucket) == "" {
		return nil, fmt.Errorf("No bucket parameter provided")
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

	storageClass := s3.StorageClassStandard
	storageClassParam, ok := parameters["storageclass"]
	if ok {
		storageClassString, ok := storageClassParam.(string)
		if !ok {
			return nil, fmt.Errorf("The storageclass parameter must be one of %v, %v invalid", []string{s3.StorageClassStandard, s3.StorageClassReducedRedundancy}, storageClassParam)
		}
		// All valid storage class parameters are UPPERCASE, so be a bit more flexible here
		storageClassString = strings.ToUpper(storageClassString)
		if storageClassString != s3.StorageClassStandard && storageClassString != s3.StorageClassReducedRedundancy {
			return nil, fmt.Errorf("The storageclass parameter must be one of %v, %v invalid", []string{s3.StorageClassStandard, s3.StorageClassReducedRedundancy}, storageClassParam)
		}
		storageClass = storageClassString
	}

	userAgent, ok := parameters["useragent"]
	if !ok {
		userAgent = ""
	}

	params := DriverParameters{
		fmt.Sprint(accessKey),
		fmt.Sprint(secretKey),
		fmt.Sprint(bucket),
		region,
		encryptBool,
		secureBool,
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
	awsConfig := aws.NewConfig()
	creds := credentials.NewChainCredentials([]credentials.Provider{
		&credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     params.AccessKey,
				SecretAccessKey: params.SecretKey,
			},
		},
		&credentials.EnvProvider{},
		&credentials.SharedCredentialsProvider{},
		&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(session.New())},
	})

	awsConfig.WithCredentials(creds)
	awsConfig.WithRegion(params.Region)
	awsConfig.WithDisableSSL(!params.Secure)
	// awsConfig.WithMaxRetries(10)

	if params.UserAgent != "" {
		awsConfig.WithHTTPClient(&http.Client{
			Transport: transport.NewTransport(http.DefaultTransport, transport.NewHeaderRequestModifier(http.Header{http.CanonicalHeaderKey("User-Agent"): []string{params.UserAgent}})),
		})
	}

	s3obj := s3.New(session.New(awsConfig))

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
		Bucket:        params.Bucket,
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
	reader, err := d.ReadStream(ctx, path, 0)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

// PutContent stores the []byte content at a location designated by "path".
func (d *driver) PutContent(ctx context.Context, path string, contents []byte) error {
	_, err := d.S3.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(d.Bucket),
		Key:                  aws.String(d.s3Path(path)),
		ContentType:          d.getContentType(),
		ACL:                  d.getACL(),
		ServerSideEncryption: d.getEncryptionMode(),
		StorageClass:         d.getStorageClass(),
		Body:                 bytes.NewReader(contents),
	})
	return parseError(path, err)
}

// ReadStream retrieves an io.ReadCloser for the content stored at "path" with a
// given byte offset.
func (d *driver) ReadStream(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	resp, err := d.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(d.Bucket),
		Key:    aws.String(d.s3Path(path)),
		Range:  aws.String("bytes=" + strconv.FormatInt(offset, 10) + "-"),
	})

	if err != nil {
		if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "InvalidRange" {
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
	var partNumber int64 = 1
	bytesRead := 0
	var putErrChan chan error
	parts := []*s3.CompletedPart{}
	done := make(chan struct{}) // stopgap to free up waiting goroutines

	resp, err := d.S3.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket:               aws.String(d.Bucket),
		Key:                  aws.String(d.s3Path(path)),
		ContentType:          d.getContentType(),
		ACL:                  d.getACL(),
		ServerSideEncryption: d.getEncryptionMode(),
		StorageClass:         d.getStorageClass(),
	})
	if err != nil {
		return 0, err
	}

	uploadID := resp.UploadId

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
			_, err := d.S3.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
				Bucket:   aws.String(d.Bucket),
				Key:      aws.String(d.s3Path(path)),
				UploadId: uploadID,
				MultipartUpload: &s3.CompletedMultipartUpload{
					Parts: parts,
				},
			})
			if err != nil {
				// TODO (brianbland): log errors here
				d.S3.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
					Bucket:   aws.String(d.Bucket),
					Key:      aws.String(d.s3Path(path)),
					UploadId: uploadID,
				})
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

			resp, err := d.S3.UploadPart(&s3.UploadPartInput{
				Bucket:     aws.String(d.Bucket),
				Key:        aws.String(d.s3Path(path)),
				PartNumber: aws.Int64(partNumber),
				UploadId:   uploadID,
				Body:       bytes.NewReader(buf[0 : int64(bytesRead)+from]),
			})
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
			parts = append(parts, &s3.CompletedPart{
				ETag:       resp.ETag,
				PartNumber: aws.Int64(partNumber),
			})
			partNumber++
		}(bytesRead, from, buf)

		buf = d.getbuf() // use a new buffer for the next call
		return nil
	}

	if offset > 0 {
		resp, err := d.S3.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(d.Bucket),
			Key:    aws.String(d.s3Path(path)),
		})
		if err != nil {
			if s3Err, ok := err.(awserr.Error); !ok || s3Err.Code() != "NoSuchKey" {
				return 0, err
			}
		}

		currentLength := int64(0)
		if err == nil && resp.ContentLength != nil {
			currentLength = *resp.ContentLength
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
				resp, err := d.S3.UploadPartCopy(&s3.UploadPartCopyInput{
					Bucket:          aws.String(d.Bucket),
					Key:             aws.String(d.s3Path(path)),
					PartNumber:      aws.Int64(partNumber),
					UploadId:        uploadID,
					CopySource:      aws.String(d.Bucket + "/" + d.s3Path(path)),
					CopySourceRange: aws.String("bytes=0-" + strconv.FormatInt(offset-1, 10)),
				})
				if err != nil {
					return 0, err
				}

				parts = append(parts, &s3.CompletedPart{
					ETag:       resp.CopyPartResult.ETag,
					PartNumber: aws.Int64(partNumber),
				})
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
					resp, err := d.S3.UploadPart(&s3.UploadPartInput{
						Bucket:     aws.String(d.Bucket),
						Key:        aws.String(d.s3Path(path)),
						PartNumber: aws.Int64(partNumber),
						UploadId:   uploadID,
						Body:       bytes.NewReader(d.zeros),
					})
					if err != nil {
						return err
					}
					bytesRead64 += d.ChunkSize

					parts = append(parts, &s3.CompletedPart{
						ETag:       resp.ETag,
						PartNumber: aws.Int64(partNumber),
					})
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

					resp, err := d.S3.UploadPart(&s3.UploadPartInput{
						Bucket:     aws.String(d.Bucket),
						Key:        aws.String(d.s3Path(path)),
						PartNumber: aws.Int64(partNumber),
						UploadId:   uploadID,
						Body:       bytes.NewReader(buf),
					})
					if err != nil {
						return totalRead, err
					}

					parts = append(parts, &s3.CompletedPart{
						ETag:       resp.ETag,
						PartNumber: aws.Int64(partNumber),
					})
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
				resp, err := d.S3.UploadPartCopy(&s3.UploadPartCopyInput{
					Bucket:     aws.String(d.Bucket),
					Key:        aws.String(d.s3Path(path)),
					PartNumber: aws.Int64(partNumber),
					UploadId:   uploadID,
					CopySource: aws.String(d.Bucket + "/" + d.s3Path(path)),
				})
				if err != nil {
					return 0, err
				}

				parts = append(parts, &s3.CompletedPart{
					ETag:       resp.CopyPartResult.ETag,
					PartNumber: aws.Int64(partNumber),
				})
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
	resp, err := d.S3.ListObjects(&s3.ListObjectsInput{
		Bucket:  aws.String(d.Bucket),
		Prefix:  aws.String(d.s3Path(path)),
		MaxKeys: aws.Int64(1),
	})
	if err != nil {
		return nil, err
	}

	fi := storagedriver.FileInfoFields{
		Path: path,
	}

	if len(resp.Contents) == 1 {
		if *resp.Contents[0].Key != d.s3Path(path) {
			fi.IsDir = true
		} else {
			fi.IsDir = false
			fi.Size = *resp.Contents[0].Size
			fi.ModTime = *resp.Contents[0].LastModified
		}
	} else if len(resp.CommonPrefixes) == 1 {
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

	resp, err := d.S3.ListObjects(&s3.ListObjectsInput{
		Bucket:    aws.String(d.Bucket),
		Prefix:    aws.String(d.s3Path(path)),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Int64(listMax),
	})
	if err != nil {
		return nil, parseError(opath, err)
	}

	files := []string{}
	directories := []string{}

	for {
		for _, key := range resp.Contents {
			files = append(files, strings.Replace(*key.Key, d.s3Path(""), prefix, 1))
		}

		for _, commonPrefix := range resp.CommonPrefixes {
			commonPrefix := *commonPrefix.Prefix
			directories = append(directories, strings.Replace(commonPrefix[0:len(commonPrefix)-1], d.s3Path(""), prefix, 1))
		}

		if *resp.IsTruncated {
			resp, err = d.S3.ListObjects(&s3.ListObjectsInput{
				Bucket:    aws.String(d.Bucket),
				Prefix:    aws.String(d.s3Path(path)),
				Delimiter: aws.String("/"),
				MaxKeys:   aws.Int64(listMax),
				Marker:    resp.NextMarker,
			})
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
	_, err := d.S3.CopyObject(&s3.CopyObjectInput{
		Bucket:               aws.String(d.Bucket),
		Key:                  aws.String(d.s3Path(destPath)),
		ContentType:          d.getContentType(),
		ACL:                  d.getACL(),
		ServerSideEncryption: d.getEncryptionMode(),
		StorageClass:         d.getStorageClass(),
		CopySource:           aws.String(d.Bucket + "/" + d.s3Path(sourcePath)),
	})
	if err != nil {
		return parseError(sourcePath, err)
	}

	return d.Delete(ctx, sourcePath)
}

// Delete recursively deletes all objects stored at "path" and its subpaths.
func (d *driver) Delete(ctx context.Context, path string) error {
	resp, err := d.S3.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(d.Bucket),
		Prefix: aws.String(d.s3Path(path)),
	})
	if err != nil || len(resp.Contents) == 0 {
		return storagedriver.PathNotFoundError{Path: path}
	}

	s3Objects := make([]*s3.ObjectIdentifier, 0, listMax)

	for len(resp.Contents) > 0 {
		for _, key := range resp.Contents {
			s3Objects = append(s3Objects, &s3.ObjectIdentifier{
				Key: key.Key,
			})
		}

		_, err := d.S3.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: aws.String(d.Bucket),
			Delete: &s3.Delete{
				Objects: s3Objects,
				Quiet:   aws.Bool(false),
			},
		})
		if err != nil {
			return nil
		}

		resp, err = d.S3.ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(d.Bucket),
			Prefix: aws.String(d.s3Path(path)),
		})
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

	expiresIn := 20 * time.Minute
	expires, ok := options["expiry"]
	if ok {
		et, ok := expires.(time.Time)
		if ok {
			expiresIn = et.Sub(time.Now())
		}
	}

	var req *request.Request

	switch methodString {
	case "GET":
		req, _ = d.S3.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(d.Bucket),
			Key:    aws.String(d.s3Path(path)),
		})
	case "HEAD":
		req, _ = d.S3.HeadObjectRequest(&s3.HeadObjectInput{
			Bucket: aws.String(d.Bucket),
			Key:    aws.String(d.s3Path(path)),
		})
	default:
		panic("unreachable")
	}

	return req.Presign(expiresIn)
}

func (d *driver) s3Path(path string) string {
	return strings.TrimLeft(strings.TrimRight(d.RootDirectory, "/")+path, "/")
}

// S3BucketKey returns the s3 bucket key for the given storage driver path.
func (d *Driver) S3BucketKey(path string) string {
	return d.StorageDriver.(*driver).s3Path(path)
}

func parseError(path string, err error) error {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "NoSuchKey" {
		return storagedriver.PathNotFoundError{Path: path}
	}

	return err
}

func (d *driver) getEncryptionMode() *string {
	if d.Encrypt {
		return aws.String("AES256")
	}
	return nil
}

func (d *driver) getContentType() *string {
	return aws.String("application/octet-stream")
}

func (d *driver) getACL() *string {
	return aws.String("private")
}

func (d *driver) getStorageClass() *string {
	return aws.String(d.StorageClass)
}

// getbuf returns a buffer from the driver's pool with length d.ChunkSize.
func (d *driver) getbuf() []byte {
	return d.pool.Get().([]byte)
}

func (d *driver) putbuf(p []byte) {
	copy(p, d.zeros)
	d.pool.Put(p)
}
