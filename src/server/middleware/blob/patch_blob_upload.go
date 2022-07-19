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
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/middleware"
)

// PatchBlobUploadMiddleware middleware to record the accepted blob size for stream blob upload
func PatchBlobUploadMiddleware() func(http.Handler) http.Handler {
	return middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		// Only record when patch blob upload success
		if statusCode != http.StatusAccepted {
			return nil
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
