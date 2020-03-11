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

package internal

import (
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type responseRecorderTestSuite struct {
	suite.Suite
	recorder *ResponseRecorder
}

func (r *responseRecorderTestSuite) SetupTest() {
	r.recorder = NewResponseRecorder(httptest.NewRecorder())
}

func (r *responseRecorderTestSuite) TestWriteHeader() {
	// write once
	r.recorder.WriteHeader(http.StatusInternalServerError)
	r.Equal(http.StatusInternalServerError, r.recorder.StatusCode)

	// write again
	r.recorder.WriteHeader(http.StatusNotFound)
	r.Equal(http.StatusInternalServerError, r.recorder.StatusCode)
}

func (r *responseRecorderTestSuite) TestWrite() {
	_, err := r.recorder.Write([]byte{'a'})
	r.Require().Nil(err)
	r.Equal(http.StatusOK, r.recorder.StatusCode)
}

func (r *responseRecorderTestSuite) TestSuccess() {
	r.recorder.WriteHeader(http.StatusInternalServerError)
	r.False(r.recorder.Success())
}

func TestResponseRecorder(t *testing.T) {
	suite.Run(t, &responseRecorderTestSuite{})
}
