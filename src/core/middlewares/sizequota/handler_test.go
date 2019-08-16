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
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/countquota"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/suite"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func genUUID() string {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return ""
	}

	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func getProjectCountUsage(projectID int64) (int64, error) {
	usage := models.QuotaUsage{Reference: "project", ReferenceID: fmt.Sprintf("%d", projectID)}
	err := dao.GetOrmer().Read(&usage, "reference", "reference_id")
	if err != nil {
		return 0, err
	}
	used, err := types.NewResourceList(usage.Used)
	if err != nil {
		return 0, err
	}

	return used[types.ResourceCount], nil
}

func getProjectStorageUsage(projectID int64) (int64, error) {
	usage := models.QuotaUsage{Reference: "project", ReferenceID: fmt.Sprintf("%d", projectID)}
	err := dao.GetOrmer().Read(&usage, "reference", "reference_id")
	if err != nil {
		return 0, err
	}
	used, err := types.NewResourceList(usage.Used)
	if err != nil {
		return 0, err
	}

	return used[types.ResourceStorage], nil
}

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}

func makeManifest(configSize int64, layerSizes []int64) schema2.Manifest {
	manifest := schema2.Manifest{
		Versioned: manifest.Versioned{SchemaVersion: 2, MediaType: schema2.MediaTypeManifest},
		Config: distribution.Descriptor{
			MediaType: schema2.MediaTypeImageConfig,
			Size:      configSize,
			Digest:    digest.FromString(randomString(15)),
		},
	}

	for _, size := range layerSizes {
		manifest.Layers = append(manifest.Layers, distribution.Descriptor{
			MediaType: schema2.MediaTypeLayer,
			Size:      size,
			Digest:    digest.FromString(randomString(15)),
		})
	}

	return manifest
}

func manifestWithAdditionalLayers(raw schema2.Manifest, layerSizes []int64) schema2.Manifest {
	var manifest schema2.Manifest

	manifest.Versioned = raw.Versioned
	manifest.Config = raw.Config
	manifest.Layers = append(manifest.Layers, raw.Layers...)

	for _, size := range layerSizes {
		manifest.Layers = append(manifest.Layers, distribution.Descriptor{
			MediaType: schema2.MediaTypeLayer,
			Size:      size,
			Digest:    digest.FromString(randomString(15)),
		})
	}

	return manifest
}

func digestOfManifest(manifest schema2.Manifest) string {
	bytes, _ := json.Marshal(manifest)

	return digest.FromBytes(bytes).String()
}

func sizeOfManifest(manifest schema2.Manifest) int64 {
	bytes, _ := json.Marshal(manifest)

	return int64(len(bytes))
}

func sizeOfImage(manifest schema2.Manifest) int64 {
	totalSizeOfLayers := manifest.Config.Size

	for _, layer := range manifest.Layers {
		totalSizeOfLayers += layer.Size
	}

	return sizeOfManifest(manifest) + totalSizeOfLayers
}

func doHandle(req *http.Request, next ...http.HandlerFunc) int {
	rr := httptest.NewRecorder()

	var n http.HandlerFunc
	if len(next) > 0 {
		n = next[0]
	} else {
		n = func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}
	}

	h := New(http.HandlerFunc(n))
	h.ServeHTTP(util.NewCustomResponseWriter(rr), req)

	return rr.Code
}

func patchBlobUpload(projectName, name, uuid, blobDigest string, chunkSize int64) {
	repository := fmt.Sprintf("%s/%s", projectName, name)

	url := fmt.Sprintf("/v2/%s/blobs/uploads/%s?digest=%s", repository, uuid, blobDigest)
	req, _ := http.NewRequest(http.MethodPatch, url, nil)

	doHandle(req, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Header().Add("Range", fmt.Sprintf("0-%d", chunkSize-1))
	})
}

