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

package error

import (
	"errors"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetHTTPStatusCode(t *testing.T) {
	// pre-defined error code
	errCode := ierror.NotFoundCode
	statusCode := getHTTPStatusCode(errCode)
	assert.Equal(t, http.StatusNotFound, statusCode)

	// not-defined error code
	errCode = "NOT_DEFINED_ERROR_CODE"
	statusCode = getHTTPStatusCode(errCode)
	assert.Equal(t, http.StatusInternalServerError, statusCode)
}

func TestAPIError(t *testing.T) {
	// ierror.Error
	err := &ierror.Error{
		Cause:   nil,
		Code:    ierror.NotFoundCode,
		Message: "resource not found",
	}
	statusCode, payload := APIError(err)
	assert.Equal(t, http.StatusNotFound, statusCode)
	assert.Equal(t, `{"errors":[{"code":"NOT_FOUND","message":"resource not found"}]}`, payload)

	// common error
	e := errors.New("customized error")
	statusCode, payload = APIError(e)
	assert.Equal(t, http.StatusInternalServerError, statusCode)
	assert.Equal(t, `{"errors":[{"code":"UNKNOWN","message":"customized error"}]}`, payload)
}
