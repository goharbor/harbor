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

package util

import (
	"context"
	"net/http"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/distribution"
)

// AcceptedBlobSizeFunc returns the size already accepted for a blob upload session.
type AcceptedBlobSizeFunc func(ctx context.Context, sessionID string) (int64, error)

// ValidateBlobUploadLength rejects a request completing a blob upload whose body
// length is unknown (e.g. Transfer-Encoding: chunked). Such a body cannot be
// accounted for against the project quota before it is written.
func ValidateBlobUploadLength(r *http.Request) error {
	if r.ContentLength < 0 {
		return errors.BadRequestError(nil).
			WithMessage("Content-Length is required to complete a blob upload")
	}
	return nil
}

// PutBlobUploadSize returns the total size of the blob completed by
// PUT /v2/<name>/blobs/uploads/<session_id>: the chunks already accepted by
// PATCH in this session plus the body of the PUT request itself.
func PutBlobUploadSize(r *http.Request, acceptedSize AcceptedBlobSizeFunc) (int64, error) {
	if err := ValidateBlobUploadLength(r); err != nil {
		return 0, err
	}
	accepted, err := acceptedSize(r.Context(), distribution.ParseSessionID(r.URL.Path))
	if err != nil {
		return 0, err
	}
	return accepted + r.ContentLength, nil
}
