// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/auth"

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
		log.Errorf("Error while decoding the json request, error: %v", err)
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

// ValidateUser checks if the request triggered by a valid user
func (b *BaseAPI) ValidateUser() int {
	userID, needsCheck, ok := b.GetUserIDForRequest()
	if !ok {
		log.Warning("No user id in session, canceling request")
		b.CustomAbort(http.StatusUnauthorized, "")
	}
	if needsCheck {
		u, err := dao.GetUser(models.User{UserID: userID})
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			b.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if u == nil {
			log.Warningf("User was deleted already, user id: %d, canceling request.", userID)
			b.CustomAbort(http.StatusUnauthorized, "")
		}
	}
	return userID
}

// GetUserIDForRequest tries to get user ID from basic auth header and session.
// It returns the user ID, whether need further verification(when the id is from session) and if the action is successful
func (b *BaseAPI) GetUserIDForRequest() (int, bool, bool) {
	username, password, ok := b.Ctx.Request.BasicAuth()
	if ok {
		log.Infof("Requst with Basic Authentication header, username: %s", username)
		user, err := auth.Login(models.AuthModel{
			Principal: username,
			Password:  password,
		})
		if err != nil {
			log.Errorf("Error while trying to login, username: %s, error: %v", username, err)
			user = nil
		}
		if user != nil {
			b.SetSession("userId", user.UserID)
			b.SetSession("username", user.Username)
			// User login successfully no further check required.
			return user.UserID, false, true
		}
	}
	sessionUserID, ok := b.GetSession("userId").(int)
	if ok {
		// The ID is from session
		return sessionUserID, true, true
	}
	log.Debug("No valid user id in session.")
	return 0, false, false
}

// Redirect does redirection to resource URI with http header status code.
func (b *BaseAPI) Redirect(statusCode int, resouceID string) {
	requestURI := b.Ctx.Request.RequestURI
	resoucreURI := requestURI + "/" + resouceID

	b.Ctx.Redirect(statusCode, resoucreURI)
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

	// SetPaginationHeader setprevious link
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

	// SetPaginationHeader setnext link
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
