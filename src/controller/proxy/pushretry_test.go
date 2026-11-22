//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package proxy

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/errors"
)

// trackingReader records how much of a blob an attempt actually consumed and
// whether it was closed.
type trackingReader struct {
	*bytes.Reader
	read   int
	closed bool
}

func newTrackingReader(content []byte) *trackingReader {
	return &trackingReader{Reader: bytes.NewReader(content)}
}

func (t *trackingReader) Read(p []byte) (int, error) {
	n, err := t.Reader.Read(p)
	t.read += n
	return n, err
}

func (t *trackingReader) Close() error {
	t.closed = true
	return nil
}

// fakeRemote hands out a fresh reader over the same content on every call, so a
// test can tell a re-opened reader from a replayed, already drained one.
type fakeRemote struct {
	content []byte
	err     error
	readers []*trackingReader
}

func (f *fakeRemote) BlobReader(_, _ string) (int64, io.ReadCloser, error) {
	if f.err != nil {
		return 0, nil, f.err
	}
	r := newTrackingReader(f.content)
	f.readers = append(f.readers, r)
	return int64(len(f.content)), r, nil
}

func (f *fakeRemote) Manifest(_, _ string) (distribution.Manifest, string, error) {
	panic("not used")
}

func (f *fakeRemote) ManifestExist(_, _ string) (bool, *distribution.Descriptor, error) {
	panic("not used")
}

func (f *fakeRemote) ListTags(_ string) ([]string, error) {
	panic("not used")
}

func (f *fakeRemote) ListReferrers(_, _, _ string) (*ocispec.Index, map[string][]string, error) {
	panic("not used")
}

type pushRetryTestSuite struct {
	suite.Suite
}

func (p *pushRetryTestSuite) TestSucceedsOnFirstAttempt() {
	calls := 0
	err := retryLocalPush(context.Background(), "blob", func() error {
		calls++
		return nil
	})
	p.Require().NoError(err)
	p.Assert().Equal(1, calls, "a successful push must not be repeated")
}

func (p *pushRetryTestSuite) TestRetriesTransientFailure() {
	// this is the reported failure: the first attempts hit a replica that does not
	// know the upload, a later one lands where it can succeed
	calls := 0
	err := retryLocalPush(context.Background(), "blob", func() error {
		calls++
		if calls < localPushAttempts {
			return errors.New("blob upload unknown to registry")
		}
		return nil
	})
	p.Require().NoError(err)
	p.Assert().Equal(localPushAttempts, calls)
}

func (p *pushRetryTestSuite) TestGivesUpAfterMaxAttempts() {
	pushErr := errors.New("manifest unknown")
	calls := 0
	err := retryLocalPush(context.Background(), "manifest", func() error {
		calls++
		return pushErr
	})
	p.Require().Error(err)
	p.Assert().Equal(localPushAttempts, calls, "the retry must be bounded")
	p.Assert().Equal(pushErr, err, "the underlying error must be surfaced, not a retry wrapper")
}

func (p *pushRetryTestSuite) TestAbortsOnCanceledContext() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	err := retryLocalPush(ctx, "blob", func() error {
		calls++
		return errors.New("some error")
	})
	p.Require().Error(err)
	p.Assert().Equal(context.Canceled, err)
	p.Assert().Equal(0, calls, "no push should be attempted once the context is done")
}

func (p *pushRetryTestSuite) TestPutBlobToLocalReopensReaderPerAttempt() {
	content := []byte("a blob that a failed attempt partially consumes")
	remote := &fakeRemote{content: content}
	local := &localInterfaceMock{}

	dig := digest.FromBytes(content)
	desc := distribution.Descriptor{Digest: dig, Size: int64(len(content))}

	// the first attempts read a few bytes and then fail, as a push whose PUT lands
	// on a replica that does not know the upload does; the last one succeeds and
	// consumes the whole blob
	partial := func(args mock.Arguments) {
		_, _ = io.CopyN(io.Discard, args.Get(2).(io.ReadCloser), 5)
	}
	whole := func(args mock.Arguments) {
		_, _ = io.ReadAll(args.Get(2).(io.ReadCloser))
	}
	for range localPushAttempts - 1 {
		local.On("PushBlob", "library/hello-world", desc, mock.Anything).
			Run(partial).Return(errors.New("blob upload unknown to registry")).Once()
	}
	local.On("PushBlob", "library/hello-world", desc, mock.Anything).
		Run(whole).Return(nil).Once()

	err := putBlobToLocal(context.Background(), local, "library/hello-world", "library/hello-world", desc, remote)
	p.Require().NoError(err)

	p.Require().Len(remote.readers, localPushAttempts, "the upstream reader must be re-opened for every attempt")
	// the decisive assertion: the final attempt saw the whole blob, so it was a
	// fresh reader rather than the drained one left behind by the failed attempts
	p.Assert().Equal(len(content), remote.readers[localPushAttempts-1].read)
	for i, r := range remote.readers {
		p.Assert().True(r.closed, "reader %d must be closed", i)
	}
	local.AssertExpectations(p.T())
}

func (p *pushRetryTestSuite) TestPutBlobToLocalRetriesReaderOpenFailure() {
	remote := &fakeRemote{err: errors.New("upstream unavailable")}
	local := &localInterfaceMock{}

	desc := distribution.Descriptor{Digest: digest.FromString("x"), Size: 1}
	err := putBlobToLocal(context.Background(), local, "library/hello-world", "library/hello-world", desc, remote)
	p.Require().Error(err)
	local.AssertNotCalled(p.T(), "PushBlob", mock.Anything, mock.Anything, mock.Anything)
}

func (p *pushRetryTestSuite) TestPushManifestToLocalRetries() {
	local := &localInterfaceMock{}
	man := &mockManifest{}

	local.On("PushManifest", "library/hello-world", "latest", man).
		Return(errors.New("manifest unknown")).Once()
	local.On("PushManifest", "library/hello-world", "latest", man).
		Return(nil).Once()

	err := pushManifestToLocal(context.Background(), local, "library/hello-world", "latest", man)
	p.Require().NoError(err)
	local.AssertExpectations(p.T())
}

func TestPushRetryTestSuite(t *testing.T) {
	suite.Run(t, &pushRetryTestSuite{})
}
