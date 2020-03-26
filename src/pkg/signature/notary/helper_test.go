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
package notary

import (
	"encoding/json"
	"fmt"
	model2 "github.com/goharbor/harbor/src/pkg/signature/notary/model"
	test2 "github.com/goharbor/harbor/src/pkg/signature/notary/test"

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"

	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/goharbor/harbor/src/common"
)

var endpoint = "jt-dev.local.goharbor.io"
var notaryServer *httptest.Server

func TestMain(m *testing.M) {
	notaryServer = test2.NewNotaryServer(endpoint)
	defer notaryServer.Close()
	var defaultConfig = map[string]interface{}{
		common.ExtEndpoint:     "https://" + endpoint,
		common.WithNotary:      true,
		common.TokenExpiration: 30,
	}

	config.Init()
	test.InitDatabaseFromEnv()
	config.Upload(defaultConfig)
	notaryCachePath = "/tmp/notary"
	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}

func TestGetInternalTargets(t *testing.T) {
	targets, err := GetInternalTargets(notaryServer.URL, "admin", "library/busybox")
	assert.Nil(t, err, fmt.Sprintf("Unexpected error: %v", err))
	assert.Equal(t, 1, len(targets), "")
	assert.Equal(t, "latest-signed", targets[0].Tag, "")
}

func TestGetTargets(t *testing.T) {
	targets, err := GetTargets(notaryServer.URL, "admin", path.Join(endpoint, "library/busybox"))
	assert.Nil(t, err, fmt.Sprintf("Unexpected error: %v", err))
	assert.Equal(t, 1, len(targets), "")
	assert.Equal(t, "latest-signed", targets[0].Tag, "")

	targets, err = GetTargets(notaryServer.URL, "admin", path.Join(endpoint, "library/notexist"))
	assert.Nil(t, err, fmt.Sprintf("Unexpected error: %v", err))
	assert.Equal(t, 0, len(targets), "Targets list should be empty for non exist repo.")
}

func TestGetDigestFromTarget(t *testing.T) {
	str := ` {
		      "tag": "1.0",
			  "hashes": {
			        "sha256": "E1lggRW5RZnlZBY4usWu8d36p5u5YFfr9B68jTOs+Kc="
				}
		}`

	var t1 model2.Target
	err := json.Unmarshal([]byte(str), &t1)
	if err != nil {
		panic(err)
	}
	hash2 := make(map[string][]byte)
	t2 := model2.Target{
		Tag:    "2.0",
		Hashes: hash2,
	}
	d1, err1 := DigestFromTarget(t1)
	assert.Nil(t, err1, "Unexpected error: %v", err1)
	assert.Equal(t, "sha256:1359608115b94599e5641638bac5aef1ddfaa79bb96057ebf41ebc8d33acf8a7", d1, "digest mismatch")
	_, err2 := DigestFromTarget(t2)
	assert.NotNil(t, err2, "")
}
