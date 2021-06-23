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

package huawei

import (
	"fmt"
	"os"
	"testing"

	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

var HWAdapter adapter

func init() {
	hwRegistry := &model.Registry{
		ID:          1,
		Name:        "Huawei",
		Description: "Adapter for SWR -- The image registry of Huawei Cloud",
		Type:        model.RegistryTypeHuawei,
		URL:         "https://swr.cn-north-1.myhuaweicloud.com",
		Credential:  &model.Credential{AccessKey: "cn-north-1@IJYZLFBKBFN8LOUITAH", AccessSecret: "f31e8e2b948265afdae32e83722a7705fd43e154585ff69e64108247750e5d"},
		Insecure:    false,
		Status:      "",
	}
	adp, err := newAdapter(hwRegistry)
	if err != nil {
		os.Exit(1)
	}
	HWAdapter = *adp.(*adapter)

	gock.InterceptClient(HWAdapter.client.GetClient())
	gock.InterceptClient(HWAdapter.oriClient)
}

func mockRequest() *gock.Request {
	return gock.New("https://swr.cn-north-1.myhuaweicloud.com")
}

func mockGetJwtToken(repository string) {
	mockRequest().Get("/swr/auth/v2/registry/auth").
		MatchParam("scope", fmt.Sprintf("repository:%s:push,pull", repository)).
		Reply(200).
		JSON(jwtToken{
			Token: "token",
		})
}

func TestAdapter_FetchArtifacts(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	mockRequest().Get("/dockyard/v2/repositories").MatchParam("filter", "center::self").
		BasicAuth("cn-north-1@IJYZLFBKBFN8LOUITAH", "f31e8e2b948265afdae32e83722a7705fd43e154585ff69e64108247750e5d").
		Reply(200).
		JSON([]hwRepoQueryResult{
			{Name: "name1"},
			{Name: "name2"},
		})

	resources, err := HWAdapter.FetchArtifacts(nil)
	assert.NoError(t, err)
	assert.Len(t, resources, 2)
}

func TestAdapter_ManifestExist(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	mockGetJwtToken("sundaymango_mango/hello-world")
	mockRequest().Get("/v2/sundaymango_mango/hello-world/manifests/latest").
		Reply(200).
		JSON(hwManifest{
			MediaType: distribution.ManifestMediaTypes()[0],
		})

	exist, _, err := HWAdapter.ManifestExist("sundaymango_mango/hello-world", "latest")
	assert.NoError(t, err)
	assert.True(t, exist)
}

func TestAdapter_DeleteManifest(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	mockGetJwtToken("sundaymango_mango/hello-world")
	mockRequest().Delete("/v2/sundaymango_mango/hello-world/manifests/latest").Reply(200)

	err := HWAdapter.DeleteManifest("sundaymango_mango/hello-world", "latest")
	assert.NoError(t, err)
}
