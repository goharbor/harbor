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
	"os"
	"testing"

	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

var hwAdapter adp.Adapter

func init() {
	var err error
	hwRegistry := &model.Registry{
		ID:          1,
		Name:        "Huawei",
		Description: "Adapter for SWR -- The image registry of Huawei Cloud",
		Type:        model.RegistryTypeHuawei,
		URL:         "https://swr.cn-north-1.myhuaweicloud.com",
		Credential:  &model.Credential{AccessKey: "cn-north-1@AQR6NF5G2MQ1V7U4FCD", AccessSecret: "2f7ec95070592fd4838a3aa4fd09338c047fd1cd654b3422197318f97281cd9"},
		Insecure:    false,
		Status:      "",
	}

	hwAdapter, err = newAdapter(hwRegistry)
	if err != nil {
		os.Exit(1)
	}

	a := hwAdapter.(*adapter)
	gock.InterceptClient(a.client.GetClient())
	gock.InterceptClient(a.oriClient)
}

func TestAdapter_Info(t *testing.T) {
	info, err := hwAdapter.Info()
	if err != nil {
		t.Error(err)
	}
	t.Log(info)
}

func TestAdapter_PrepareForPush(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	mockRequest().Get("/dockyard/v2/namespaces/domain_repo_new").
		Reply(200).BodyString("{}")

	mockRequest().Post("/dockyard/v2/namespaces").BodyString(`{"namespace":"domain_repo_new"}`).
		Reply(200)

	repository := &model.Repository{
		Name:     "domain_repo_new",
		Metadata: make(map[string]interface{}),
	}
	resource := &model.Resource{}
	metadata := &model.ResourceMetadata{
		Repository: repository,
	}
	resource.Metadata = metadata
	err := hwAdapter.PrepareForPush([]*model.Resource{resource})
	assert.NoError(t, err)
}

func TestAdapter_HealthCheck(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	health, err := hwAdapter.HealthCheck()
	if err != nil {
		t.Error(err)
	}
	t.Log(health)
}
