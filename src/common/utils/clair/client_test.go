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
package clair

import (
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/clair/test"
	"github.com/stretchr/testify/assert"
)

var (
	notificationID = "ec45ec87-bfc8-4129-a1c3-d2b82622175a"
	layerName      = "03adedf41d4e0ea1b2458546a5b4717bf5f24b23489b25589e20c692aaf84d19"
	client         *Client
)

func TestMain(m *testing.M) {
	mockClairServer := test.NewMockServer()
	defer mockClairServer.Close()
	client = NewClient(mockClairServer.URL, nil)
	rc := m.Run()
	if rc != 0 {
		os.Exit(rc)
	}
}

func TestListNamespaces(t *testing.T) {
	assert := assert.New(t)
	ns, err := client.ListNamespaces()
	assert.Nil(err)
	assert.Equal(25, len(ns))
}

func TestNotifications(t *testing.T) {
	assert := assert.New(t)
	n, err := client.GetNotification(notificationID)
	assert.Nil(err)
	assert.Equal(notificationID, n.Name)
	_, err = client.GetNotification("noexist")
	assert.NotNil(err)
	err = client.DeleteNotification(notificationID)
	assert.Nil(err)
}

func TestLaysers(t *testing.T) {
	assert := assert.New(t)
	layer := models.ClairLayer{
		Name:       "fakelayer",
		ParentName: "parent",
		Path:       "http://registry:5000/layers/xxx",
	}
	err := client.ScanLayer(layer)
	assert.Nil(err)
	data, err := client.GetResult(layerName)
	assert.Nil(err)
	assert.Equal(layerName, data.Layer.Name)
	_, err = client.GetResult("notexist")
	assert.NotNil(err)
}