func putBlobUpload(projectName, name, uuid, blobDigest string, blobSize ...int64) {
	repository := fmt.Sprintf("%s/%s", projectName, name)

	url := fmt.Sprintf("/v2/%s/blobs/uploads/%s?digest=%s", repository, uuid, blobDigest)
	req, _ := http.NewRequest(http.MethodPut, url, nil)
	if len(blobSize) > 0 {
		req.Header.Add("Content-Length", strconv.FormatInt(blobSize[0], 10))
	}

	doHandle(req, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
}

func mountBlob(projectName, name, blobDigest, fromRepository string) {
	repository := fmt.Sprintf("%s/%s", projectName, name)

	url := fmt.Sprintf("/v2/%s/blobs/uploads/?mount=%s&from=%s", repository, blobDigest, fromRepository)
	req, _ := http.NewRequest(http.MethodPost, url, nil)

	doHandle(req, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
}

func deleteManifest(projectName, name, digest string, accepted ...func() bool) {
	repository := fmt.Sprintf("%s/%s", projectName, name)

	url := fmt.Sprintf("/v2/%s/manifests/%s", repository, digest)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)

	next := countquota.New(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if len(accepted) > 0 {
			if accepted[0]() {
				w.WriteHeader(http.StatusAccepted)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}

			return
		}

		w.WriteHeader(http.StatusAccepted)
	}))

	rr := httptest.NewRecorder()
	h := New(next)
	h.ServeHTTP(util.NewCustomResponseWriter(rr), req)
}

func putManifest(projectName, name, tag string, manifest schema2.Manifest) {
	repository := fmt.Sprintf("%s/%s", projectName, name)

	buf, _ := json.Marshal(manifest)

	url := fmt.Sprintf("/v2/%s/manifests/%s", repository, tag)
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewReader(buf))
	req.Header.Add("Content-Type", manifest.MediaType)

	next := countquota.New(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	rr := httptest.NewRecorder()
	h := New(next)
	h.ServeHTTP(util.NewCustomResponseWriter(rr), req)
}

func pushImage(projectName, name, tag string, manifest schema2.Manifest) {
	putBlobUpload(projectName, name, genUUID(), manifest.Config.Digest.String(), manifest.Config.Size)
	for _, layer := range manifest.Layers {
		putBlobUpload(projectName, name, genUUID(), layer.Digest.String(), layer.Size)
	}

	putManifest(projectName, name, tag, manifest)
}

func withProject(f func(int64, string)) {
	projectName := randomString(5)

	projectID, err := dao.AddProject(models.Project{
		Name:    projectName,
		OwnerID: 1,
	})
	if err != nil {
		panic(err)
	}

	defer func() {
		dao.DeleteProject(projectID)
	}()

	f(projectID, projectName)
}

type HandlerSuite struct {
	suite.Suite
}

func (suite *HandlerSuite) checkCountUsage(expected, projectID int64) {
	count, err := getProjectCountUsage(projectID)
	suite.Nil(err, fmt.Sprintf("Failed to get count usage of project %d, error: %v", projectID, err))
	suite.Equal(expected, count, "Failed to check count usage for project %d", projectID)
}

func (suite *HandlerSuite) checkStorageUsage(expected, projectID int64) {
	value, err := getProjectStorageUsage(projectID)
	suite.Nil(err, fmt.Sprintf("Failed to get storage usage of project %d, error: %v", projectID, err))
	suite.Equal(expected, value, "Failed to check storage usage for project %d", projectID)
}

func (suite *HandlerSuite) TearDownTest() {
	for _, table := range []string{
		"artifact", "blob",
		"artifact_blob", "project_blob",
		"quota", "quota_usage",
	} {
		dao.ClearTable(table)
	}
}

func (suite *HandlerSuite) TestPatchBlobUpload() {
	withProject(func(projectID int64, projectName string) {
		uuid := genUUID()
		blobDigest := digest.FromString(randomString(15)).String()
		patchBlobUpload(projectName, "photon", uuid, blobDigest, 1024)
		size, err := getUploadedBlobSize(uuid)
		suite.Nil(err)
		suite.Equal(int64(1024), size)
	})
}

func (suite *HandlerSuite) TestPutBlobUpload() {
	withProject(func(projectID int64, projectName string) {
		uuid := genUUID()
		blobDigest := digest.FromString(randomString(15)).String()
		putBlobUpload(projectName, "photon", uuid, blobDigest, 1024)
		suite.checkStorageUsage(1024, projectID)

		blob, err := dao.GetBlob(blobDigest)
		suite.Nil(err)
		suite.Equal(int64(1024), blob.Size)
	})
}

func (suite *HandlerSuite) TestPutBlobUploadWithPatch() {
	withProject(func(projectID int64, projectName string) {
		uuid := genUUID()
		blobDigest := digest.FromString(randomString(15)).String()
		patchBlobUpload(projectName, "photon", uuid, blobDigest, 1024)

		putBlobUpload(projectName, "photon", uuid, blobDigest)
		suite.checkStorageUsage(1024, projectID)

		blob, err := dao.GetBlob(blobDigest)
		suite.Nil(err)
		suite.Equal(int64(1024), blob.Size)
	})
}

