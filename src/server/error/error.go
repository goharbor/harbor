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
	"net/http"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	ierror "github.com/goharbor/harbor/src/internal/error"
)

var (
	codeMap = map[string]int{
		ierror.BadRequestCode:   http.StatusBadRequest,
		ierror.UnAuthorizedCode: http.StatusUnauthorized,
		ierror.ForbiddenCode:    http.StatusForbidden,
		ierror.NotFoundCode:     http.StatusNotFound,
		ierror.ConflictCode:     http.StatusConflict,
		ierror.PreconditionCode: http.StatusPreconditionFailed,
		ierror.GeneralCode:      http.StatusInternalServerError,
	}
)

// APIError generates the HTTP status code and error payload based on the input err
func APIError(err error) (int, string) {
	return getHTTPStatusCode(ierror.ErrCode(err)), ierror.NewErrs(err).Error()
}

func getHTTPStatusCode(errCode string) int {
	statusCode, ok := codeMap[errCode]
	if !ok {
		statusCode = http.StatusInternalServerError
	}
	return statusCode
}

var _ middleware.Responder = &ErrResponder{}

// ErrResponder error responder
type ErrResponder struct {
	err error
}

// WriteResponse ...
func (r *ErrResponder) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	code := ierror.ErrCode(r.err)
	rw.WriteHeader(getHTTPStatusCode(code))

	var e *ierror.Error
	if !errors.As(r.err, &e) {
		e = &ierror.Error{
			Code:    code,
			Message: r.err.Error(),
		}
	}

	if err := producer.Produce(rw, e); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// NewErrResponder returns responder for err
func NewErrResponder(err error) *ErrResponder {
	return &ErrResponder{err: err}
}
