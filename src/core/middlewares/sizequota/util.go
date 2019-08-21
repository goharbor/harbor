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
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/opencontainers/go-digest"
)

var (
	blobUploadURLRe         = regexp.MustCompile(`^/v2/((?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)+)blobs/uploads/([a-zA-Z0-9-_.=]+)/?$`)
	initiateBlobUploadURLRe = regexp.MustCompile(`^/v2/((?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)+)blobs/uploads/?$`)
)

// parseUploadedBlobSize parse the blob stream upload response and return the size blob uploaded
func parseUploadedBlobSize(w http.ResponseWriter) (int64, error) {
	// Range: Range indicating the current progress of the upload.
	// https://github.com/opencontainers/distribution-spec/blob/master/spec.md#get-blob-upload
	r := w.Header().Get("Range")

	end := strings.Split(r, "-")[1]
	size, err := strconv.ParseInt(end, 10, 64)
	if err != nil {
		return 0, err
	}

	// docker registry did '-1' in the response
	if size > 0 {
		size = size + 1
	}

	return size, nil
}

// setUploadedBlobSize update the size of stream upload blob
func setUploadedBlobSize(uuid string, size int64) (bool, error) {
	conn, err := util.GetRegRedisCon()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	key := fmt.Sprintf("upload:%s:size", uuid)
	reply, err := redis.String(conn.Do("SET", key, size))
	if err != nil {
		return false, err
	}
	return reply == "OK", nil

}

// getUploadedBlobSize returns the size of stream upload blob
func getUploadedBlobSize(uuid string) (int64, error) {
	conn, err := util.GetRegRedisCon()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	key := fmt.Sprintf("upload:%s:size", uuid)
	size, err := redis.Int64(conn.Do("GET", key))
	if err != nil {
		return 0, err
	}

	return size, nil
}

// parseBlobSize returns blob size from blob upload complete request
func parseBlobSize(req *http.Request, uuid string) (int64, error) {
	size, err := strconv.ParseInt(req.Header.Get("Content-Length"), 10, 64)
	if err == nil && size != 0 {
		return size, nil
	}

	return getUploadedBlobSize(uuid)
}

// match returns true if request method equal method and path match re
func match(req *http.Request, method string, re *regexp.Regexp) bool {
	return req.Method == method && re.MatchString(req.URL.Path)
}

// parseBlobInfoFromComplete returns blob info from blob upload complete request
func parseBlobInfoFromComplete(req *http.Request) (*util.BlobInfo, error) {
	if !match(req, http.MethodPut, blobUploadURLRe) {
		return nil, fmt.Errorf("not match url %s for blob upload complete", req.URL.Path)
	}

	s := blobUploadURLRe.FindStringSubmatch(req.URL.Path)
	repository, uuid := s[1][:len(s[1])-1], s[2]

	projectName, _ := utils.ParseRepository(repository)
	project, err := dao.GetProjectByName(projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s, error: %v", projectName, err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %s not found", projectName)
	}

	dgt, err := digest.Parse(req.FormValue("digest"))
	if err != nil {
		return nil, fmt.Errorf("blob digest invalid for upload %s", uuid)
	}

	size, err := parseBlobSize(req, uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get content length of blob upload %s, error: %v", uuid, err)
	}

	return &util.BlobInfo{
		ProjectID:  project.ProjectID,
		Repository: repository,
		Digest:     dgt.String(),
		Size:       size,
	}, nil
}

// parseBlobInfoFromManifest returns blob info from put the manifest request
func parseBlobInfoFromManifest(req *http.Request) (*util.BlobInfo, error) {
	info, ok := util.ManifestInfoFromContext(req.Context())
	if !ok {
		manifest, err := util.ParseManifestInfoFromReq(req)
		if err != nil {
			return nil, err
		}

		info = manifest

		// replace the request with manifest info
		*req = *(req.WithContext(util.NewManifestInfoContext(req.Context(), info)))
	}

	return &util.BlobInfo{
		ProjectID:   info.ProjectID,
		Repository:  info.Repository,
		Digest:      info.Descriptor.Digest.String(),
		Size:        info.Descriptor.Size,
		ContentType: info.Descriptor.MediaType,
	}, nil
}