func (suite *HandlerSuite) TestMountBlob() {
	withProject(func(projectID int64, projectName string) {
		blobDigest := digest.FromString(randomString(15)).String()
		putBlobUpload(projectName, "photon", genUUID(), blobDigest, 1024)
		suite.checkStorageUsage(1024, projectID)

		repository := fmt.Sprintf("%s/%s", projectName, "photon")

		withProject(func(projectID int64, projectName string) {
			mountBlob(projectName, "harbor", blobDigest, repository)
			suite.checkStorageUsage(1024, projectID)
		})
	})
}

func (suite *HandlerSuite) TestPutManifestCreated() {
	withProject(func(projectID int64, projectName string) {
		manifest := makeManifest(100, []int64{100, 100})

		putBlobUpload(projectName, "photon", genUUID(), manifest.Config.Digest.String(), manifest.Config.Size)
		for _, layer := range manifest.Layers {
			putBlobUpload(projectName, "photon", genUUID(), layer.Digest.String(), layer.Size)
		}

		putManifest(projectName, "photon", "latest", manifest)

		suite.checkStorageUsage(int64(300+sizeOfManifest(manifest)), projectID)
	})
}

func (suite *HandlerSuite) TestDeleteManifest() {
	withProject(func(projectID int64, projectName string) {
		manifest := makeManifest(1, []int64{2, 3, 4, 5})
		size := sizeOfImage(manifest)

		pushImage(projectName, "photon", "latest", manifest)
		suite.checkStorageUsage(size, projectID)

		deleteManifest(projectName, "photon", digestOfManifest(manifest))
		suite.checkStorageUsage(0, projectID)
	})
}

func (suite *HandlerSuite) TestImageOverwrite() {
	withProject(func(projectID int64, projectName string) {
		manifest1 := makeManifest(1, []int64{2, 3, 4, 5})
		size1 := sizeOfImage(manifest1)
		pushImage(projectName, "photon", "latest", manifest1)
		suite.checkCountUsage(1, projectID)
		suite.checkStorageUsage(size1, projectID)

		manifest2 := makeManifest(1, []int64{2, 3, 4, 5})
		size2 := sizeOfImage(manifest2)
		pushImage(projectName, "photon", "latest", manifest2)
		suite.checkCountUsage(1, projectID)
		suite.checkStorageUsage(size1+size2, projectID)

		manifest3 := makeManifest(1, []int64{2, 3, 4, 5})
		size3 := sizeOfImage(manifest2)
		pushImage(projectName, "photon", "latest", manifest3)
		suite.checkCountUsage(1, projectID)
		suite.checkStorageUsage(size1+size2+size3, projectID)
	})
}

func (suite *HandlerSuite) TestPushImageMultiTimes() {
	withProject(func(projectID int64, projectName string) {
		manifest := makeManifest(1, []int64{2, 3, 4, 5})
		size := sizeOfImage(manifest)

		pushImage(projectName, "photon", "latest", manifest)
		suite.checkCountUsage(1, projectID)
		suite.checkStorageUsage(size, projectID)

		pushImage(projectName, "photon", "latest", manifest)
		suite.checkCountUsage(1, projectID)
		suite.checkStorageUsage(size, projectID)

		pushImage(projectName, "photon", "latest", manifest)
		suite.checkCountUsage(1, projectID)
		suite.checkStorageUsage(size, projectID)
	})
}

func (suite *HandlerSuite) TestPushImageToSameRepository() {
	withProject(func(projectID int64, projectName string) {
		manifest := makeManifest(1, []int64{2, 3, 4, 5})
		size := sizeOfImage(manifest)

		pushImage(projectName, "photon", "latest", manifest)
		suite.checkCountUsage(1, projectID)
		suite.checkStorageUsage(size, projectID)

		pushImage(projectName, "photon", "dev", manifest)
		suite.checkCountUsage(2, projectID)
		suite.checkStorageUsage(size, projectID)
	})
}

func (suite *HandlerSuite) TestPushImageToDifferentRepositories() {
	withProject(func(projectID int64, projectName string) {
		manifest := makeManifest(1, []int64{2, 3, 4, 5})
		size := sizeOfImage(manifest)

		pushImage(projectName, "mysql", "latest", manifest)
		suite.checkStorageUsage(size, projectID)

		pushImage(projectName, "redis", "latest", manifest)
		suite.checkStorageUsage(size+sizeOfManifest(manifest), projectID)

		pushImage(projectName, "postgres", "latest", manifest)
		suite.checkStorageUsage(size+2*sizeOfManifest(manifest), projectID)
	})
}

