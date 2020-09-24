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

package security

import (
	"fmt"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	_ "github.com/goharbor/harbor/src/core/auth/db"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"net/http"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	c := map[string]interface{}{
		common.AUTHMode: common.DBAuth,
	}
	config.InitWithSettings(c)

	user := models.User{
		Username: "tester",
		Password: "Harbor12345",
	}
	uid, err := dao.Register(user)
	defer func(id int64) {
		sql := fmt.Sprintf("DELETE FROM harbor_user WHERE user_id=%d", id)
		dao.ExecuteBatchSQL([]string{sql})
	}(uid)
	basicAuth := &basicAuth{}
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	req.SetBasicAuth("tester", "Harbor12345")
	ctx := basicAuth.Generate(req)
	assert.NotNil(t, ctx)
}
