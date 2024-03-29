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

package blob

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/middleware"
)

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

// /v2/conformance/testrepo/blobs/uploads/96a6fe7e-6683-4dce-a0d5-80fdc50f3822?_state=PwxpQahplvWdosoKCWat7zK2PtCo_4pUEEAmWzV2YOl7Ik5hbWUiOiJjb25mb3JtYW5jZS90ZXN0cmVwbyIsIlVVSUQiOiI5NmE2ZmU3ZS02NjgzLTRkY2UtYTBkNS04MGZkYzUwZjM4MjIiLCJPZmZzZXQiOjAsIlN0YXJ0ZWRBdCI6IjIwMjQtMDMtMTBUMTU6MzA6MjMuMjE1MjU1MzVaIn0%3D
func unpackUploadState(r *http.Request) (blobUploadState, error) {
	var state blobUploadState
	token := r.FormValue("_state")
	tokenBytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return state, err
	}

	secret := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer")
	mac := hmac.New(sha256.New, []byte(secret))
	if len(tokenBytes) < mac.Size() {
		return state, err
	}
	macBytes := tokenBytes[:mac.Size()]
	messageBytes := tokenBytes[mac.Size():]

	mac.Write(messageBytes)
	if !hmac.Equal(mac.Sum(nil), macBytes) {
		return state, err
	}

	if err := json.Unmarshal(messageBytes, &state); err != nil {
		return state, err
	}

	return state, nil
}

func isDisorder(state *blobUploadState, r *http.Request) (bool, error) {
	cntRange := r.Header.Get("Content-Range")
	startstr := strings.Split(cntRange, "-")[0]
	offset := state.Offset

	start, err := strconv.ParseInt(startstr, 10, 64)
	if err != nil {
		return false, err
	}
	if start > offset {
		return true, nil
	}
	return false, nil
}

// PatchBlobUploadMiddleware middleware to record the accepted blob size for stream blob upload
func PatchBlobUploadMiddleware() func(http.Handler) http.Handler {
	return middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		// Only record when patch blob upload success
		if statusCode != http.StatusAccepted {
			return nil
		}
		//check if disorder when upload by chunk
		state, err := unpackUploadState(r)
		if err != nil {
			return err
		}
		disorder, err := isDisorder(&state, r)
		if err != nil {
			return err
		}
		if disorder {
			return errors.New(nil).WithCode(errors.RangeUnsatisfy).WithMessage("Request Range is disordered")
		}

		size, err := parseAcceptedBlobSize(w.Header().Get("Range"))
		if err != nil {
			return err
		}

		sessionID := distribution.ParseSessionID(r.URL.Path)

		return blobController.SetAcceptedBlobSize(r.Context(), sessionID, size)
	})
}

// parseAcceptedBlobSize parse the blob stream upload response and return the size blob accepted
func parseAcceptedBlobSize(rangeHeader string) (int64, error) {
	// Range: Range indicating the current progress of the upload.
	// https://github.com/opencontainers/distribution-spec/blob/master/spec.md#get-blob-upload
	if rangeHeader == "" {
		return 0, fmt.Errorf("range header required")
	}

	parts := strings.SplitN(rangeHeader, "-", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("range header bad value: %s", rangeHeader)
	}

	size, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, err
	}

	// docker registry did '-1' in the response
	if size > 0 {
		size = size + 1
	}

	return size, nil
}