func (suite *HandlerSuite) TestPushImageToDifferentProjects() {
	withProject(func(projectID int64, projectName string) {
		manifest := makeManifest(1, []int64{2, 3, 4, 5})
		size := sizeOfImage(manifest)

		pushImage(projectName, "mysql", "latest", manifest)
		suite.checkStorageUsage(size, projectID)

		withProject(func(id int64, name string) {
			manifest := makeManifest(1, []int64{2, 3, 4, 5})
			size := sizeOfImage(manifest)

			pushImage(name, "mysql", "latest", manifest)
			suite.checkStorageUsage(size, id)

			suite.checkStorageUsage(size, projectID)
		})
	})
}

func (suite *HandlerSuite) TestDeleteManifestShareLayersInSameRepository() {
	withProject(func(projectID int64, projectName string) {
		manifest1 := makeManifest(1, []int64{2, 3, 4, 5})
		size1 := sizeOfImage(manifest1)

		pushImage(projectName, "mysql", "latest", manifest1)
		suite.checkCountUsage(1, projectID)
		suite.checkStorageUsage(size1, projectID)

		manifest2 := manifestWithAdditionalLayers(manifest1, []int64{6, 7})
		pushImage(projectName, "mysql", "dev", manifest2)
		suite.checkCountUsage(2, projectID)

		totalSize := size1 + sizeOfManifest(manifest2) + 6 + 7
		suite.checkStorageUsage(totalSize, projectID)

		deleteManifest(projectName, "mysql", digestOfManifest(manifest1))
		suite.checkCountUsage(1, projectID)
		suite.checkStorageUsage(totalSize-sizeOfManifest(manifest1), projectID)
	})
}

func (suite *HandlerSuite) TestDeleteManifestShareLayersInDifferentRepositories() {
	withProject(func(projectID int64, projectName string) {
		manifest1 := makeManifest(1, []int64{2, 3, 4, 5})
		size1 := sizeOfImage(manifest1)

		pushImage(projectName, "mysql", "latest", manifest1)
		suite.checkCountUsage(1, projectID)
		suite.checkStorageUsage(size1, projectID)

		pushImage(projectName, "mysql", "dev", manifest1)
		suite.checkCountUsage(2, projectID)
		suite.checkStorageUsage(size1, projectID)

		manifest2 := manifestWithAdditionalLayers(manifest1, []int64{6, 7})
		pushImage(projectName, "mariadb", "latest", manifest2)
		suite.checkCountUsage(3, projectID)

		totalSize := size1 + sizeOfManifest(manifest2) + 6 + 7
		suite.checkStorageUsage(totalSize, projectID)

		deleteManifest(projectName, "mysql", digestOfManifest(manifest1))
		suite.checkCountUsage(1, projectID)
		suite.checkStorageUsage(totalSize-sizeOfManifest(manifest1), projectID)
	})
}

func (suite *HandlerSuite) TestDeleteManifestInSameRepository() {
	withProject(func(projectID int64, projectName string) {
		manifest := makeManifest(1, []int64{2, 3, 4, 5})
		size := sizeOfImage(manifest)

		pushImage(projectName, "photon", "latest", manifest)
		suite.checkCountUsage(1, projectID)
		suite.checkStorageUsage(size, projectID)

		pushImage(projectName, "photon", "dev", manifest)
		suite.checkCountUsage(2, projectID)
		suite.checkStorageUsage(size, projectID)

		deleteManifest(projectName, "photon", digestOfManifest(manifest))
		suite.checkCountUsage(0, projectID)
		suite.checkStorageUsage(0, projectID)
	})
}

func (suite *HandlerSuite) TestDeleteManifestInDifferentRepositories() {
	withProject(func(projectID int64, projectName string) {
		manifest := makeManifest(1, []int64{2, 3, 4, 5})
		size := sizeOfImage(manifest)

		pushImage(projectName, "mysql", "latest", manifest)
		suite.checkCountUsage(1, projectID)
		suite.checkStorageUsage(size, projectID)

		pushImage(projectName, "mysql", "5.6", manifest)
		suite.checkCountUsage(2, projectID)
		suite.checkStorageUsage(size, projectID)

		pushImage(projectName, "redis", "latest", manifest)
		suite.checkCountUsage(3, projectID)
		suite.checkStorageUsage(size+sizeOfManifest(manifest), projectID)

		deleteManifest(projectName, "redis", digestOfManifest(manifest))
		suite.checkCountUsage(2, projectID)
		suite.checkStorageUsage(size, projectID)

		pushImage(projectName, "redis", "latest", manifest)
		suite.checkCountUsage(3, projectID)
		suite.checkStorageUsage(size+sizeOfManifest(manifest), projectID)
	})
}

