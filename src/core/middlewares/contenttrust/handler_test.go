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

package contenttrust

import (
	"github.com/goharbor/harbor/src/common"
	notarytest "github.com/goharbor/harbor/src/common/utils/notary/test"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"os"
	"testing"
)

var endpoint = "jt-dev.local.goharbor.io"
var notaryServer *httptest.Server

var admiralEndpoint = "http://127.0.0.1:8282"
var token = ""

func TestMain(m *testing.M) {
	notaryServer = notarytest.NewNotaryServer(endpoint)
	defer notaryServer.Close()
	NotaryEndpoint = notaryServer.URL
	var defaultConfig = map[string]interface{}{
		common.ExtEndpoint:     "https://" + endpoint,
		common.WithNotary:      true,
		common.TokenExpiration: 30,
	}
	config.InitWithSettings(defaultConfig)
	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}

func TestMatchNotaryDigest(t *testing.T) {
	assert := assert.New(t)
	// The data from common/utils/notary/helper_test.go
	img1 := util.ImageInfo{Repository: "library/busybox", Reference: "latest-signed", ProjectName: "library", Digest: "sha256:dd97a3fe6d721c5cf03abac0f50e2848dc583f7c4e41bf39102ceb42edfd1808"}
	img2 := util.ImageInfo{Repository: "library/busybox", Reference: "2.0", ProjectName: "notary-demo", Digest: "sha256:12345678"}

	res1, err := matchNotaryDigest(img1)
	assert.Nil(err, "Unexpected error: %v, image: %#v", err, img1)
	assert.True(res1)

	res2, err := matchNotaryDigest(img2)
	assert.Nil(err, "Unexpected error: %v, image: %#v, take 2", err, img2)
	assert.False(res2)
}
