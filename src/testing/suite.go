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

package testing

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"time"

	o "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/internal/orm"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/suite"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	once sync.Once
)

// Suite ...
type Suite struct {
	suite.Suite
	ClearTables []string
	ClearSQLs   []string
}

// SetupSuite ...
func (suite *Suite) SetupSuite() {
	once.Do(func() {
		config.Init()
		dao.PrepareTestForPostgresSQL()
	})
}

// TearDownSuite ...
func (suite *Suite) TearDownSuite() {
	for _, sql := range suite.ClearSQLs {
		suite.ExecSQL(sql)
	}

	for _, table := range suite.ClearTables {
		dao.ClearTable(table)
	}
}

// RandString ...
func (suite *Suite) RandString(n int, letters ...string) string {
	if len(letters) == 0 || len(letters[0]) == 0 {
		letters = []string{"abcdefghijklmnopqrstuvwxyz"}
	}

	letterBytes := []byte(letters[0])

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// Digest ...
func (suite *Suite) Digest() digest.Digest {
	return digest.FromString(suite.RandString(128))
}

// DigestString ...
func (suite *Suite) DigestString() string {
	return suite.Digest().String()
}

// WithProject ...
func (suite *Suite) WithProject(f func(int64, string), projectNames ...string) {
	var projectName string
	if len(projectNames) > 0 {
		projectName = projectNames[0]
	} else {
		projectName = suite.RandString(5)
	}

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

// Context ...
func (suite *Suite) Context() context.Context {
	return orm.NewContext(context.TODO(), o.NewOrm())
}

// NewRequest ...
func (suite *Suite) NewRequest(method, target string, body io.Reader, queries ...map[string]string) *http.Request {
	req := httptest.NewRequest(method, target, body)

	if len(queries) > 0 {
		q := req.URL.Query()
		for key, value := range queries[0] {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	return req.WithContext(suite.Context())
}

// NextHandler ...
func (suite *Suite) NextHandler(statusCode int, headers map[string]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		for key, value := range headers {
			w.Header().Set(key, value)
		}
	})
}

// ExecSQL ...
func (suite *Suite) ExecSQL(query string, args ...interface{}) {
	o := o.NewOrm()

	_, err := o.Raw(query, args...).Exec()
	suite.Nil(err)
}

// IsNotFoundErr ...
func (suite *Suite) IsNotFoundErr(err error) bool {
	return suite.True(ierror.IsNotFoundErr(err))
}

// AssertResourceUsage ...
func (suite *Suite) AssertResourceUsage(expected int64, resource types.ResourceName, projectID int64) {
	usage := models.QuotaUsage{Reference: "project", ReferenceID: strconv.FormatInt(projectID, 10)}
	err := dao.GetOrmer().Read(&usage, "reference", "reference_id")
	suite.Nil(err, fmt.Sprintf("Failed to get resource %s usage of project %d, error: %v", resource, projectID, err))

	used, err := types.NewResourceList(usage.Used)
	suite.Nil(err, "Bad resource usage of project %d", projectID)
	suite.Equal(expected, used[resource])
}
