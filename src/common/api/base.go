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

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/astaxie/beego/validation"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/utils/log"

	"github.com/astaxie/beego"
)

const (
	defaultPageSize int64 = 500
	maxPageSize     int64 = 500
)

// BaseAPI wraps common methods for controllers to host API
type BaseAPI struct {
	beego.Controller
}

// GetStringFromPath gets the param from path and returns it as string
func (b *BaseAPI) GetStringFromPath(key string) string {
	return b.Ctx.Input.Param(key)
}

// GetInt64FromPath gets the param from path and returns it as int64
func (b *BaseAPI) GetInt64FromPath(key string) (int64, error) {
	value := b.Ctx.Input.Param(key)
	return strconv.ParseInt(value, 10, 64)
}

// HandleNotFound ...
func (b *BaseAPI) HandleNotFound(text string) {
	log.Info(text)
	b.RenderError(http.StatusNotFound, text)
}

// HandleUnauthorized ...
func (b *BaseAPI) HandleUnauthorized() {
	log.Info("unauthorized")
	b.RenderError(http.StatusUnauthorized, "")
}

// HandleForbidden ...
func (b *BaseAPI) HandleForbidden(text string) {
	log.Infof("forbidden: %s", text)
	b.RenderError(http.StatusForbidden, text)
}

// HandleBadRequest ...
func (b *BaseAPI) HandleBadRequest(text string) {
	log.Info(text)
	b.RenderError(http.StatusBadRequest, text)
}

// HandleStatusPreconditionFailed ...
func (b *BaseAPI) HandleStatusPreconditionFailed(text string) {
	log.Info(text)
	b.RenderError(http.StatusPreconditionFailed, text)
}

// HandleConflict ...
func (b *BaseAPI) HandleConflict(text ...string) {
	msg := ""
	if len(text) > 0 {
		msg = text[0]
	}
	log.Infof("conflict: %s", msg)

	b.RenderError(http.StatusConflict, msg)
}

// HandleInternalServerError ...
func (b *BaseAPI) HandleInternalServerError(text string) {
	log.Error(text)
	b.RenderError(http.StatusInternalServerError, "")
}

// ParseAndHandleError : if the err is an instance of utils/error.Error,
// return the status code and the detail message contained in err, otherwise
// return 500
func (b *BaseAPI) ParseAndHandleError(text string, err error) {
	if err == nil {
		return
	}
	log.Errorf("%s: %v", text, err)
	if e, ok := err.(*commonhttp.Error); ok {
		b.RenderError(e.Code, e.Message)
		return
	}
	b.RenderError(http.StatusInternalServerError, "")
}

// Render returns nil as it won't render template
func (b *BaseAPI) Render() error {
	return nil
}

// RenderError provides shortcut to render http error
func (b *BaseAPI) RenderError(code int, text string) {
	http.Error(b.Ctx.ResponseWriter, text, code)
}

// DecodeJSONReq decodes a json request
func (b *BaseAPI) DecodeJSONReq(v interface{}) {
	err := json.Unmarshal(b.Ctx.Input.CopyBody(1<<32), v)
	if err != nil {
		log.Errorf("Error while decoding the json request, error: %v, %v",
			err, string(b.Ctx.Input.CopyBody(1 << 32)[:]))
		b.CustomAbort(http.StatusBadRequest, "Invalid json request")
	}
}

// Validate validates v if it implements interface validation.ValidFormer
func (b *BaseAPI) Validate(v interface{}) {
	validator := validation.Validation{}
	isValid, err := validator.Valid(v)
	if err != nil {
		log.Errorf("failed to validate: %v", err)
		b.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if !isValid {
		message := ""
		for _, e := range validator.Errors {
			message += fmt.Sprintf("%s %s \n", e.Field, e.Message)
		}
		b.CustomAbort(http.StatusBadRequest, message)
	}
}

// DecodeJSONReqAndValidate does both decoding and validation
func (b *BaseAPI) DecodeJSONReqAndValidate(v interface{}) {
	b.DecodeJSONReq(v)
	b.Validate(v)
}

// Redirect does redirection to resource URI with http header status code.
func (b *BaseAPI) Redirect(statusCode int, resouceID string) {
	requestURI := b.Ctx.Request.RequestURI
	resourceURI := requestURI + "/" + resouceID

	b.Ctx.Redirect(statusCode, resourceURI)
}

// GetIDFromURL checks the ID in request URL
func (b *BaseAPI) GetIDFromURL() int64 {
	idStr := b.Ctx.Input.Param(":id")
	if len(idStr) == 0 {
		b.CustomAbort(http.StatusBadRequest, "invalid ID in URL")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		b.CustomAbort(http.StatusBadRequest, "invalid ID in URL")
	}

	return id
}

// SetPaginationHeader set"Link" and "X-Total-Count" header for pagination request
func (b *BaseAPI) SetPaginationHeader(total, page, pageSize int64) {
	b.Ctx.ResponseWriter.Header().Set("X-Total-Count", strconv.FormatInt(total, 10))

	link := ""

	// SetPaginationHeader set previous link
	if page > 1 && (page-1)*pageSize <= total {
		u := *(b.Ctx.Request.URL)
		q := u.Query()
		q.Set("page", strconv.FormatInt(page-1, 10))
		u.RawQuery = q.Encode()
		if len(link) != 0 {
			link += ", "
		}
		link += fmt.Sprintf("<%s>; rel=\"prev\"", u.String())
	}

	// SetPaginationHeader set next link
	if pageSize*page < total {
		u := *(b.Ctx.Request.URL)
		q := u.Query()
		q.Set("page", strconv.FormatInt(page+1, 10))
		u.RawQuery = q.Encode()
		if len(link) != 0 {
			link += ", "
		}
		link += fmt.Sprintf("<%s>; rel=\"next\"", u.String())
	}

	if len(link) != 0 {
		b.Ctx.ResponseWriter.Header().Set("Link", link)
	}
}

// GetPaginationParams ...
func (b *BaseAPI) GetPaginationParams() (page, pageSize int64) {
	page, err := b.GetInt64("page", 1)
	if err != nil || page <= 0 {
		b.CustomAbort(http.StatusBadRequest, "invalid page")
	}

	pageSize, err = b.GetInt64("page_size", defaultPageSize)
	if err != nil || pageSize <= 0 {
		b.CustomAbort(http.StatusBadRequest, "invalid page_size")
	}

	if pageSize > maxPageSize {
		pageSize = maxPageSize
		log.Debugf("the parameter page_size %d exceeds the max %d, set it to max", pageSize, maxPageSize)
	}

	return page, pageSize
}