func (suite *HandlerSuite) TestDeleteManifestInDifferentProjects() {
	withProject(func(projectID int64, projectName string) {
		manifest := makeManifest(1, []int64{2, 3, 4, 5})
		size := sizeOfImage(manifest)

		pushImage(projectName, "mysql", "latest", manifest)
		suite.checkStorageUsage(size, projectID)

		withProject(func(id int64, name string) {
			pushImage(name, "mysql", "latest", manifest)
			suite.checkStorageUsage(size, id)

			suite.checkStorageUsage(size, projectID)
			deleteManifest(projectName, "mysql", digestOfManifest(manifest))
			suite.checkCountUsage(0, projectID)
			suite.checkStorageUsage(0, projectID)
		})

	})
}

func (suite *HandlerSuite) TestPushDeletePush() {
	withProject(func(projectID int64, projectName string) {
		manifest := makeManifest(1, []int64{2, 3, 4, 5})
		size := sizeOfImage(manifest)

		pushImage(projectName, "photon", "latest", manifest)
		suite.checkStorageUsage(size, projectID)

		deleteManifest(projectName, "photon", digestOfManifest(manifest))
		suite.checkStorageUsage(0, projectID)

		pushImage(projectName, "photon", "latest", manifest)
		suite.checkStorageUsage(size, projectID)
	})
}

func (suite *HandlerSuite) TestPushImageRace() {
	withProject(func(projectID int64, projectName string) {
		manifest := makeManifest(1, []int64{2, 3, 4, 5})
		size := sizeOfImage(manifest)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				pushImage(projectName, "photon", "latest", manifest)
			}()
		}
		wg.Wait()

		suite.checkCountUsage(1, projectID)
		suite.checkStorageUsage(size, projectID)
	})
}

func (suite *HandlerSuite) TestDeleteImageRace() {
	withProject(func(projectID int64, projectName string) {
		manifest := makeManifest(1, []int64{2, 3, 4, 5})
		pushImage(projectName, "photon", "latest", manifest)

		count := 100
		size := sizeOfImage(manifest)
		for i := 0; i < count; i++ {
			manifest := makeManifest(1, []int64{2, 3, 4, 5})
			pushImage(projectName, "mysql", fmt.Sprintf("tag%d", i), manifest)
			size += sizeOfImage(manifest)
		}

		suite.checkCountUsage(int64(count+1), projectID)
		suite.checkStorageUsage(size, projectID)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				deleteManifest(projectName, "photon", digestOfManifest(manifest), func() bool {
					return i == 0
				})
			}(i)
		}
		wg.Wait()

		suite.checkCountUsage(int64(count), projectID)
		suite.checkStorageUsage(size-sizeOfImage(manifest), projectID)
	})
}

func (suite *HandlerSuite) TestDisableProjectQuota() {
	withProject(func(projectID int64, projectName string) {
		manifest := makeManifest(1, []int64{2, 3, 4, 5})
		pushImage(projectName, "photon", "latest", manifest)

		quotas, err := dao.ListQuotas(&models.QuotaQuery{
			Reference:   "project",
			ReferenceID: strconv.FormatInt(projectID, 10),
		})

		suite.Nil(err)
		suite.Len(quotas, 1)
	})

	withProject(func(projectID int64, projectName string) {
		cfg := config.GetCfgManager()
		cfg.Set(common.QuotaPerProjectEnable, false)
		defer cfg.Set(common.QuotaPerProjectEnable, true)

		manifest := makeManifest(1, []int64{2, 3, 4, 5})
		pushImage(projectName, "photon", "latest", manifest)

		quotas, err := dao.ListQuotas(&models.QuotaQuery{
			Reference:   "project",
			ReferenceID: strconv.FormatInt(projectID, 10),
		})

		suite.Nil(err)
		suite.Len(quotas, 0)
	})
}

func TestMain(m *testing.M) {
	config.Init()
	dao.PrepareTestForPostgresSQL()

	if result := m.Run(); result != 0 {
		os.Exit(result)
	}
}

func TestRunHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}
