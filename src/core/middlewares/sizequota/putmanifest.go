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

package sizequota

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"io/ioutil"
	"net/http"
	"strings"
)

// PutManifestInterceptor ...
type PutManifestInterceptor struct {
	blobInfo *util.BlobInfo
	mfInfo   *util.MfInfo
}

// NewPutManifestInterceptor ...
func NewPutManifestInterceptor(blobInfo *util.BlobInfo, mfInfo *util.MfInfo) *PutManifestInterceptor {
	return &PutManifestInterceptor{
		blobInfo: blobInfo,
		mfInfo:   mfInfo,
	}
}

// HandleRequest ...
func (pmi *PutManifestInterceptor) HandleRequest(req *http.Request) error {
	mediaType := req.Header.Get("Content-Type")
	if mediaType == schema1.MediaTypeManifest ||
		mediaType == schema1.MediaTypeSignedManifest ||
		mediaType == schema2.MediaTypeManifest {

		con, err := util.GetRegRedisCon()
		if err != nil {
			log.Infof("failed to get registry redis connection, %v", err)
			return err
		}

		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Warningf("Error occurred when to copy manifest body %v", err)
			return err
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		manifest, desc, err := distribution.UnmarshalManifest(mediaType, data)
		if err != nil {
			log.Warningf("Error occurred when to Unmarshal Manifest %v", err)
			return err
		}
		projectID, err := util.GetProjectID(strings.Split(pmi.mfInfo.Repository, "/")[0])
		if err != nil {
			log.Warningf("Error occurred when to get project ID %v", err)
			return err
		}

		pmi.mfInfo.ProjectID = projectID
		pmi.mfInfo.Refrerence = manifest.References()
		pmi.mfInfo.Digest = desc.Digest.String()
		pmi.blobInfo.ProjectID = projectID
		pmi.blobInfo.Digest = desc.Digest.String()
		pmi.blobInfo.Size = desc.Size
		pmi.blobInfo.ContentType = mediaType

		if err := requireQuota(con, pmi.blobInfo); err != nil {
			return err
		}

		*req = *(req.WithContext(context.WithValue(req.Context(), util.MFInfokKey, pmi.mfInfo)))
		*req = *(req.WithContext(context.WithValue(req.Context(), util.BBInfokKey, pmi.blobInfo)))

		return nil
	}

	return fmt.Errorf("unsupported content type for manifest: %s", mediaType)
}

// HandleResponse ...
func (pmi *PutManifestInterceptor) HandleResponse(rw util.CustomResponseWriter, req *http.Request) {
	if err := HandleBlobCommon(rw, req); err != nil {
		log.Error(err)
		return
	}
}
