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
	"fmt"
	openapi "github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/utils/log"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"net/http"
	"strings"
)

var (
	codeMap = map[string]int{
		ierror.BadRequestCode:                  http.StatusBadRequest,
		ierror.DIGESTINVALID:                   http.StatusBadRequest,
		ierror.UnAuthorizedCode:                http.StatusUnauthorized,
		ierror.ForbiddenCode:                   http.StatusForbidden,
		ierror.DENIED:                          http.StatusForbidden,
		ierror.NotFoundCode:                    http.StatusNotFound,
		ierror.ConflictCode:                    http.StatusConflict,
		ierror.PreconditionCode:                http.StatusPreconditionFailed,
		ierror.ViolateForeignKeyConstraintCode: http.StatusPreconditionFailed,
		ierror.PROJECTPOLICYVIOLATION:          http.StatusPreconditionFailed,
		ierror.GeneralCode:                     http.StatusInternalServerError,
	}
)

// TODO use "SendError" instead in the v1 APIs?

// SendError tries to parse the HTTP status code from the specified error, envelops it into
// an error array as the error payload and returns the code and payload to the response.
// And the error is logged as well
func SendError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	statusCode, errPayload := apiError(err)
	// the error detail is logged only, and will not be sent to the client to avoid leaking server information
	if statusCode >= http.StatusInternalServerError {
		log.Error(errPayload)
		err = ierror.New(nil).WithCode(ierror.GeneralCode).WithMessage("internal server error")
		errPayload = ierror.NewErrs(err).Error()
	} else {
		// only log the error whose status code < 500 when debugging to avoid log flooding
		log.Debug(errPayload)
	}
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, errPayload)
}

// generates the HTTP status code based on the specified error,
// envelops the error into an error array as the payload and return them
func apiError(err error) (statusCode int, errPayload string) {
	code := 0
	var openAPIErr openapi.Error
	if errors.As(err, &openAPIErr) {
		// Before executing operation handler, go-swagger will bind a parameters object to a request and validate the request,
		// it will return directly when bind and validate failed.
		// The response format of the default ServeError implementation does not match the internal error response format.
		// So we needed to convert the format to the internal error response format.
		code = int(openAPIErr.Code())
		errCode := strings.Replace(strings.ToUpper(http.StatusText(code)), " ", "_", -1)
		err = ierror.New(nil).WithCode(errCode).WithMessage(openAPIErr.Error())
	} else {
		code = codeMap[ierror.ErrCode(err)]
	}
	if code == 0 {
		code = http.StatusInternalServerError
	}
	return code, ierror.NewErrs(err).Error()
}

var _ middleware.Responder = &ErrResponder{}

// ErrResponder error responder
type ErrResponder struct {
	err error
}

// WriteResponse ...
func (r *ErrResponder) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	SendError(rw, r.err)
}

// NewErrResponder returns responder for err
func NewErrResponder(err error) *ErrResponder {
	return &ErrResponder{err: err}
}