// parseBlobInfoFromMount returns blob info from blob mount request
func parseBlobInfoFromMount(req *http.Request) (*util.BlobInfo, error) {
	if !match(req, http.MethodPost, initiateBlobUploadURLRe) {
		return nil, fmt.Errorf("not match url %s for mount blob", req.URL.Path)
	}

	if req.FormValue("mount") == "" || req.FormValue("from") == "" {
		return nil, fmt.Errorf("not match url %s for mount blob", req.URL.Path)
	}

	dgt, err := digest.Parse(req.FormValue("mount"))
	if err != nil {
		return nil, errors.New("mount must be digest")
	}

	s := initiateBlobUploadURLRe.FindStringSubmatch(req.URL.Path)
	repository := strings.TrimSuffix(s[1], "/")

	projectName, _ := utils.ParseRepository(repository)
	project, err := dao.GetProjectByName(projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s, error: %v", projectName, err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %s not found", projectName)
	}

	blob, err := dao.GetBlob(dgt.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get blob %s, error: %v", dgt.String(), err)
	}
	if blob == nil {
		return nil, fmt.Errorf("the blob in the mount request with digest: %s doesn't exist", dgt.String())
	}

	return &util.BlobInfo{
		ProjectID:  project.ProjectID,
		Repository: repository,
		Digest:     dgt.String(),
		Size:       blob.Size,
	}, nil
}

// getBlobInfoParser return parse blob info function for request
// returns parseBlobInfoFromComplete when request match PUT /v2/<name>/blobs/uploads/<uuid>?digest=<digest>
// returns parseBlobInfoFromMount    when request match POST /v2/<name>/blobs/uploads/?mount=<digest>&from=<repository name>
func getBlobInfoParser(req *http.Request) func(*http.Request) (*util.BlobInfo, error) {
	if match(req, http.MethodPut, blobUploadURLRe) {
		if req.FormValue("digest") != "" {
			return parseBlobInfoFromComplete
		}
	}

	if match(req, http.MethodPost, initiateBlobUploadURLRe) {
		if req.FormValue("mount") != "" && req.FormValue("from") != "" {
			return parseBlobInfoFromMount
		}
	}

	return nil
}

// computeResourcesForBlob returns storage required for blob, no storage required if blob exists in project
func computeResourcesForBlob(req *http.Request) (types.ResourceList, error) {
	info, ok := util.BlobInfoFromContext(req.Context())
	if !ok {
		return nil, errors.New("blob info missing")
	}

	exist, err := info.BlobExists()
	if err != nil {
		return nil, err
	}

	if exist {
		return nil, nil
	}

	return types.ResourceList{types.ResourceStorage: info.Size}, nil
}

// computeResourcesForManifestCreation returns storage resource required for manifest
// no storage required if manifest exists in project
// the sum size of manifest itself and blobs not in project will return if manifest not exists in project
func computeResourcesForManifestCreation(req *http.Request) (types.ResourceList, error) {
	info, ok := util.ManifestInfoFromContext(req.Context())
	if !ok {
		return nil, errors.New("manifest info missing")
	}

	exist, err := info.ManifestExists()
	if err != nil {
		return nil, err
	}

	// manifest exist in project, so no storage quota required
	if exist {
		return nil, nil
	}

	blobs, err := info.GetBlobsNotInProject()
	if err != nil {
		return nil, err
	}

	size := info.Descriptor.Size

	for _, blob := range blobs {
		size += blob.Size
	}

	return types.ResourceList{types.ResourceStorage: size}, nil
}

// computeResourcesForManifestDeletion returns storage resource will be released when manifest deleted
// then result will be the sum of manifest itself and blobs which will not be used by other manifests of project
func computeResourcesForManifestDeletion(req *http.Request) (types.ResourceList, error) {
	info, ok := util.ManifestInfoFromContext(req.Context())
	if !ok {
		return nil, errors.New("manifest info missing")
	}

	blobs, err := dao.GetExclusiveBlobs(info.ProjectID, info.Repository, info.Digest)
	if err != nil {
		return nil, err
	}

	info.ExclusiveBlobs = blobs

	var size int64
	for _, blob := range blobs {
		size = size + blob.Size
	}

	return types.ResourceList{types.ResourceStorage: size}, nil
}

// syncBlobInfoToProject create the blob and add it to project
func syncBlobInfoToProject(info *util.BlobInfo) error {
	_, blob, err := dao.GetOrCreateBlob(&models.Blob{
		Digest:       info.Digest,
		ContentType:  info.ContentType,
		Size:         info.Size,
		CreationTime: time.Now(),
	})
	if err != nil {
		return err
	}

	if _, err := dao.AddBlobToProject(blob.ID, info.ProjectID); err != nil {
		return err
	}

	return nil
}
