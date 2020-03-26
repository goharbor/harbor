// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lib

import (
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type responseBufferTestSuite struct {
	suite.Suite
	recorder *httptest.ResponseRecorder
	buffer   *ResponseBuffer
}

func (r *responseBufferTestSuite) SetupTest() {
	r.recorder = httptest.NewRecorder()
	r.buffer = NewResponseBuffer(r.recorder)
}

func (r *responseBufferTestSuite) TestWriteHeader() {
	// write once
	r.buffer.WriteHeader(http.StatusInternalServerError)
	r.Equal(http.StatusInternalServerError, r.buffer.code)
	r.Equal(http.StatusOK, r.recorder.Code)

	// write again
	r.buffer.WriteHeader(http.StatusNotFound)
	r.Equal(http.StatusInternalServerError, r.buffer.code)
	r.Equal(http.StatusOK, r.recorder.Code)
}

func (r *responseBufferTestSuite) TestWrite() {
	_, err := r.buffer.Write([]byte{'a'})
	r.Require().Nil(err)
	r.Equal([]byte{'a'}, r.buffer.buffer.Bytes())
	r.Empty(r.recorder.Body.Bytes())

	// try to write header after calling write
	r.buffer.WriteHeader(http.StatusNotFound)
	r.Equal(http.StatusOK, r.buffer.code)
}

func (r *responseBufferTestSuite) TestHeader() {
	header := r.buffer.Header()
	header.Add("k", "v")
	r.Equal("v", r.buffer.header.Get("k"))
	r.Empty(r.recorder.Header())
}
func (r *responseBufferTestSuite) TestFlush() {
	r.buffer.WriteHeader(http.StatusOK)
	_, err := r.buffer.Write([]byte{'a'})
	r.Require().Nil(err)
	_, err = r.buffer.Flush()
	r.Require().Nil(err)
	r.Equal(http.StatusOK, r.recorder.Code)
	r.Equal([]byte{'a'}, r.recorder.Body.Bytes())
}

func (r *responseBufferTestSuite) TestSuccess() {
	r.buffer.WriteHeader(http.StatusInternalServerError)
	r.False(r.buffer.Success())

	// reset wroteHeader
	r.buffer.wroteHeader = false
	r.buffer.WriteHeader(http.StatusOK)
	r.True(r.buffer.Success())
}

func TestResponseBuffer(t *testing.T) {
	suite.Run(t, &responseBufferTestSuite{})
}
