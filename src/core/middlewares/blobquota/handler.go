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

package blobquota

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type blobQuotaHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &blobQuotaHandler{
		next: next,
	}
}

// ServeHTTP ...
func (bqh blobQuotaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPut {
		match, _ := util.MatchPutBlobURL(req)
		if match {
			dgstStr := req.FormValue("digest")
			if dgstStr == "" {
				http.Error(rw, util.MarshalError("InternalServerError", "blob digest missing"), http.StatusInternalServerError)
				return
			}
			dgst, err := digest.Parse(dgstStr)
			if err != nil {
				http.Error(rw, util.MarshalError("InternalServerError", "blob digest parsing failed"), http.StatusInternalServerError)
				return
			}
			// ToDo lock digest with redis

			// ToDo read placeholder from config
			state, err := hmacKey("placeholder").unpackUploadState(req.FormValue("_state"))
			if err != nil {
				http.Error(rw, util.MarshalError("InternalServerError", "failed to decode state"), http.StatusInternalServerError)
				return
			}
			log.Infof("we need to insert blob data into DB.")
			log.Infof("blob digest, %v", dgst)
			log.Infof("blob size, %v", state.Offset)
		}

	}
	bqh.next.ServeHTTP(rw, req)
}

// blobUploadState captures the state serializable state of the blob upload.
type blobUploadState struct {
	// name is the primary repository under which the blob will be linked.
	Name string

	// UUID identifies the upload.
	UUID string

	// offset contains the current progress of the upload.
	Offset int64

	// StartedAt is the original start time of the upload.
	StartedAt time.Time
}

type hmacKey string

var errInvalidSecret = errors.New("invalid secret")

// unpackUploadState unpacks and validates the blob upload state from the
// token, using the hmacKey secret.
func (secret hmacKey) unpackUploadState(token string) (blobUploadState, error) {
	var state blobUploadState

	tokenBytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return state, err
	}
	mac := hmac.New(sha256.New, []byte(secret))

	if len(tokenBytes) < mac.Size() {
		return state, errInvalidSecret
	}

	macBytes := tokenBytes[:mac.Size()]
	messageBytes := tokenBytes[mac.Size():]

	mac.Write(messageBytes)
	if !hmac.Equal(mac.Sum(nil), macBytes) {
		return state, errInvalidSecret
	}

	if err := json.Unmarshal(messageBytes, &state); err != nil {
		return state, err
	}

	return state, nil
}
