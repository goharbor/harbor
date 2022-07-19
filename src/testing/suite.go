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
	"sync"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/config"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"

	o "github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
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

	ctx := suite.Context()

	projectID, err := suite.AddProject(ctx, &proModels.Project{
		Name:    projectName,
		OwnerID: 1,
	})
	if err != nil {
		panic(err)
	}

	defer func() {
		suite.DeleteProject(ctx, projectName, projectID)
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
	return suite.True(errors.IsNotFoundErr(err))
}

// AddProject put here is to avoid import cycle
func (suite *Suite) AddProject(ctx context.Context, project *proModels.Project) (int64, error) {
	// Create create a project instance
	var projectID int64

	h := func(ctx context.Context) error {
		o, err := orm.FromContext(ctx)
		if err != nil {
			return err
		}

		now := time.Now()
		project.CreationTime = now
		project.UpdateTime = now

		projectID, err = o.Insert(project)
		if err != nil {
			return orm.WrapConflictError(err, "The project named %s already exists", project.Name)
		}

		member := &proModels.Member{
			ProjectID:    projectID,
			EntityID:     project.OwnerID,
			Role:         common.RoleProjectAdmin,
			EntityType:   common.UserMember,
			CreationTime: now,
			UpdateTime:   now,
		}

		if _, err := o.Insert(member); err != nil {
			return err
		}

		return nil
	}

	if err := orm.WithTransaction(h)(ctx); err != nil {
		return 0, err
	}

	return projectID, nil
}

// DeleteProject put here is to avoid import cycle
func (suite *Suite) DeleteProject(ctx context.Context, projectName string, projectID int64) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s#%d", projectName, projectID)
	sql := `update project
		set deleted = true, name = ?
		where project_id = ?`
	_, err = o.Raw(sql, name, projectID).Exec()
	return err
}
